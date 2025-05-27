package k8s

import (
	"context"
	"log"
	"sync"

	"github.com/simli1333/vidra/internal/domain"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func (f *DynamicWatcherFactory) StartWatchingGVRs(
	dynamicClient dynamic.Interface,
	gvrs []schema.GroupVersionResource,
	onEvent domain.ResourceCallback,
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
	onEvent domain.ResourceCallback,
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
