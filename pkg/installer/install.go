package installer

import (
	"context"
	"fmt"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Installs and updates Grey Matter core components and dependencies.
// Also labels namespaces specified in each Mesh CR to trigger meshobject configuration
// for their deployments and statefulsets, as well as sidecar injection for their pods.
// If auto-inject is enabled (default=true), labels each existing appsv1.Deployment/StatefulSet
// plus their pod templates so that those workloads are added to the mesh automatically.
func (i *Installer) ApplyMesh(mesh *v1alpha1.Mesh, init bool) {

	// Get a copy of the version specified in the Mesh CR.
	// Assume the value is valid since the CRD enumerates acceptable values for the apiserver.
	v := i.versions[mesh.Spec.ReleaseVersion].Copy()

	// Apply options for mutating the version copy's internal Cue value
	// These options are defined in the Mesh CR as well as the bootstrap config.
	opts := append(mesh.InstallOptions(), version.ImagePullSecretName(i.imagePullSecret.Name))
	v.Apply(opts...)

	// Generate manifests from install configs and send them to the apiserver.
	manifests := v.Manifests()

	// Save this mesh's Proxy InstallConfig to use later for sidecar injection
	i.sidecars[mesh.Name] = v.Sidecar()

	// Obtain the scheme used by our client for
	scheme := i.client.Scheme()

	// Create the imagePullSecret in the namespace prior to installing core services
	if init {
		secret := i.imagePullSecret.DeepCopy()
		secret.Namespace = mesh.Namespace
		i.apply(secret, mesh, scheme)
	}

MANIFEST_LOOP:
	for _, group := range manifests {
		if group.Deployment.Name == "gm-redis" && mesh.Spec.ExternalRedis.URL != "" {
			continue MANIFEST_LOOP
		}
		if group.ServiceAccount != nil {
			i.apply(group.ServiceAccount, mesh, scheme)
		}
		for _, configMap := range group.ConfigMaps {
			i.apply(configMap, mesh, scheme)
		}
		i.apply(group.Deployment, mesh, scheme)
		for _, service := range group.Services {
			i.apply(service, mesh, scheme)
		}
	}
}

func (i *Installer) apply(obj client.Object, mesh *v1alpha1.Mesh, scheme *runtime.Scheme) error {
	var kind string
	gvk, err := apiutil.GVKForObject(obj.(runtime.Object), scheme)
	if err != nil {
		kind = "Object"
	} else {
		kind = gvk.Kind
	}

	// Set an owner reference on the manifest for garbage collection if the mesh is deleted.
	if err := controllerutil.SetOwnerReference(mesh, obj, scheme); err != nil {
		logger.Error(err, "Failed SetOwnerReference", "Namespace", mesh.Namespace, "Mesh", mesh.Name, kind, obj.GetName())
		return err
	}
	// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/controller/controllerutil/example_test.go#L35
	// TODO: Ensure no mutateFn callback is needed for updates.
	// TODO: Use controllerutil.OperationResult to update Mesh.Status with an event stream of what was created, updated, etc.
	fmt.Printf("CreateOrUpdate: %#v\n", obj)
	if _, err := controllerutil.CreateOrUpdate(context.TODO(), i.client, obj, func() error { return nil }); err != nil {
		logger.Error(err, "Failed CreateOrUpdate", "Namespace", mesh.Namespace, "Mesh", mesh.Name, kind, obj.GetName())
		return err
	}

	return nil
}

// TODO: Add service account to cluster role binding subjects list.
func (i *Installer) linkServiceAccount(sa *corev1.ServiceAccount) error {
	// unimplemented
	return nil
}

// Removes all resources created for installing Grey Matter core components of a mesh.
// Also removes the mesh from labels of resources (removing the labels if no more meshes are listed)
// i.e. namespaces, deployments, statefulsets, and pods (via pod templates).
func (i *Installer) RemoveMesh(name string) {
}

// Given a slice of corev1.Containers, injects sidecar(s) to enable traffic for each mesh specified.
func (i *Installer) Inject(containers []corev1.Container, xdsCluster string, meshes []string) []corev1.Container {
	return containers
}
