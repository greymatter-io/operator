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

	// TODO: Use the Mesh validating webhook to ensure mesh.Spec.ReleaseVersion is valid.
	version := i.versions[mesh.Spec.ReleaseVersion].Copy()

	// Apply options defined in Mesh CR
	version.Apply(mesh.InstallOptions()...)
	manifests := version.Manifests()

	// Save this mesh's Proxy InstallConfig to use later for sidecar injection
	i.sidecars[mesh.Name] = version.Sidecar()

	// Generate manifests from install configs and send them to the K8s apiserver.
INSTALL_LOOP:
	for _, group := range manifests {
		if group.Deployment.Name == "greymatter-redis" && mesh.Spec.ExternalRedis != nil && mesh.Spec.ExternalRedis.URL != "" {
			continue INSTALL_LOOP
		}
		// Set an owner reference on the manifest for garbage collection if the mesh is deleted.
		controllerutil.SetOwnerReference(mesh, group.Deployment, scheme)
		// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/controller/controllerutil/example_test.go#L35
		// TODO: Ensure no mutateFn callback is needed for updates.
		// TODO: Use controllerutil.OperationResult to update Mesh.Status with an event stream of what was created, updated, etc.
		controllerutil.CreateOrUpdate(context.TODO(), c, group.Deployment, func() error { return nil })
		for _, service := range group.Services {
			controllerutil.SetOwnerReference(mesh, service, scheme)
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
