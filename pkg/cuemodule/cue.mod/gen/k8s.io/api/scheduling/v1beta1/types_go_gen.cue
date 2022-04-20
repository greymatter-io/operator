// Code generated by cue get go. DO NOT EDIT.

//cue:generate cue get go k8s.io/api/scheduling/v1beta1

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/api/core/v1"
)

// DEPRECATED - This group version of PriorityClass is deprecated by scheduling.k8s.io/v1/PriorityClass.
// PriorityClass defines mapping from a priority class name to the priority
// integer value. The value can be any valid integer.
#PriorityClass: {
	metav1.#TypeMeta

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metadata?: metav1.#ObjectMeta @go(ObjectMeta) @protobuf(1,bytes,opt)

	// The value of this priority class. This is the actual priority that pods
	// receive when they have the name of this class in their pod spec.
	value: int32 @go(Value) @protobuf(2,bytes,opt)

	// globalDefault specifies whether this PriorityClass should be considered as
	// the default priority for pods that do not have any priority class.
	// Only one PriorityClass can be marked as `globalDefault`. However, if more than
	// one PriorityClasses exists with their `globalDefault` field set to true,
	// the smallest value of such global default PriorityClasses will be used as the default priority.
	// +optional
	globalDefault?: bool @go(GlobalDefault) @protobuf(3,bytes,opt)

	// description is an arbitrary string that usually provides guidelines on
	// when this priority class should be used.
	// +optional
	description?: string @go(Description) @protobuf(4,bytes,opt)

	// PreemptionPolicy is the Policy for preempting pods with lower priority.
	// One of Never, PreemptLowerPriority.
	// Defaults to PreemptLowerPriority if unset.
	// This field is beta-level, gated by the NonPreemptingPriority feature-gate.
	// +optional
	preemptionPolicy?: null | apiv1.#PreemptionPolicy @go(PreemptionPolicy,*apiv1.PreemptionPolicy) @protobuf(5,bytes,opt)
}

// PriorityClassList is a collection of priority classes.
#PriorityClassList: {
	metav1.#TypeMeta

	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metadata?: metav1.#ListMeta @go(ListMeta) @protobuf(1,bytes,opt)

	// items is the list of PriorityClasses
	items: [...#PriorityClass] @go(Items,[]PriorityClass) @protobuf(2,bytes,rep)
}
