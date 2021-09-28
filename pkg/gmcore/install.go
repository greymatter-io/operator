package gmcore

import (
	"context"
	"fmt"

	"github.com/greymatter-io/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Installs and updates Grey Matter core components and dependencies.
// Also labels namespaces specified in each Mesh CR to trigger meshobject configuration
// for their deployments and statefulsets, as well as sidecar injection for their pods.
// If auto-inject is enabled (default=true), labels each existing appsv1.Deployment/StatefulSet
// plus their pod templates so that those workloads are added to the mesh automatically.
func (i *Installer) ApplyMesh(c client.Client, mesh v1alpha1.Mesh) error {
	// TODO: Assign version once we have the value in our spec. For now use v1.6 by default
	version := "v1.6"
	base, ok := i.baseValues[version]
	if !ok {
		return fmt.Errorf("unknown version %s", version)
	}

	// Get proxy values for this
	i.proxyValues[mesh.Name] = mesh.ProxyValues(base)

	// Generate resources with owner references to the Mesh
	resources := mesh.ResourceGroups(base, i.scheme)

	for _, group := range resources {
		// TODO: get and update or create
		c.Create(context.TODO(), group.Deployment)
		for _, service := range group.Services {
			c.Create(context.TODO(), service)
		}
	}

	return nil
}

// Removes all resources created for installing Grey Matter core components of a mesh.
// Also removes the mesh from labels of resources (removing the labels if no more meshes are listed)
// i.e. namespaces, deployments, statefulsets, and pods (via pod templates).
func (i *Installer) RemoveMesh(c client.Client, name string) {
}

// Given a slice of corev1.Containers, injects sidecar(s) to enable traffic for each mesh specified.
func (i *Installer) Inject(containers []corev1.Container, xdsCluster string, meshes []string) []corev1.Container {
	return containers
}
