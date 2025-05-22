package controller

// import (
// 	"context"
// 	"log"
// 	"time"

// 	infrahubv1alpha1 "github.com/simli1333/vidra/api/v1alpha1"
// 	appsv1 "k8s.io/api/apps/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/client-go/informers"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/rest"
// 	"k8s.io/client-go/tools/cache"
// 	"sigs.k8s.io/controller-runtime/pkg/event"
// 	"sigs.k8s.io/controller-runtime/pkg/predicate"
// )

// func (r *InfrahubResourceReconciler) startDeploymentWatcher(cfg *rest.Config) {
// 	k8sClient, err := kubernetes.NewForConfig(cfg)
// 	if err != nil {
// 		log.Printf("failed to create k8s client for deployment watcher: %v", err)
// 		return
// 	}

// 	factory := informers.NewSharedInformerFactory(k8sClient, 30*time.Second)
// 	informer := factory.Apps().V1().Deployments().Informer()

// 	genChanged := predicate.GenerationChangedPredicate{}

// 	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
// 		UpdateFunc: func(oldObj, newObj interface{}) {
// 			oldDep := oldObj.(*appsv1.Deployment)
// 			newDep := newObj.(*appsv1.Deployment)

// 			// Only trigger if GenerationChangedPredicate returns true
// 			if genChanged.Update(event.UpdateEvent{
// 				ObjectOld: oldDep,
// 				ObjectNew: newDep,
// 			}) {
// 				log.Printf("[WATCH] Deployment generation changed: %s/%s", newDep.Namespace, newDep.Name)
// 				r.triggerReconcileForOwner(newDep)
// 			}
// 		},
// 		DeleteFunc: func(obj interface{}) {
// 			dep := obj.(*appsv1.Deployment)
// 			log.Printf("[WATCH] Deployment deleted: %s/%s", dep.Namespace, dep.Name)
// 			r.triggerReconcileForOwner(dep)
// 		},
// 	})

// 	go factory.Start(r.setupStopChannel())
// }

// func (r *InfrahubResourceReconciler) setupStopChannel() <-chan struct{} {
// 	stop := make(chan struct{})
// 	// You can store this on r if you want to stop informers on shutdown
// 	return stop
// }

// func (r *InfrahubResourceReconciler) triggerReconcileForOwner(dep *appsv1.Deployment) {
// 	for _, owner := range dep.GetOwnerReferences() {
// 		if owner.Kind == "InfrahubResource" && owner.APIVersion == infrahubv1alpha1.GroupVersion.String() {
// 			var res infrahubv1alpha1.InfrahubResource
// 			err := r.Client.Get(context.Background(), types.NamespacedName{
// 				Name:      owner.Name,
// 				Namespace: dep.Namespace,
// 			}, &res)
// 			if err != nil {
// 				log.Printf("[WATCH] Failed to get InfrahubResource %s/%s: %v", dep.Namespace, owner.Name, err)
// 				return
// 			}

// 			// Touch the resource to update its annotation (this triggers reconcile)
// 			if res.Annotations == nil {
// 				res.Annotations = map[string]string{}
// 			}
// 			res.Spec.ReconciledAt = metav1.Time{Time: time.Now()}

// 			if err := r.Client.Update(context.Background(), &res); err != nil {
// 				log.Printf("[WATCH] Failed to patch InfrahubResource %s/%s: %v", res.Namespace, res.Name, err)
// 			} else {
// 				log.Printf("[WATCH] Triggered reconcile of InfrahubResource %s/%s", res.Namespace, res.Name)
// 			}
// 		}
// 	}
// }
