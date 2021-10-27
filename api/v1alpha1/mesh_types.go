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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reference: https://book.kubebuilder.io/reference/markers/crd.html

// MeshSpec defines the desired state of a Grey Matter mesh.
type MeshSpec struct {

	// The base url of the openshift cluster
	// (ex: console-openshift-console.apps.coffee.greymatter.services would be apps.coffee.greymatter.services)
	ClusterUrl string `json:"cluster_url"`

	// The version of Grey Matter to install for this mesh.
	// +kubebuilder:validation:Enum="1.6";"1.7"
	// +kubebuilder:default="1.7"
	ReleaseVersion string `json:"release_version"`

	// Label this mesh as belonging to a particular zone.
	// +kubebuilder:default=default-zone
	Zone string `json:"zone"`

	// Namespace where mesh core components and dependencies should be installed.
	InstallNamespace string `json:"install_namespace"`

	// Namespaces to include in the mesh network.
	// +optional
	WatchNamespaces []string `json:"watch_namespaces,omitempty"`

	// Add user tokens to the JWT Security Service.
	// +optional
	UserTokens []UserToken `json:"user_tokens,omitempty"`

	// Adds an external Redis provider for caching Grey Matter configuration state.
	// +optional
	ExternalRedis *ExternalRedisConfig `json:"redis,omitempty"`
}

// MeshStatus describes the observed state of a Grey Matter mesh.
type MeshStatus struct {
}

// Markers for generating manifests for webhooks.
//+kubebuilder:webhook:path=/mutate-mesh,mutating=true,failurePolicy=fail,sideEffects=None,groups=greymatter.io,resources=meshes,verbs=create;update,versions=v1alpha1,name=mutate-mesh.greymatter.io,admissionReviewVersions={v1,v1beta1}
//+kubebuilder:webhook:path=/validate-mesh,mutating=false,failurePolicy=fail,sideEffects=None,groups=greymatter.io,resources=meshes,verbs=create;update;delete,versions=v1alpha1,name=validate-mesh.greymatter.io,admissionReviewVersions={v1,v1beta1}
//+kubebuilder:webhook:path=/mutate-workload,mutating=true,failurePolicy=fail,sideEffects=None,groups=core;apps,resources=pods;deployments;statefulsets,verbs=create;update;delete,versions=v1,name=mutate-workload.greymatter.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Mesh defines a Grey Matter mesh's desired state and describes its observed state.
type Mesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:Required
	Spec   MeshSpec   `json:"spec,omitempty"`
	Status MeshStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MeshList contains a list of Mesh custom resources managed by the Grey Matter Operator.
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mesh{}, &MeshList{})
}
