package controller

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	infrahubv1alpha1 "github.com/simli1333/vidra/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *InfrahubResourceReconciler) StartDynamicWatchers(cfg *rest.Config) {
	discoClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		log.Printf("[WATCH] Discovery client error: %v", err)
		return
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Printf("[WATCH] Dynamic client error: %v", err)
		return
	}

	gvrs, err := discoverWatchableResources(discoClient)
	if err != nil {
		log.Printf("[WATCH] Failed to discover resources: %v", err)
		return
	}

	seen := map[schema.GroupVersionResource]struct{}{}
	for _, gvr := range gvrs {
		if _, ok := seen[gvr]; ok {
			continue // Avoid duplicate informer for the same GVR
		}
		seen[gvr] = struct{}{}
		go r.watchGVR(dynamicClient, gvr)
	}
}

// discoverWatchableResources returns all watchable namespaced GVRs (excluding subresources)
func discoverWatchableResources(d discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error) {
	var gvrs []schema.GroupVersionResource

	apiLists, err := d.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	for _, apiList := range apiLists {
		gv, err := schema.ParseGroupVersion(apiList.GroupVersion)
		if err != nil {
			log.Printf("Skipping malformed GroupVersion: %s", apiList.GroupVersion)
			continue
		}

		for _, res := range apiList.APIResources {
			if strings.Contains(res.Name, "/") {
				continue // skip subresources
			}
			if !res.Namespaced {
				continue
			}
			if !containsVerb(res.Verbs, "watch") {
				continue
			}
			gvr := gv.WithResource(res.Name)
			gvrs = append(gvrs, gvr)
		}
	}

	return gvrs, nil
}

func containsVerb(verbs []string, target string) bool {
	for _, v := range verbs {
		if v == target {
			return true
		}
	}
	return false
}

func (r *InfrahubResourceReconciler) watchGVR(client dynamic.Interface, gvr schema.GroupVersionResource) {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc: func(opts v1.ListOptions) (runtime.Object, error) {
				opts.LabelSelector = labels.SelectorFromSet(labels.Set{"managed-by": "vida"}).String()
				return client.Resource(gvr).Namespace("").List(context.TODO(), opts)
			},
			WatchFunc: func(opts v1.ListOptions) (watch.Interface, error) {
				opts.LabelSelector = labels.SelectorFromSet(labels.Set{"managed-by": "vida"}).String()
				return client.Resource(gvr).Namespace("").Watch(context.TODO(), opts)
			},
		},
		&unstructured.Unstructured{},
		0, // no resync
	)

	genChanged := predicate.GenerationChangedPredicate{}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			r.handleLabeledResource(obj.(*unstructured.Unstructured), gvr)
			// r.triggerReconcileForOwner(obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldUnstr := oldObj.(*unstructured.Unstructured)
			newUnstr := newObj.(*unstructured.Unstructured)

			// Only trigger if GenerationChangedPredicate returns true
			if genChanged.Update(event.UpdateEvent{
				ObjectOld: oldUnstr,
				ObjectNew: newUnstr,
			}) {
				r.handleLabeledResource(newUnstr, gvr)
				r.triggerReconcileForOwner(newUnstr)
			}
		},
		DeleteFunc: func(obj interface{}) {
			r.handleLabeledResource(obj.(*unstructured.Unstructured), gvr)
			// r.triggerReconcileForOwner(obj.(*unstructured.Unstructured))
		},
	})

	stopCh := r.setupStopChannel()
	log.Printf("[WATCH] Started watching: %s", gvr.String())
	informer.Run(stopCh)
}

func (r *InfrahubResourceReconciler) handleLabeledResource(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	log.Printf("[WATCH] Change detected on resource: %s/%s (%s)", obj.GetNamespace(), obj.GetName(), gvr.Resource)
	// TODO: Optionally map this to reconcile relevant InfrahubResource
}

var stopOnce sync.Once
var stopChan chan struct{}

func (r *InfrahubResourceReconciler) setupStopChannel() <-chan struct{} {
	stopOnce.Do(func() {
		stopChan = make(chan struct{})
	})
	return stopChan
}

func (r *InfrahubResourceReconciler) triggerReconcileForOwner(dep *unstructured.Unstructured) {
	for _, owner := range dep.GetOwnerReferences() {
		if owner.Kind == "InfrahubResource" && owner.APIVersion == infrahubv1alpha1.GroupVersion.String() {
			var res infrahubv1alpha1.InfrahubResource
			err := r.Client.Get(context.Background(), types.NamespacedName{
				Name: owner.Name,
			}, &res)
			if err != nil {
				log.Printf("[WATCH] Failed to get InfrahubResource %s: %v", owner.Name, err)
				return
			}

			// Touch the resource to update its annotation (this triggers reconcile)
			if res.Annotations == nil {
				res.Annotations = map[string]string{}
			}
			res.Spec.ReconciledAt = v1.Time{Time: time.Now()}

			if err := r.Client.Update(context.Background(), &res); err != nil {
				log.Printf("[WATCH] Failed to patch InfrahubResource %s/%s: %v", res.Namespace, res.Name, err)
			} else {
				log.Printf("[WATCH] Triggered reconcile of InfrahubResource %s/%s", res.Namespace, res.Name)
			}
		}
	}
}
