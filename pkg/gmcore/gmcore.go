// Package gmcore exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package gmcore

import (
	"github.com/greymatter-io/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// References a Grey Matter version for each mesh
type Versions struct {
	each map[string]string // TODO: Expand as options get added to Mesh CR
}

// Returns *Versions for tracking which Grey Matter version is installed for each mesh
func New() *Versions {
	return &Versions{make(map[string]string)}
}

// Applies all resources necessary for installing Grey Matter core components of a mesh.
// Also labels namespaces specified in each Mesh CR to trigger meshobject configuration
// for their deployments and statefulsets, as well as sidecar injection for their pods.
// If auto-inject is enabled (default=true), labels each existing appsv1.Deployment/StatefulSet
// plus their pod templates so that those workloads are added to the mesh automatically.
func (v *Versions) ApplyMesh(c client.Client, mesh v1alpha1.Mesh) {
}

// Removes all resources created for installing Grey Matter core components of a mesh.
// Also removes the mesh from labels of resources (removing the labels if no more meshes are listed)
// i.e. namespaces, deployments, statefulsets, and pods (via pod templates).
func (v *Versions) RemoveMesh(c client.Client, name string) {
}

// Given a slice of corev1.Containers, injects sidecar(s) to enable traffic for each mesh specified.
func (v *Versions) Inject(containers []corev1.Container, xdsCluster string, meshes []string) []corev1.Container {
	return containers
}
