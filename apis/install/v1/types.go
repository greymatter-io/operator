package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Operator is a top-level type. A client is created for it.
type Operator struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OperatorSpec `json:"spec,omitempty"`
	// +optional
	Status OperatorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorList is a top-level list type. The client methods for lists are automatically created.
// You are not supposed to create a separated client for this one.
type OperatorList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Operator `json:"items"`
}

type OperatorSpec struct {
	Profile string `json:"profile"`
}

type OperatorStatus struct {
	Description string `json:"status,omitempty"`
}
