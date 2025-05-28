/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	infrahubv1alpha1 "github.com/infrahub-operator/vidra/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/infrahub-operator/vidra/internal/adapter/infrahub"
	"github.com/infrahub-operator/vidra/internal/adapter/k8s"
	"github.com/infrahub-operator/vidra/internal/domain"
)

type InfrahubSyncReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	RequeueAfter   time.Duration
	QueryName      string
	InfrahubClient domain.InfrahubClient
}

// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubsyncs/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrahub.operators.com,resources=infrahubresources/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get

// Reconcile reconciles the InfrahubSync resource.
// controller/infrahubsync_controller.go

func (r *InfrahubSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting Sync reconciliation", "request", req.NamespacedName)

	infrahubSync := &infrahubv1alpha1.InfrahubSync{}
	if err := r.Get(ctx, req.NamespacedName, infrahubSync); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("InfrahubSync resource not found, skipping")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get InfrahubSync resource")
		return ctrl.Result{RequeueAfter: r.RequeueAfter}, MarkStateFailed(ctx, r.Client, infrahubSync, err)
	}

	// Mark the InfrahubSync resource as running
	if err := MarkState(ctx, r.Client, infrahubSync, func() {
		infrahubSync.Status.SyncState = infrahubv1alpha1.StateRunning
	}); err != nil {
		logger.Error(err, "Failed to update SyncState to Running")
		return ctrl.Result{RequeueAfter: r.RequeueAfter}, err
	}

	var apiURL = infrahubSync.Spec.Source.InfrahubAPIURL

	// Get authentication credentials from Kubernetes Secret
	username, password, err := r.getCredentials(ctx, apiURL)
	if err != nil {
		logger.Error(err, "Failed to get credentials from Secret")
		return ctrl.Result{RequeueAfter: r.RequeueAfter}, MarkStateFailed(ctx, r.Client, infrahubSync, err)
	}

	// Get authentication token using the Infrahub client
	token, err := r.InfrahubClient.Login(apiURL, username, password)
	if err != nil {
		logger.Error(err, "Failed to login to Infrahub")
		return ctrl.Result{RequeueAfter: r.RequeueAfter}, MarkStateFailed(ctx, r.Client, infrahubSync, err)
	}

	// Run the query and process the results using the Infrahub client
	queryResult, err := r.InfrahubClient.RunQuery(
		r.QueryName,
		apiURL,
		infrahubSync.Spec.Source.ArtifactName,
		infrahubSync.Spec.Source.TargetBranch,
		infrahubSync.Spec.Source.TargetDate,
		token)
	if err != nil {
		logger.Error(err, "Failed to execute query")
		return ctrl.Result{}, MarkStateFailed(ctx, r.Client, infrahubSync, err)
	}
	logger.Info("Query executed successfully", "result", queryResult)

	// Process query results and compare with existing resources
	err = r.processArtifacts(ctx, infrahubSync, queryResult, token)
	if err != nil {
		logger.Error(err, "Error processing artifacts")
		return ctrl.Result{RequeueAfter: r.RequeueAfter}, MarkStateFailed(ctx, r.Client, infrahubSync, err)
	}

	// Update the status of the InfrahubSync resource
	if err := MarkState(ctx, r.Client, infrahubSync, func() {
		infrahubSync.Status.SyncState = infrahubv1alpha1.StateSucceeded
		infrahubSync.Status.LastSyncTime = metav1.Now()
	}); err != nil {
		logger.Error(err, "Failed to update SyncState to Success")
		return ctrl.Result{RequeueAfter: r.RequeueAfter}, err
	}

	return ctrl.Result{RequeueAfter: r.RequeueAfter}, nil
}

// getCredentials fetches Infrahub API credentials from Kubernetes Secret
func (r *InfrahubSyncReconciler) getCredentials(ctx context.Context, apiURL string) (string, string, error) {
	secretList := &v1.SecretList{}

	trimmedAPIURL := strings.TrimPrefix(strings.Split(apiURL, ":")[1], "//") // Remove https and port
	if err := k8s.GetSortedListByLabel(ctx, r.Client, "infrahub-api-url", trimmedAPIURL, secretList); err != nil {
		return "", "", fmt.Errorf("no secret found with InfrahubAPIURL: %s, error: %w", apiURL, err)
	}

	var usernameRaw, passwordRaw []byte
	var found bool

	for _, secret := range secretList.Items {
		if u, uok := secret.Data["username"]; uok {
			if p, pok := secret.Data["password"]; pok {
				usernameRaw = u
				passwordRaw = p
				found = true
				break
			}
		}
	}

	if !found {
		return "", "", fmt.Errorf("no secret found with both username and password fields")
	}

	username := string(bytes.TrimSpace(usernameRaw))
	password := string(bytes.TrimSpace(passwordRaw))

	return username, password, nil
}

// processArtifacts processes the artifacts retrieved from Infrahub and syncs resources
func (r *InfrahubSyncReconciler) processArtifacts(
	ctx context.Context,
	infrahubSync *infrahubv1alpha1.InfrahubSync,
	artifacts *[]domain.Artifact,
	token string,
) error {
	log := log.FromContext(ctx)

	log.Info("Processing artifacts", "artifactCount", len(*artifacts))

	// Build current artifact ID map
	currentArtifactIDs := make(map[string]struct{}, len(*artifacts))
	for _, artifact := range *artifacts {
		currentArtifactIDs[artifact.ID] = struct{}{}
	}

	// List all VidraResources in the same namespace
	var resourceList infrahubv1alpha1.VidraResourceList
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.List(ctx, &resourceList, client.InNamespace(infrahubSync.Namespace))
	}); err != nil {
		return fmt.Errorf("failed to list VidraResources: %w", err)
	}

	// Delete stale resources
	for _, res := range resourceList.Items {
		if res.Spec.Destination.Server == infrahubSync.Spec.Destination.Server &&
			res.Spec.Destination.Namespace == infrahubSync.Spec.Destination.Namespace {

			if _, exists := currentArtifactIDs[res.Name]; !exists {
				if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
					return r.Delete(ctx, &res)
				}); err != nil {
					return fmt.Errorf("failed to delete stale VidraResource %s: %w", res.Name, err)
				}
				log.Info("Deleted stale VidraResource", "name", res.Name)
			}
		}
	}

	// Create or update resources for current artifacts
	for _, artifact := range *artifacts {

		found := false
		for _, checksum := range infrahubSync.Status.Checksums {
			if artifact.Checksum == checksum {
				found = true
				break
			}
		}
		if !found {
			contentReader, err := r.InfrahubClient.DownloadArtifact(
				infrahubSync.Spec.Source.InfrahubAPIURL,
				artifact.ID,
				infrahubSync.Spec.Source.TargetBranch,
				infrahubSync.Spec.Source.TargetDate,
				token,
			)
			if err != nil {
				return fmt.Errorf("failed to download artifact: %w", err)
			}
			var sb strings.Builder
			if _, err := io.Copy(&sb, contentReader); err != nil {
				return fmt.Errorf("failed to read artifact content: %w", err)
			}

			resource := &infrahubv1alpha1.VidraResource{
				ObjectMeta: metav1.ObjectMeta{
					Name: artifact.ID,
					Finalizers: []string{
						"vidraresource.infrahub.operators.com/finalizer",
					},
				},
			}

			if err := ctrl.SetControllerReference(infrahubSync, resource, r.Scheme); err != nil {
				return fmt.Errorf("failed to set controller reference: %w", err)
			}

			var opResult controllerutil.OperationResult
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				var innerErr error
				opResult, innerErr = ctrl.CreateOrUpdate(ctx, r.Client, resource, func() error {
					resource.Spec.Destination = infrahubv1alpha1.InfrahubSyncDestination{
						Server:            infrahubSync.Spec.Destination.Server,
						Namespace:         infrahubSync.Spec.Destination.Namespace,
						ReconcileOnEvents: infrahubSync.Spec.Destination.ReconcileOnEvents,
					}
					resource.Spec.Manifest = strings.TrimSpace(sb.String())
					return nil
				})
				return innerErr
			})

			if err != nil {
				return fmt.Errorf("failed to create or update VidraResource %s: %w", resource.Name, err)
			}

			log.Info("Synced Infrahub to VidraResources", "name", resource.Name, "operation", opResult)
		}
	}

	return nil
}

func (r *InfrahubSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.InfrahubClient = infrahub.NewClient()
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

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrahubv1alpha1.InfrahubSync{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *InfrahubSyncReconciler) InitConfigWithClient(ctx context.Context, k8sClient client.Client, labelKey, labelValue string) error {
	const defaultRequeue = time.Minute
	const defaultQueryName = "ArtifactIDs"

	// Start with the default values
	r.RequeueAfter = defaultRequeue
	r.QueryName = defaultQueryName
	var configMaps v1.ConfigMapList
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

	var configMap *v1.ConfigMap
	for _, cm := range configMaps.Items {
		if okRequeue := cm.Data["requeueSyncAfter"] != ""; okRequeue || (cm.Data["queryName"] != "") {
			configMap = &cm
			break
		}
	}
	// Check for 'requeueAfter' and update if available
	requeueAfter, ok := configMap.Data["requeueSyncAfter"]
	if ok {
		duration, err := time.ParseDuration(requeueAfter)
		if err == nil {
			r.RequeueAfter = duration
		}
	}

	// Check for 'queryName' and update if available
	queryName, ok := configMap.Data["queryName"]
	if ok {
		r.QueryName = queryName
	}

	return nil
}
