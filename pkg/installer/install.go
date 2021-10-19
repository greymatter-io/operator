package installer

import (
	"context"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/version"

	appsv1 "k8s.io/api/apps/v1"
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
func (i *Installer) ApplyMesh(prev, mesh *v1alpha1.Mesh) {

	// Obtain the scheme used by our client
	scheme := i.client.Scheme()

	// Get a copy of the version specified in the Mesh CR.
	// Assume the value is valid since the CRD enumerates acceptable values for the apiserver.
	i.RLock()
	v := i.versions[mesh.Spec.ReleaseVersion].Copy()
	i.RUnlock()

	// Apply options for mutating the version copy's internal Cue value.
	opts := append(
		mesh.InstallOptions(),
		// Note: Each copied ImagePullSecret will always be named "gm-docker-secret"
		// even if the original secret in the gm-operator namespace has a different name.
		version.ImagePullSecretName("gm-docker-secret"),
		version.JWTSecrets,
	)
	v.Apply(opts...)

	// Create a Docker image pull secret and service account in this namespace if this Mesh is new.
	if prev == nil {
		secret := i.imagePullSecret.DeepCopy()
		secret.Name = "gm-docker-secret"
		secret.Namespace = mesh.Namespace
		apply(i.client, secret, mesh, scheme)
		// If this is the first mesh, setup RBAC for control plane service accounts to view pods.
		if len(i.sidecars) == 0 {
			applyClusterRBAC(i.client, scheme)
		}
		applyServiceAccount(i.client, mesh, scheme)
	}

	// Save this mesh's sidecar template to use later for sidecar injection
	i.Lock()
	i.sidecars[mesh.Name] = v.SidecarTemplate()
	i.Unlock()

	// Mark namespaces as belonging to this Mesh (and namespaces that are removed).
	watch := make(map[string]struct{})
	i.Lock()
	{
		i.namespaces[mesh.Namespace] = mesh.Name
		for _, namespace := range mesh.Spec.WatchNamespaces {
			i.namespaces[namespace] = mesh.Name
			watch[namespace] = struct{}{}
			// Also inject the Docker image pull secret where sidecars will be injected.
			if namespace != mesh.Namespace {
				secret := i.imagePullSecret.DeepCopy()
				secret.Name = "gm-docker-secret"
				secret.Namespace = namespace
				apply(i.client, secret, mesh, scheme)
			}
		}
		// If the Mesh is being updated, note any removed watch namespaces.
		if prev != nil {
			for _, namespace := range prev.Spec.WatchNamespaces {
				if _, ok := watch[namespace]; !ok {
					delete(i.namespaces, namespace)
					// TODO: Remove the Docker image pull secret when a watch namespace is removed.
				}
			}
		}
	}
	i.Unlock()

	// Label existing deployments and statefulsets in this Mesh's namespaces
	deployments := &appsv1.DeploymentList{}
	i.client.List(context.TODO(), deployments)
	for _, deployment := range deployments.Items {
		if _, ok := watch[deployment.Namespace]; ok || deployment.Namespace == mesh.Namespace {
			if _, ok := deployment.Spec.Template.Labels["greymatter.io/cluster"]; !ok {
				if deployment.Spec.Template.Labels == nil {
					deployment.Spec.Template.Labels = make(map[string]string)
				}
				deployment.Spec.Template.Labels["greymatter.io/cluster"] = deployment.Name
				apply(i.client, &deployment, nil, scheme)
			}
		}
	}
	statefulsets := &appsv1.StatefulSetList{}
	i.client.List(context.TODO(), statefulsets)
	for _, statefulset := range statefulsets.Items {
		if _, ok := watch[statefulset.Namespace]; ok || statefulset.Namespace == mesh.Namespace {
			if _, ok := statefulset.Spec.Template.Labels["greymatter.io/cluster"]; !ok {
				if statefulset.Spec.Template.Labels == nil {
					statefulset.Spec.Template.Labels = make(map[string]string)
				}
				statefulset.Spec.Template.Labels["greymatter.io/cluster"] = statefulset.Name
				apply(i.client, &statefulset, nil, scheme)
			}
		}
	}

MANIFEST_LOOP:
	for _, group := range v.Manifests() {
		// If an external Redis server is configured, don't install an internal Redis.
		if group.StatefulSet != nil &&
			group.StatefulSet.Name == "gm-redis" &&
			mesh.Spec.ExternalRedis != nil &&
			mesh.Spec.ExternalRedis.URL != "" {
			continue MANIFEST_LOOP
		}

		for _, configMap := range group.ConfigMaps {
			apply(i.client, configMap, mesh, scheme)
		}
		for _, secret := range group.Secrets {
			apply(i.client, secret, mesh, scheme)
		}
		if group.Deployment != nil {
			apply(i.client, group.Deployment, mesh, scheme)
		}
		if group.StatefulSet != nil {
			apply(i.client, group.StatefulSet, mesh, scheme)
		}
		if group.Service != nil {
			apply(i.client, group.Service, mesh, scheme)
		}
		if group.Ingress != nil {
			apply(i.client, group.Ingress, mesh, scheme)
		}
	}
}

func apply(c client.Client, obj, owner client.Object, scheme *runtime.Scheme) {
	var kind string
	if gvk, err := apiutil.GVKForObject(obj.(runtime.Object), scheme); err != nil {
		kind = "Object"
	} else {
		kind = gvk.Kind
	}

	// Set an owner reference on the manifest for garbage collection if the mesh is deleted.
	if owner != nil {
		if err := controllerutil.SetOwnerReference(owner, obj, scheme); err != nil {
			logger.Error(err, "SetOwnerReference", "result", "failed", "Owner", owner.GetName(), "Namespace", obj.GetNamespace(), kind, obj.GetName())
			return
		}
	}

	action, result, err := createOrUpdate(context.TODO(), c, obj)
	if err != nil {
		if owner != nil {
			logger.Error(err, action, "result", "failed", "Owner", owner.GetName(), "Namespace", obj.GetNamespace(), kind, obj.GetName())
		} else {
			logger.Error(err, action, "result", "failed", "Namespace", obj.GetNamespace(), kind, obj.GetName())
		}
		return
	}

	if owner != nil {
		logger.Info(action, "result", result, "Owner", owner.GetName(), "Namespace", obj.GetNamespace(), kind, obj.GetName())
	} else {
		logger.Info(action, "result", result, "Namespace", obj.GetNamespace(), kind, obj.GetName())
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

func applyClusterRBAC(c client.Client, scheme *runtime.Scheme) {
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
	apply(c, cr, nil, scheme)

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-control"},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "gm-control",
		},
		Subjects: []rbacv1.Subject{},
	}
	apply(c, crb, cr, scheme)
}

func applyServiceAccount(c client.Client, mesh *v1alpha1.Mesh, scheme *runtime.Scheme) {
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
	apply(c, sa, mesh, scheme)

	crb := &rbacv1.ClusterRoleBinding{}
	if err := c.Get(context.TODO(), client.ObjectKey{Name: "gm-control"}, crb); err != nil {
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

	apply(c, crb, nil, scheme)
}

// Cleanup if a Mesh CR is deleted.
func (i *Installer) RemoveMesh(mesh *v1alpha1.Mesh) {
	watch := make(map[string]struct{})
	watch[mesh.Namespace] = struct{}{}

	i.Lock()
	delete(i.namespaces, mesh.Namespace)
	for _, namespace := range mesh.Spec.WatchNamespaces {
		watch[namespace] = struct{}{}
		delete(i.namespaces, namespace)
	}
	i.Unlock()

	scheme := i.client.Scheme()

	// Remove label for existing deployments and statefulsets
	deployments := &appsv1.DeploymentList{}
	i.client.List(context.TODO(), deployments)
	for _, deployment := range deployments.Items {
		if _, ok := watch[deployment.Namespace]; ok {
			if deployment.ObjectMeta.Labels["app.kubernetes.io/created-by"] != "gm-operator" {
				if deployment.Spec.Template.Labels == nil {
					deployment.Spec.Template.Labels = make(map[string]string)
				}
				if _, ok := deployment.Spec.Template.Labels["greymatter.io/cluster"]; ok {
					delete(deployment.Spec.Template.Labels, "greymatter.io/cluster")
					apply(i.client, &deployment, nil, scheme)
				}
			}
		}
	}

	statefulsets := &appsv1.StatefulSetList{}
	i.client.List(context.TODO(), statefulsets)
	for _, statefulset := range statefulsets.Items {
		if _, ok := watch[statefulset.Namespace]; ok {
			if statefulset.ObjectMeta.Labels["app.kubernetes.io/created-by"] != "gm-operator" {
				if statefulset.Spec.Template.Labels == nil {
					statefulset.Spec.Template.Labels = make(map[string]string)
				}
				if _, ok := statefulset.Spec.Template.Labels["greymatter.io/cluster"]; ok {
					delete(statefulset.Spec.Template.Labels, "greymatter.io/cluster")
					apply(i.client, &statefulset, nil, scheme)
				}
			}
		}
	}
}

func (i *Installer) WatchedBy(namespace string) string {
	i.RLock()
	defer i.RUnlock()

	return i.namespaces[namespace]
}

func (i *Installer) Sidecar(namespace, xdsCluster string) (version.Sidecar, bool) {
	i.RLock()
	defer i.RUnlock()

	meshName, ok := i.namespaces[namespace]
	if !ok {
		return version.Sidecar{}, false
	}

	sidecar, ok := i.sidecars[meshName]
	if !ok {
		return version.Sidecar{}, false
	}

	return sidecar(xdsCluster), true
}
