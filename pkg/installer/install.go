package installer

import (
	"context"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/version"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Apply options for mutating the version copy's internal Cue value.
	opts := append(
		mesh.InstallOptions(),
		version.ImagePullSecretName(i.imagePullSecret.Name),
		version.JWTSecrets,
	)
	v.Apply(opts...)

	// Generate manifests from install configs and send them to the apiserver.
	manifests := v.Manifests()

	// Save this mesh's Proxy InstallConfig to use later for sidecar injection
	i.sidecars[mesh.Name] = v.Sidecar()

	// Obtain the scheme used by our client for
	scheme := i.client.Scheme()

	// TODO: ERROR HANDLE APPLY CALLS !!!

	// Create a Docker image pull secret and service account in this namespace if this Mesh is new.
	if init {
		secret := i.imagePullSecret.DeepCopy()
		secret.Namespace = mesh.Namespace
		i.apply(secret, mesh, scheme)

		sa := i.serviceAccount.DeepCopy()
		sa.Namespace = mesh.Namespace
		i.applyServiceAccount(sa, mesh, scheme)
	}

MANIFEST_LOOP:
	for _, group := range manifests {
		// If an external Redis server is configured, don't install an internal Redis.
		if group.StatefulSet != nil &&
			group.StatefulSet.Name == "gm-redis" &&
			mesh.Spec.ExternalRedis != nil &&
			mesh.Spec.ExternalRedis.URL != "" {
			continue MANIFEST_LOOP
		}

		for _, configMap := range group.ConfigMaps {
			i.apply(configMap, mesh, scheme)
		}
		for _, secret := range group.Secrets {
			i.apply(secret, mesh, scheme)
		}
		for _, service := range group.Services {
			i.apply(service, mesh, scheme)
		}
		if group.Deployment != nil {
			i.apply(group.Deployment, mesh, scheme)
		}
		if group.StatefulSet != nil {
			i.apply(group.StatefulSet, mesh, scheme)
		}
	}
}

func (i *Installer) apply(obj client.Object, mesh *v1alpha1.Mesh, scheme *runtime.Scheme) error {
	var kind string
	if gvk, err := apiutil.GVKForObject(obj.(runtime.Object), scheme); err != nil {
		kind = "Object"
	} else {
		kind = gvk.Kind
	}

	// Set an owner reference on the manifest for garbage collection if the mesh is deleted.
	if mesh != nil {
		if err := controllerutil.SetOwnerReference(mesh, obj, scheme); err != nil {
			logger.Error(err, "Failed SetOwnerReference", "Mesh", mesh.Name, "Namespace", obj.GetNamespace(), kind, obj.GetName())
			return err
		}
	}
	// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/controller/controllerutil/example_test.go#L35
	// TODO: Ensure no mutateFn callback is needed for updates.
	// TODO: Use controllerutil.OperationResult to update Mesh.Status with an event stream of what was created, updated, etc.
	if _, err := controllerutil.CreateOrUpdate(context.TODO(), i.client, obj, func() error { return nil }); err != nil {
		if mesh != nil {
			logger.Error(err, "Failed CreateOrUpdate", "Mesh", mesh.Name, "Namespace", obj.GetNamespace(), kind, obj.GetName())
		} else {
			logger.Error(err, "Failed CreateOrUpdate", "Namespace", obj.GetNamespace(), kind, obj.GetName())
		}
		return err
	}

	if mesh != nil {
		logger.Info("Applied", "Mesh", mesh.Name, "Namespace", obj.GetNamespace(), kind, obj.GetName())
	} else {
		logger.Info("Applied", "Namespace", obj.GetNamespace(), kind, obj.GetName())
	}

	return nil
}

func (i *Installer) applyServiceAccount(sa *corev1.ServiceAccount, mesh *v1alpha1.Mesh, scheme *runtime.Scheme) error {
	if err := i.apply(sa, mesh, scheme); err != nil {
		return err
	}

	crb := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "gm-control"}}
	if err := i.client.Get(context.TODO(), client.ObjectKeyFromObject(crb), crb); err != nil {
		logger.Error(err, "Failed Get", "ClusterRoleBinding", "gm-control")
		return err
	}

	var found bool
	for _, sub := range crb.Subjects {
		if sub.Kind == "ServiceAccount" && sub.Name == sa.Name && sub.Namespace == sa.Namespace {
			found = true
		}
	}
	if found {
		return nil
	}

	crb.Subjects = append(crb.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      sa.Name,
		Namespace: sa.Namespace,
	})
	if err := i.apply(crb, nil, scheme); err != nil {
		return err
	}

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
