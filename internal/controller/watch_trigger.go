//go:build codecovskip
// +build codecovskip

package controller

import (
	"context"
	"log"
	"time"

	infrahubv1alpha1 "github.com/infrahub-operator/vidra/api/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Codecov: skip file as the test does not run in gh actions
// Callback function to warch_resources_factory

func (r *VidraResourceReconciler) handleLabeledResource(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	log.Printf("[WATCH] Change detected on resource: %s/%s (%s)", obj.GetNamespace(), obj.GetName(), gvr.Resource)
	r.triggerReconcileForOwner(obj)
}

func (r *VidraResourceReconciler) triggerReconcileForOwner(obj *unstructured.Unstructured) {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Kind == "VidraResource" && owner.APIVersion == infrahubv1alpha1.GroupVersion.String() {
			var res infrahubv1alpha1.VidraResource
			err := r.Get(context.Background(), types.NamespacedName{
				Name:      owner.Name,
				Namespace: obj.GetNamespace(), // You were missing this
			}, &res)
			if err != nil {
				log.Printf("[WATCH] Failed to get VidraResource %s/%s: %v", obj.GetNamespace(), owner.Name, err)
				continue
			}

			if res.Spec.ReconciledAt.Time.Before(time.Now().Add(-2 * time.Second)) {
				res.Spec.ReconciledAt = v1.Time{Time: time.Now()}
				if err := r.Update(context.Background(), &res); err != nil {
					log.Printf("[WATCH] Failed to update VidraResource %s/%s: %v", res.Namespace, res.Name, err)
				} else {
					log.Printf("[WATCH] Triggered reconcile of VidraResource %s/%s", res.Namespace, res.Name)
				}
			} else {
				log.Printf("[WATCH] VidraResource %s/%s is already up-to-date, skipping reconcile trigger", res.Namespace, res.Name)
			}
		}
	}
}
