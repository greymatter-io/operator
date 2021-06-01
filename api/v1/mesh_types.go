/*
Copyright Decipher Technology Studios 2021.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Defines the desired state of a Mesh.
type MeshSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Which version of Grey Matter to install.
	// If not specified, the latest version will be installed.
	// +optional
	Version string `json:"version,omitempty"`

	// The name of the secret used for pulling Grey Matter service Docker images.
	// If not specified, defaults to "docker.secret".
	// +optional
	ImagePullSecret string `json:"image_pull_secret,omitempty"`
}

// Defines the observed state of a Mesh.
type MeshStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Whether the mesh was deployed or not.
	Deployed bool `json:"deployed"`

	// TODO
	// Conditions []Condition `json:"conditions"`
}

type Condition struct {
	// The type of the Mesh condition
	Type string `json:"type"`
	// The status of the condition, one of True, False, Unknown
	Status string `json:"status"`
	// A one-word camelCase reason for the condition's last transition
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human-readable message indicating details about last transition
	// +optional
	Message string `json:"message,omitempty"`
	// The last time the condition transitioned from one status to another
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Mesh is the Schema for the meshes API
type Mesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshSpec   `json:"spec,omitempty"`
	Status MeshStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MeshList contains a list of Mesh
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mesh{}, &MeshList{})
}
