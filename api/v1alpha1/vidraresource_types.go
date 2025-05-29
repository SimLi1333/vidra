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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VidraResourceSpec defines the desired state of VidraResource
type VidraResourceSpec struct {
	// Destination contains the destination information for the resource
	Destination InfrahubSyncDestination `json:"destination,omitempty" protobuf:"bytes,2,name=destination"`

	// Manifest contains the manifest information for the resource
	Manifest string `json:"manifest,omitempty" protobuf:"bytes,2,name=manifest"`

	// If true, the operator will reconcile resources based on k8s events. (default: false)
	// +kubebuilder:default:=false
	ReconcileOnEvents bool `json:"reconcileOnEvents,omitempty" protobuf:"varint,4,opt,name=reconcileOnEvents"`

	// The last time the resource was reconciled
	ReconciledAt metav1.Time `json:"reconciledAt,omitempty" protobuf:"bytes,5,name=reconciledAt"`
}

// VidraResourceStatus defines the observed state of VidraResource
type VidraResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ManagedResources contains a list of resources managed by this VidraResource
	ManagedResources []ManagedResourceStatus `json:"managedResources,omitempty"`

	// DeployState indicates the current state of the deployment
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Stale
	DeployState State `json:"DeployState,omitempty"`
	// LastError contains the last error message if any
	LastError string `json:"lastError,omitempty"`
	// LastSyncTime indicates the last time the resource was synchronized
	LastSyncTime metav1.Time `json:"lastSyncTime,omitempty"`
}

type ManagedResourceStatus struct {
	// Kind of the resource (e.g., Deployment, Service)
	Kind string `json:"kind"`
	// APIVersion of the resource (e.g., apps/v1)
	APIVersion string `json:"apiVersion"`
	// Name of the resource
	Name string `json:"name"`
	// Namespace of the resource
	Namespace string `json:"namespace,omitempty"`
}

type State string

const (
	StateRunning   State = "Running"
	StateSucceeded State = "Succeeded"
	StateFailed    State = "Failed"
	StateStale     State = "Stale"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// VidraResource is the Schema for the Vidraresources API
type VidraResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// VidraResourceSpec defines the desired state of VidraResource
	Spec VidraResourceSpec `json:"spec,omitempty"`
	// VidraResourceStatus defines the observed state of VidraResource
	Status VidraResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VidraResourceList contains a list of VidraResource
type VidraResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VidraResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VidraResource{}, &VidraResourceList{})
}
