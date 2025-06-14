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

// InfrahubSyncSpec defines the desired state of InfrahubSync
type InfrahubSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of InfrahubSync. Edit infrahubsync_types.go to remove/update
	// Source contains the source information for the Infrahub API interaction
	Source InfrahubSyncSource `json:"source" protobuf:"bytes,1,name=source"`

	// Destination contains the destination information for the resource
	Destination InfrahubSyncDestination `json:"destination,omitempty" protobuf:"bytes,2,name=destination"`
}

// VidraResourceSource contains the source information for the resource
type InfrahubSyncSource struct {
	// URL for the Infrahub API (e.g., https://infrahub.example.com)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^(http|https)://[a-zA-Z0-9.-]+(:[0-9]+)?(?:/[a-zA-Z0-9-]+)*$"
	InfrahubAPIURL string `json:"infrahubAPIURL" protobuf:"bytes,1,name=infrahubAPIURL"`

	// The target branch in Infrahub to interact with
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default:="main"
	TargetBranch string `json:"targetBranch" protobuf:"bytes,2,name=targetBranch"`

	// The target date in Infrahub for all the interactions (e.g., "2025-01-01T00:00:00Z or -2d" for the artifact from two days ago). If not set, the operator will use the current date.
	// +kubebuilder:validation:Optional
	TargetDate string `json:"targetDate,omitempty" protobuf:"bytes,3,name=targetDate"`

	// Artifact name that is being handled by the operator, this is used to identify the resource in Infrahub
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ArtifactName string `json:"artefactName" protobuf:"bytes,4,name=artefactName"`
}

// VidraResourceDestination contains information about where the resource will be sent
type InfrahubSyncDestination struct {
	// Only needed if you need to deploy to two Kubernetis cluster (multicluster) if set to "httlps://kubernetes.default.svc" or omitted, the operator will use the current cluster
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="^(http|https)://[a-zA-Z0-9.-]+(:[0-9]+)?(?:/[a-zA-Z0-9-]+)*$"
	Server string `json:"server,omitempty" protobuf:"bytes,1,name=server"`

	// Default Namespace in the Kubernetes cluster where the resource should be sent, if they do not hava a namespace already set
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,name=namespace"`

	// If true, the operator will reconcile resources based on k8s events. (default: false) - changes to the resource will trigger a reconciliation
	// +kubebuilder:default:=false
	ReconcileOnEvents bool `json:"reconcileOnEvents,omitempty" protobuf:"varint,4,opt,name=reconcileOnEvents"`
}

// InfrahubSyncStatus defines the observed state of InfrahubSync
type InfrahubSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Checksums contains a list of checksums for synced resources
	Checksums []string `json:"checksums,omitempty"`

	// SyncState indicates the current state of the sync operation
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Stale
	SyncState State `json:"syncState,omitempty"`

	// LastError provides details about the last error encountered during the sync operation
	LastError string `json:"lastError,omitempty"`

	// LastSyncTime indicates the last time the sync operation was performed
	LastSyncTime metav1.Time `json:"lastSyncTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// InfrahubSync is the Schema for the infrahubsyncs API
type InfrahubSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of InfrahubSync
	Spec InfrahubSyncSpec `json:"spec,omitempty"`
	// Status defines the observed state of InfrahubSync
	Status InfrahubSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InfrahubSyncList contains a list of InfrahubSync
type InfrahubSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfrahubSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InfrahubSync{}, &InfrahubSyncList{})
}
