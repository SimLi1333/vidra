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

// InfrahubResourceSpec defines the desired state of InfrahubResource
type InfrahubResourceSpec struct {
	// Source contains the source information for the Infrahub API interaction
	Source InfrahubSyncSource `json:"source" protobuf:"bytes,1,name=source"`

	// Destination contains the destination information for the resource
	Destination InfrahubSyncDestination `json:"destination,omitempty" protobuf:"bytes,2,name=destination"`

	// IDs contains important identifiers for the resource
	IDs InfrahubResourceIDs `json:"ids,omitempty" protobuf:"bytes,3,name=ids"`
}

// InfrahubResourceIDs contains identifiers for the resource
type InfrahubResourceIDs struct {
	// Unique identifier for the artifact
	// +kubebuilder:validation:Required
	ArtifactID string `json:"artefactID,omitempty" protobuf:"bytes,1,name=artefactID"`

	// Checksum of the artifact
	// +kubebuilder:validation:Required
	Checksum string `json:"checksum,omitempty" protobuf:"bytes,2,name=checksum"`

	// Storage ID for the artifact
	// +kubebuilder:validation:Required
	StorageID string `json:"storageID,omitempty" protobuf:"bytes,3,name=storageID"`
}

// InfrahubResourceStatus defines the observed state of InfrahubResource
type InfrahubResourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Checksum         string                  `json:"checksum,omitempty"`
	ManagedResources []ManagedResourceStatus `json:"managedResources,omitempty"`
	Manifests        string                  `json:"manifests,omitempty"`
	LastSyncTime     metav1.Time             `json:"lastSyncTime,omitempty"`
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Stale
	DeployState State  `json:"DeployState,omitempty"`
	LastError   string `json:"lastError,omitempty"`
}

type ManagedResourceStatus struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
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

// InfrahubResource is the Schema for the infrahubresources API
type InfrahubResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InfrahubResourceSpec   `json:"spec,omitempty"`
	Status InfrahubResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InfrahubResourceList contains a list of InfrahubResource
type InfrahubResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfrahubResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InfrahubResource{}, &InfrahubResourceList{})
}
