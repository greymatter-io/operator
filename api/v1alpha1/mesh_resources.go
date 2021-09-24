package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type ResourceGroup struct {
	Deployment *appsv1.Deployment
	Services   []*corev1.Service
}

// Generates resources necessary for installing Grey Matter core components and dependencies
// with options configured in the Mesh CR.
func (m Mesh) ResourceGroups(base *InstallValues, scheme *runtime.Scheme) []ResourceGroup {
	// copied := base.DeepCopy()
	// TODO: apply overlays
	var groups []ResourceGroup
	// TODO: SetControllerReference
	return groups
}

// Generates the values used for sidecar injection into a pod
// with options configured in the Mesh CR.
func (m Mesh) ProxyValues(base *InstallValues) *Values {
	// TODO
	return &Values{}
}
