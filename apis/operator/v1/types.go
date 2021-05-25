package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mesh is a top-level type. A client is created for it.
type Mesh struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MeshSpec `json:"spec,omitempty"`
	// +optional
	Status MeshStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeshList is a top-level list type. The client methods for lists are automatically created.
// You are not supposed to create a separated client for this one.
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

type MeshSpec struct {
	Version string `json:"version"`
}

type MeshStatus struct {
	Deployed bool `json:"deployed,omitempty"`
}
