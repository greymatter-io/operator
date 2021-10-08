package installer

import (
	"context"
	"fmt"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/version"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

	// Obtain the scheme used by our client for
	scheme := i.client.Scheme()

	// Create a Docker image pull secret and service account in this namespace if this Mesh is new.
	if init {
		if mesh.Namespace != "gm-operator" {
			secret := i.imagePullSecret.DeepCopy()
			secret.Namespace = mesh.Namespace
			i.apply(secret, mesh, scheme)
		}
		// If this is the first mesh, setup RBAC for control plane service accounts to view pods.
		if len(i.sidecars) == 0 {
			i.applyClusterRBAC()
		}
		i.applyServiceAccount(mesh, scheme)
	}

	// Save this mesh's Proxy InstallConfig to use later for sidecar injection
	// TODO: Inject sidecars into existing deployments and statefulsets!
	i.sidecars[mesh.Name] = v.Sidecar()

	for _, namespace := range mesh.Spec.WatchNamespaces {
		if namespace != mesh.Namespace {
			secret := i.imagePullSecret.DeepCopy()
			secret.Namespace = namespace
			i.apply(secret, mesh, scheme)
		}
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

func (i *Installer) apply(obj, owner client.Object, scheme *runtime.Scheme) {
	var kind string
	if gvk, err := apiutil.GVKForObject(obj.(runtime.Object), scheme); err != nil {
		kind = "Object"
	} else {
		kind = gvk.Kind
	}

	// Set an owner reference on the manifest for garbage collection if the mesh is deleted.
	if owner != nil {
		if err := controllerutil.SetOwnerReference(owner, obj, scheme); err != nil {
			logger.Error(err, "Failed SetOwnerReference", "Owner", owner.GetName(), "Namespace", obj.GetNamespace(), kind, obj.GetName())
			return
		}
	}

	action, result, err := createOrUpdate(context.TODO(), i.client, obj)
	if err != nil {
		if owner != nil {
			logger.Error(err, fmt.Sprintf("Failed %s", action), "Owner", owner.GetName(), "Namespace", obj.GetNamespace(), kind, obj.GetName())
		} else {
			logger.Error(err, fmt.Sprintf("Failed %s", action), "Namespace", obj.GetNamespace(), kind, obj.GetName())
		}
		return
	}

	if owner != nil {
		logger.Info(action, "Result", result, "Mesh", owner.GetName(), "Namespace", obj.GetNamespace(), kind, obj.GetName())
	} else {
		logger.Info(action, "Result", result, "Namespace", obj.GetNamespace(), kind, obj.GetName())
	}
}

func createOrUpdate(ctx context.Context, c client.Client, obj client.Object) (string, string, error) {
	key := client.ObjectKeyFromObject(obj)

	// Make a pointer copy of the object so that our actual object is not modified by client.Get.
	// This way, the object passed into client.Update still has our desired state.
	existing := obj.DeepCopyObject()
	if err := c.Get(ctx, key, existing.(client.Object)); err != nil {
		if !errors.IsNotFound(err) {
			return "create/update", "fail", err
		}
		if err := c.Create(ctx, obj); err != nil {
			return "create", "fail", err
		}
		return "create", "success", nil
	}

	if err := c.Update(ctx, obj); err != nil {
		return "update", "fail", err
	}

	return "update", "success", nil
}

func (i *Installer) applyClusterRBAC() {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-control"},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
	i.apply(cr, nil, i.client.Scheme())

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-control"},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "gm-control",
		},
		Subjects: []rbacv1.Subject{},
	}
	i.apply(crb, cr, i.client.Scheme())
}

func (i *Installer) applyServiceAccount(mesh *v1alpha1.Mesh, scheme *runtime.Scheme) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gm-control",
			Namespace: mesh.Namespace,
		},
		AutomountServiceAccountToken: func() *bool {
			b := true
			return &b
		}(),
	}
	i.apply(sa, mesh, scheme)

	crb := &rbacv1.ClusterRoleBinding{}
	if err := i.client.Get(context.TODO(), client.ObjectKey{Name: "gm-control"}, crb); err != nil {
		logger.Error(err, "Failed Get", "ClusterRoleBinding", "gm-control")
		return
	}

	for _, subject := range crb.Subjects {
		if subject.Kind == "ServiceAccount" &&
			subject.Name == sa.Name &&
			subject.Namespace == sa.Namespace {
			return
		}
	}

	crb.Subjects = append(crb.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      sa.Name,
		Namespace: sa.Namespace,
	})

	i.apply(crb, nil, scheme)
}

// Removes all resources created for installing Grey Matter core components of a mesh.
// Also removes the mesh from labels of resources (removing the labels if no more meshes are listed)
// i.e. namespaces, deployments, statefulsets, and pods (via pod templates).
func (i *Installer) RemoveMesh(name string) {
}

// Given a slice of corev1.Containers, injects a sidecar to enable traffic for each mesh specified.
func (i *Installer) InjectSidecar(containers []corev1.Container, xdsCluster, mesh string) []corev1.Container {
	return containers
}
