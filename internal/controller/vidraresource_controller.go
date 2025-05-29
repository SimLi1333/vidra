package controller

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	infrahubv1alpha1 "github.com/infrahub-operator/vidra/api/v1alpha1"
	"github.com/infrahub-operator/vidra/internal/adapter/infrahub"
	"github.com/infrahub-operator/vidra/internal/adapter/k8s"
	"github.com/infrahub-operator/vidra/internal/domain"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	FinalizerName   = "vidraresource.infrahub.operators.com/finalizer"
	OwnerAnnotation = "vidraresource.infrahub.operators.com/owned-by"
	vidraOperator   = "vidra"
)

type VidraResourceReconciler struct {
	client.Client
	Scheme                     *runtime.Scheme
	RESTMapper                 meta.RESTMapper
	InfrahubClient             domain.InfrahubClient
	DynamicMulticlusterFactory domain.DynamicMulticlusterFactory
	DynamicWatcherFactory      domain.DynamicWatcherFactory
	DynamicWatcherClient       dynamic.Interface
	RequeueAfter               time.Duration
	EventBasedReconcile        bool
}

// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubresources/finalizers,verbs=update
// +kubebuilder:rbac:groups="*",resources="*",verbs=get;list;watch;create;update;patch;delete

func (r *VidraResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling", "resource", req.NamespacedName)

	res := &infrahubv1alpha1.VidraResource{}
	if err := r.Get(ctx, req.NamespacedName, res); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("VidraResource resource not found, skipping")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get VidraResource resource")
		return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, err)
	}
	var destClient client.Client
	if res.Spec.Destination.Server == "" || res.Spec.Destination.Server == "https://kubernetes.default.svc" {
		logger.Info("Using local client for destination")

		destClient = r.Client
	} else {
		logger.Info("Using cached client for destination", "server", res.Spec.Destination.Server)
		var err error
		destClient, err = r.DynamicMulticlusterFactory.GetCachedClientFor(ctx, res.Spec.Destination.Server, r.Client)
		if err != nil {
			return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, fmt.Errorf("failed to get client for destination: %w", err))
		}
	}

	if !res.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, res, destClient)
	}

	if err := MarkState(ctx, r.Client, res, func() {
		res.Status.DeployState = infrahubv1alpha1.StateRunning
	}); err != nil {
		return ctrl.Result{}, err
	}

	if !r.hasFinalizer(res) {
		logger.Info("Adding finalizer")
		if err := r.addFinalizer(ctx, res); err != nil {
			return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, err)
		}
	}

	if res.Spec.Manifest == "" {
		logger.Error(nil, "No manifests available in spec to reconcile")
		return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, fmt.Errorf("no manifests available in spec to reconcile"))
	}
	contentReader := strings.NewReader(res.Spec.Manifest)

	newResources, gvrList, err := r.decodeAndApplyResources(ctx, res, contentReader, destClient)
	if err != nil {
		logger.Error(err, "Failed to decode and apply resources")
		return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, err)
	}

	if err := r.cleanupRemovedResources(ctx, res, newResources, destClient); err != nil {
		logger.Error(err, "Failed to clean up removed resources")
		if res.Status.DeployState == infrahubv1alpha1.StateStale {
			logger.Error(err, "State is stale, returning error")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, err)
	}

	res.Status.ManagedResources = buildFinalResourceList(res.Status.ManagedResources, newResources)
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Status().Update(ctx, res)
	}); err != nil {
		return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, err)
	}

	if r.EventBasedReconcile || res.Spec.Destination.ReconcileOnEvents {
		r.DynamicWatcherFactory.StartWatchingGVRs(
			r.DynamicWatcherClient,
			gvrList,
			func(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
				r.handleLabeledResource(obj, gvr)
				r.triggerReconcileForOwner(obj)
			},
		)
		r.RequeueAfter = 0 // Disable default requeue for event-based reconciliation
	}

	if err := MarkState(ctx, r.Client, res, func() {
		res.Status.LastSyncTime = metav1.Now()
		res.Status.DeployState = infrahubv1alpha1.StateSucceeded
	}); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Reconciliation complete")
	return ctrl.Result{RequeueAfter: r.RequeueAfter}, nil
}

func (r *VidraResourceReconciler) handleDeletion(ctx context.Context, res *infrahubv1alpha1.VidraResource, destClient client.Client) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if !r.hasFinalizer(res) {
		return ctrl.Result{}, nil
	}
	logger.Info("Cleaning up managed resources")

	for _, mr := range res.Status.ManagedResources {
		if err := r.deleteManagedResource(ctx, res, mr, destClient); err != nil {
			return ctrl.Result{}, MarkStateFailed(ctx, r.Client, res, err)
		}
	}
	return ctrl.Result{}, r.removeFinalizer(ctx, res)
}

func (r *VidraResourceReconciler) decodeAndApplyResources(
	ctx context.Context,
	res *infrahubv1alpha1.VidraResource,
	contentReader io.Reader,
	destClient client.Client,
) (map[string]infrahubv1alpha1.ManagedResourceStatus, []schema.GroupVersionResource, error) {
	logger := log.FromContext(ctx).WithValues("resource", res.Name)

	reader := bufio.NewReaderSize(contentReader, 4096)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)

	resources := map[string]infrahubv1alpha1.ManagedResourceStatus{}
	gvrList := []schema.GroupVersionResource{}
	seenGVR := map[schema.GroupVersionResource]struct{}{}

	for {
		u := &unstructured.Unstructured{}
		if err := decoder.Decode(u); err != nil {
			if err == io.EOF {
				break
			}
			logger.Error(err, "Failed to decode")
			return nil, nil, fmt.Errorf("decode artifact: %w", err)
		}

		gvk := u.GroupVersionKind()
		mapping, err := r.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, nil, fmt.Errorf("REST mapping: %w", err)
		}

		// Collect GVRs for dynamic watcher
		if (r.EventBasedReconcile || res.Spec.Destination.ReconcileOnEvents) && destClient == r.Client {
			gvr := mapping.Resource
			if _, exists := seenGVR[gvr]; !exists {
				gvrList = append(gvrList, gvr)
				seenGVR[gvr] = struct{}{}
			}
		}

		if mapping.Scope.Name() == meta.RESTScopeNameNamespace && u.GetNamespace() == "" {
			u.SetNamespace(res.Spec.Destination.Namespace)
		}
		if destClient == r.Client {
			if err := ctrl.SetControllerReference(res, u, r.Scheme); err != nil {
				return nil, nil, fmt.Errorf("set controller reference: %w", err)
			}
		}

		annotateWithOwner(u, res.Name)
		if err := r.applyResource(ctx, res, u, destClient); err != nil {
			logger.Error(err, "apply resource failed", "GVK", gvk, "Name", u.GetName())
			return nil, nil, err
		}

		status := infrahubv1alpha1.ManagedResourceStatus{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
			Name:       u.GetName(),
			Namespace:  u.GetNamespace(),
		}
		resources[resourceKey(status)] = status
	}

	return resources, gvrList, nil
}

func (r *VidraResourceReconciler) cleanupRemovedResources(
	ctx context.Context,
	res *infrahubv1alpha1.VidraResource,
	current map[string]infrahubv1alpha1.ManagedResourceStatus,
	destClient client.Client,
) error {
	logger := log.FromContext(ctx)
	var remaining []infrahubv1alpha1.ManagedResourceStatus

	for _, old := range res.Status.ManagedResources {
		key := resourceKey(old)
		if _, stillExists := current[key]; stillExists {
			// Still managed — retain it
			remaining = append(remaining, old)
			continue
		}

		// Resource is stale — prepare for deletion
		if err := r.deleteManagedResource(ctx, res, old, destClient); err != nil {
			logger.Error(err, "Failed to delete managed resource", "resource", old)
			return err
		}
	}

	res.Status.ManagedResources = remaining
	return nil
}

func (r *VidraResourceReconciler) deleteManagedResource(ctx context.Context, res *infrahubv1alpha1.VidraResource, old infrahubv1alpha1.ManagedResourceStatus, destClient client.Client) error {
	logger := log.FromContext(ctx).WithValues("kind", old.Kind, "name", old.Name, "namespace", old.Namespace)
	logger.Info("Deleting stale managed resource")

	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(old.APIVersion)
	obj.SetKind(old.Kind)
	obj.SetName(old.Name)
	obj.SetNamespace(old.Namespace)

	if err := destClient.Get(ctx, types.NamespacedName{Name: old.Name, Namespace: old.Namespace}, obj); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource not found, skipping deletion")
			return nil
		}
		logger.Error(err, "Failed to fetch resource from cluster")
		return fmt.Errorf("fetch resource %s: %w", old.Name, err)
	}

	if obj.GetAnnotations()["managed-by"] != vidraOperator {
		logger.Info("Skipping deletion, resource not managed by this Infrahub Operator",
			"expectedOwner", res.Name,
			"actualAnnotations", obj.GetAnnotations())
		return nil
	}

	if obj.GetAnnotations()[OwnerAnnotation] != res.Name {
		owners := obj.GetAnnotations()[OwnerAnnotation]
		ownerList := removeString(strings.Split(owners, ","), res.Name)
		objAnnotations := obj.GetAnnotations()
		objAnnotations[OwnerAnnotation] = strings.Join(ownerList, ",")
		obj.SetAnnotations(objAnnotations)
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			return destClient.Update(ctx, obj)
		}); err != nil {
			return fmt.Errorf("failed to update resource annotations: %w", err)
		}
		return nil
	}
	logger.Info("Deleting resource", "resource", obj)
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return client.IgnoreNotFound(destClient.Delete(ctx, obj))
	}); err != nil {
		logger.Error(err, "Failed to delete stale resource")
		if err := MarkState(ctx, r.Client, res, func() {
			res.Status.DeployState = infrahubv1alpha1.StateStale
		}); err != nil {
			logger.Error(err, "Failed to update SyncState to Stale")
			return err
		}
		return err
	}

	return nil
}

func (r *VidraResourceReconciler) applyResource(ctx context.Context, res *infrahubv1alpha1.VidraResource, desired *unstructured.Unstructured, destClient client.Client) error {
	logger := log.FromContext(ctx)

	// Prepare the existing resource object
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(desired.GroupVersionKind())
	existing.SetNamespace(desired.GetNamespace())
	existing.SetName(desired.GetName())
	// Add label "managed-by": "vida" to the resource
	labels := desired.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels["managed-by"] = "vida"
	desired.SetLabels(labels)

	// Try fetching the existing resource
	err := destClient.Get(ctx, client.ObjectKeyFromObject(existing), existing)
	if err != nil {
		if errors.IsNotFound(err) {
			// Resource doesn't exist, create it
			return destClient.Create(ctx, desired)
		}
		return err
	}

	// Log resource existence and check if it's managed by the operator
	logger.Info("Resource already exists", "name", existing.GetName(), "namespace", existing.GetNamespace())

	if existing.GetAnnotations()["managed-by"] != vidraOperator && res.Status.LastSyncTime.IsZero() {
		fmt.Printf("Resource %s/%s already exists but is not managed by this operator\n", existing.GetNamespace(), existing.GetName())
		return fmt.Errorf("resource %s/%s already exists but is not managed by this operator", existing.GetNamespace(), existing.GetName())
	}

	// Check if annotations match, if so, update resource
	if r.shouldUpdateResource(existing, desired) {
		logger.Info("Resource already exists and is managed by this vidraResource -> updating", "name", existing.GetName(), "namespace", existing.GetNamespace())
		if r.isEqual(existing, desired) {
			return nil
		}
		desired.SetResourceVersion(existing.GetResourceVersion())
		return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			return destClient.Update(ctx, desired)
		})
	}

	// Normalize spec maps before comparing
	if r.isEqual(existing, desired) {
		// If specs are equal, patch the owner annotation
		logger.Info("Resource is managed by another vidraResource, patching owner annotation", "name", existing.GetName(), "namespace", existing.GetNamespace())
		if err := MarkState(ctx, r.Client, res, func() {
			res.Status.LastError = fmt.Sprintf("Warning: resource is already managed by vidraResource: %s", existing.GetAnnotations()[OwnerAnnotation])
		}); err != nil {
			logger.Error(err, "Failed to update LastError with warning")
			return err
		}
		return r.patchOwnerAnnotation(ctx, desired, existing, destClient)
	}
	logger.Info("updating changed resource", "name", existing.GetName(), "namespace", existing.GetNamespace())
	desired.SetResourceVersion(existing.GetResourceVersion())
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return destClient.Update(ctx, desired)
	})
}

func (r *VidraResourceReconciler) shouldUpdateResource(existing, desired *unstructured.Unstructured) bool {
	existingAnnotations := existing.GetAnnotations()
	desiredAnnotations := desired.GetAnnotations()

	return existingAnnotations[OwnerAnnotation] == desiredAnnotations[OwnerAnnotation]
}

func (r *VidraResourceReconciler) isEqual(existing, desired *unstructured.Unstructured) bool {
	// Prepare deep copies for safe comparison
	existingCopy := existing.DeepCopy()
	desiredCopy := desired.DeepCopy()

	// Remove finalizers, status, and metadata from both objects
	r.removeFinalizers(existingCopy)
	r.removeFinalizers(desiredCopy)
	delete(existingCopy.Object, "status")
	delete(desiredCopy.Object, "status")
	delete(existingCopy.Object, "metadata")
	delete(desiredCopy.Object, "metadata")

	return equality.Semantic.DeepEqual(desiredCopy.Object, existingCopy.Object)
}

func (r *VidraResourceReconciler) removeFinalizers(resource *unstructured.Unstructured) {
	if spec, ok := resource.Object["spec"].(map[string]interface{}); ok {
		delete(spec, "finalizers")
	} else {
		resource.Object["spec"] = map[string]interface{}{}
	}
}

func (r *VidraResourceReconciler) patchOwnerAnnotation(ctx context.Context, desired, existing *unstructured.Unstructured, destClient client.Client) error {
	// Prepare a patch to update the owner annotation
	patch := existing.DeepCopy()
	patchAnnotations := patch.GetAnnotations()
	if patchAnnotations == nil {
		patchAnnotations = map[string]string{}
	}

	// Combine owner annotations
	existingAnnotations := existing.GetAnnotations()
	desiredAnnotations := desired.GetAnnotations()
	patchAnnotations[OwnerAnnotation] = fmt.Sprintf("%s,%s", existingAnnotations[OwnerAnnotation], desiredAnnotations[OwnerAnnotation])
	existing.SetAnnotations(patchAnnotations)

	// Retry on conflict when patching annotations
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return destClient.Patch(ctx, existing, client.MergeFrom(patch))
	})
}

// Finalizer helpers
func (r *VidraResourceReconciler) hasFinalizer(obj *infrahubv1alpha1.VidraResource) bool {
	return containsString(obj.Finalizers, FinalizerName)
}

func (r *VidraResourceReconciler) addFinalizer(ctx context.Context, obj *infrahubv1alpha1.VidraResource) error {
	if containsString(obj.Finalizers, FinalizerName) {
		return nil
	}
	patch := client.MergeFrom(obj.DeepCopy())
	obj.Finalizers = append(obj.Finalizers, FinalizerName)
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Patch(ctx, obj, patch)
	})
}

func (r *VidraResourceReconciler) removeFinalizer(ctx context.Context, obj *infrahubv1alpha1.VidraResource) error {
	patch := client.MergeFrom(obj.DeepCopy())
	obj.Finalizers = removeString(obj.Finalizers, FinalizerName)
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Patch(ctx, obj, patch)
	})
}

// Setup
func (r *VidraResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.InfrahubClient = infrahub.NewClient()
	r.DynamicMulticlusterFactory = k8s.NewDynamicMulticlusterFactory()

	// Create a direct (non-cached) client
	cfg := mgr.GetConfig()
	scheme := mgr.GetScheme()

	nonCachedClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create non-cached client: %w", err)
	}

	// Use the non-cached client to fetch the config map via label selector
	labelKey := "app"
	labelValue := "vidra"

	if err := r.InitConfigWithClient(context.Background(), nonCachedClient, labelKey, labelValue); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}
	// Set up the dynamic watcher factory and dynamic client
	r.DynamicWatcherFactory = k8s.NewDynamicWatcherFactory()
	r.DynamicWatcherClient, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrahubv1alpha1.VidraResource{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *VidraResourceReconciler) InitConfigWithClient(ctx context.Context, k8sClient client.Client, labelKey, labelValue string) error {
	const defaultRequeue = 10 * time.Minute

	// Start with the default values
	r.RequeueAfter = defaultRequeue
	var configMaps corev1.ConfigMapList
	if err := k8s.GetSortedListByLabel(ctx, k8sClient, labelKey, labelValue, &configMaps); err != nil {
		if strings.Contains(err.Error(), "no resources found with label") {
			return nil // Use default values and continue
		}
		return err
	}

	// If no ConfigMap is found, return with defaults
	if len(configMaps.Items) == 0 {
		return nil
	}

	var configMap *corev1.ConfigMap
	for _, cm := range configMaps.Items {
		if cm.Data["requeueRecourcesAfter"] != "" {
			configMap = &cm
			break
		}
	}
	// Check for 'requeueAfter' and update if available
	requeueAfter, ok := configMap.Data["requeueRecourcesAfter"]
	if ok {
		duration, err := time.ParseDuration(requeueAfter)
		if err == nil {
			r.RequeueAfter = duration
		}
	}

	eventBasedReconcileStr, ok := configMap.Data["eventBasedReconcile"]
	if ok {
		val := strings.ToLower(strings.TrimSpace(eventBasedReconcileStr))
		if val == "true" {
			r.EventBasedReconcile = true
		} else {
			r.EventBasedReconcile = false
		}
	} else {
		r.EventBasedReconcile = false
	}

	return nil
}

// Utilities
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
func removeString(slice []string, s string) []string {
	var out []string
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
func annotateWithOwner(obj *unstructured.Unstructured, owner string) {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = map[string]string{}
	}
	ann[OwnerAnnotation] = owner
	ann["managed-by"] = vidraOperator
	obj.SetAnnotations(ann)
}
func resourceKey(res infrahubv1alpha1.ManagedResourceStatus) string {
	return fmt.Sprintf("%s:%s:%s:%s", res.APIVersion, res.Kind, res.Namespace, res.Name)
}

func buildFinalResourceList(existing []infrahubv1alpha1.ManagedResourceStatus, new map[string]infrahubv1alpha1.ManagedResourceStatus) []infrahubv1alpha1.ManagedResourceStatus {
	result := existing[:0]
	seen := map[string]bool{}
	for _, r := range existing {
		key := resourceKey(r)
		if _, ok := new[key]; ok {
			result = append(result, r)
			seen[key] = true
		}
	}
	for key, r := range new {
		if !seen[key] {
			result = append(result, r)
		}
	}
	return result
}
