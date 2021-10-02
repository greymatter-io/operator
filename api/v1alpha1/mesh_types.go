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

package v1alpha1

import (
	"github.com/greymatter-io/operator/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reference: https://book.kubebuilder.io/reference/markers/crd.html

// Defines the desired state of a Grey Matter mesh.
type MeshSpec struct {
	// +kubebuilder:validation:Enum="1.6";"1.7"
	// +kubebuilder:default="1.6"
	ReleaseVersion string `json:"release_version"`
	// Adds an external Redis provider for caching Grey Matter configuration state.
	// +optional
	// +nullable
	ExternalRedis *version.ExternalRedisConfig `json:"redis,omitempty"`
}

// Describes the observed state of a Grey Matter mesh.
type MeshStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// The schema used to define a Grey Matter mesh's desired state and describe its observed state.
type Mesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:Required
	Spec   MeshSpec   `json:"spec,omitempty"`
	Status MeshStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Contains a list of Mesh custom resources managed by the Grey Matter Operator.
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mesh{}, &MeshList{})
}
