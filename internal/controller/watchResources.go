package controller

import (
	"context"
	"log"
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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DynamicWatcherFactory struct {
	mu       sync.Mutex
	started  map[schema.GroupVersionResource]struct{}
	stopChan chan struct{}
}

func NewDynamicWatcherFactory() *DynamicWatcherFactory {
	return &DynamicWatcherFactory{
		started:  make(map[schema.GroupVersionResource]struct{}),
		stopChan: make(chan struct{}),
	}
}

type ResourceCallback func(obj *unstructured.Unstructured, gvr schema.GroupVersionResource)

func (f *DynamicWatcherFactory) StartWatchingGVRs(
	dynamicClient dynamic.Interface,
	gvrs []schema.GroupVersionResource,
	onEvent ResourceCallback,
) {
	for _, gvr := range gvrs {
		f.mu.Lock()
		if _, ok := f.started[gvr]; ok {
			f.mu.Unlock()
			continue // already watching
		}
		f.started[gvr] = struct{}{}
		f.mu.Unlock()

		go f.watchGVR(dynamicClient, gvr, onEvent)
	}
}

func (f *DynamicWatcherFactory) watchGVR(
	client dynamic.Interface,
	gvr schema.GroupVersionResource,
	onEvent ResourceCallback,
) {
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
			unstrObj, ok := obj.(*unstructured.Unstructured)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Printf("AddFunc: unexpected type %T", obj)
					return
				}
				unstrObj, ok = tombstone.Obj.(*unstructured.Unstructured)
				if !ok {
					log.Printf("AddFunc: unexpected tombstone type %T", tombstone.Obj)
					return
				}
			}
			onEvent(unstrObj, gvr)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newUnstr, ok1 := newObj.(*unstructured.Unstructured)
			oldUnstr, ok2 := oldObj.(*unstructured.Unstructured)
			if !ok1 || !ok2 {
				log.Printf("UpdateFunc: unexpected types old=%T new=%T", oldObj, newObj)
				return
			}
			if genChanged.Update(event.UpdateEvent{
				ObjectOld: oldUnstr,
				ObjectNew: newUnstr,
			}) {
				onEvent(newUnstr, gvr)
			}
		},
		DeleteFunc: func(obj interface{}) {
			unstrObj, ok := obj.(*unstructured.Unstructured)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Printf("DeleteFunc: unexpected type %T", obj)
					return
				}
				unstrObj, ok = tombstone.Obj.(*unstructured.Unstructured)
				if !ok {
					log.Printf("DeleteFunc: unexpected tombstone type %T", tombstone.Obj)
					return
				}
			}
			onEvent(unstrObj, gvr)
		},
	})

	log.Printf("[WATCH] Started watching: %s", gvr.String())
	informer.Run(f.stopChan)
}

func (r *InfrahubResourceReconciler) handleLabeledResource(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	log.Printf("[WATCH] Change detected on resource: %s/%s (%s)", obj.GetNamespace(), obj.GetName(), gvr.Resource)
	r.triggerReconcileForOwner(obj)
}

func (r *InfrahubResourceReconciler) triggerReconcileForOwner(obj *unstructured.Unstructured) {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Kind == "InfrahubResource" && owner.APIVersion == infrahubv1alpha1.GroupVersion.String() {
			var res infrahubv1alpha1.InfrahubResource
			err := r.Client.Get(context.Background(), types.NamespacedName{
				Name:      owner.Name,
				Namespace: obj.GetNamespace(), // You were missing this
			}, &res)
			if err != nil {
				log.Printf("[WATCH] Failed to get InfrahubResource %s/%s: %v", obj.GetNamespace(), owner.Name, err)
				continue
			}

			res.Spec.ReconciledAt = v1.Time{Time: time.Now()}

			if err := r.Client.Update(context.Background(), &res); err != nil {
				log.Printf("[WATCH] Failed to update InfrahubResource %s/%s: %v", res.Namespace, res.Name, err)
			} else {
				log.Printf("[WATCH] Triggered reconcile of InfrahubResource %s/%s", res.Namespace, res.Name)
			}
		}
	}
}
