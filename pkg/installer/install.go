package installer

import (
	"context"

	"github.com/greymatter-io/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Installs and updates Grey Matter core components and dependencies.
// Also labels namespaces specified in each Mesh CR to trigger meshobject configuration
// for their deployments and statefulsets, as well as sidecar injection for their pods.
// If auto-inject is enabled (default=true), labels each existing appsv1.Deployment/StatefulSet
// plus their pod templates so that those workloads are added to the mesh automatically.
func (i *Installer) ApplyMesh(c client.Client, mesh *v1alpha1.Mesh) {

	// DeepCopy base values for the Grey Matter version specified, so the original is not mutated.
	// TODO: Assign version once we have the value in our spec. For now, use v1.6 by default.
	// TODO: Use the Mesh validating webhook to ensure mesh.Spec.ReleaseVersion is valid.
	values := i.baseValues[mesh.Spec.ReleaseVersion].DeepCopy()

	// Apply options defined in Mesh CR
	values.Apply(mesh.InstallOpts()...)

	// Generate manifests from from values and send them to the K8s apiserver.
	for _, group := range values.GenerateManifests() {
		// Set an owner reference on the manifest for garbage collection if the mesh is deleted.
		controllerutil.SetOwnerReference(mesh, group.Deployment, i.scheme)
		// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/controller/controllerutil/example_test.go#L35
		// TODO: Ensure no mutateFn callback is needed for updates.
		// TODO: Use controllerutil.OperationResult to update Mesh.Status with an event stream of what was created, updated, etc.
		controllerutil.CreateOrUpdate(context.TODO(), c, group.Deployment, func() error { return nil })
		for _, service := range group.Services {
			controllerutil.SetOwnerReference(mesh, service, i.scheme)
			controllerutil.CreateOrUpdate(context.TODO(), c, service, func() error { return nil })
		}
	}
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
