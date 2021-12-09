package installer

import (
	"context"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	"github.com/greymatter-io/operator/pkg/version"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyMesh installs and updates Grey Matter core components and dependencies for a single mesh.
func (i *Installer) ApplyMesh(prev, mesh *v1alpha1.Mesh) {
	if prev == nil {
		logger.Info("Installing Mesh", "Name", mesh.Name)
	} else {
		logger.Info("Upgrading Mesh", "Name", mesh.Name)
	}

	// Get a copy of the version specified in the Mesh CR.
	// Assume the value is valid since the CRD enumerates acceptable values for the apiserver.
	i.RLock()
	v := i.versions[mesh.Spec.ReleaseVersion].Copy()
	i.RUnlock()

	// Apply options for mutating the version copy's internal Cue value.
	options := mesh.Options(i.clusterIngressDomain)
	v.Unify(options...)

	go i.ConfigureMeshClient(mesh, options)

	// Create a Docker image pull secret and service account in this namespace if this Mesh is new.
	if prev == nil {
		secret := i.imagePullSecret.DeepCopy()
		secret.Namespace = mesh.Spec.InstallNamespace
		k8sapi.Apply(i.client, secret, mesh, k8sapi.GetOrCreate)
		applyServiceAccount(i.client, mesh, i.owner)
	}

	// Save this mesh's sidecar template to use later for sidecar injection
	i.Lock()
	i.sidecars[mesh.Name] = v.SidecarTemplate()
	i.Unlock()

	// Mark namespaces as belonging to this Mesh (and namespaces that are removed).
	watch := make(map[string]struct{})
	i.Lock()
	{
		i.namespaces[mesh.Spec.InstallNamespace] = mesh.Name
		for _, namespace := range mesh.Spec.WatchNamespaces {
			i.namespaces[namespace] = mesh.Name
			watch[namespace] = struct{}{}
			// Also inject the Docker image pull secret where sidecars will be injected.
			if namespace != mesh.Spec.InstallNamespace {
				secret := i.imagePullSecret.DeepCopy()
				secret.Name = "gm-docker-secret"
				secret.Namespace = namespace
				k8sapi.Apply(i.client, secret, mesh, k8sapi.GetOrCreate)
			}
		}
		// If the Mesh is being updated, clean up any removed watch namespaces.
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
		if _, ok := watch[deployment.Namespace]; ok || deployment.Namespace == mesh.Spec.InstallNamespace {
			if deployment.Annotations == nil {
				deployment.Annotations = make(map[string]string)
			}
			deployment.Annotations["greymatter.io/last-applied"] = time.Now().String()
			k8sapi.Apply(i.client, &deployment, nil, k8sapi.CreateOrUpdate)
		}
	}
	statefulsets := &appsv1.StatefulSetList{}
	i.client.List(context.TODO(), statefulsets)
	for _, statefulset := range statefulsets.Items {
		if _, ok := watch[statefulset.Namespace]; ok || statefulset.Namespace == mesh.Spec.InstallNamespace {
			if statefulset.Annotations == nil {
				statefulset.Annotations = make(map[string]string)
			}
			statefulset.Annotations["greymatter.io/last-applied"] = time.Now().String()
			k8sapi.Apply(i.client, &statefulset, nil, k8sapi.CreateOrUpdate)
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

		// These resources are applied with 'GetOrCreate' or 'CreateOrUpdate'.
		// We should only use 'CreateOrUpdate' when we know values may have changed
		// based on the parameters set in the Mesh CR.
		for _, configMap := range group.ConfigMaps {
			k8sapi.Apply(i.client, configMap, mesh, k8sapi.GetOrCreate)
		}
		for _, secret := range group.Secrets {
			k8sapi.Apply(i.client, secret, mesh, k8sapi.GetOrCreate)
		}
		if group.Deployment != nil {
			k8sapi.Apply(i.client, group.Deployment, mesh, k8sapi.CreateOrUpdate)
		}
		if group.StatefulSet != nil {
			k8sapi.Apply(i.client, group.StatefulSet, mesh, k8sapi.CreateOrUpdate)
		}
		if group.Service != nil {
			k8sapi.Apply(i.client, group.Service, mesh, k8sapi.GetOrCreate)
		}
		if group.Ingress != nil {
			k8sapi.Apply(i.client, group.Ingress, mesh, k8sapi.CreateOrUpdate)
		}
	}
}

func applyServiceAccount(c client.Client, mesh *v1alpha1.Mesh, crd *extv1.CustomResourceDefinition) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gm-control",
			Namespace: mesh.Spec.InstallNamespace,
		},
		AutomountServiceAccountToken: func() *bool {
			b := true
			return &b
		}(),
	}
	k8sapi.Apply(c, sa, mesh, k8sapi.GetOrCreate)

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

	k8sapi.Apply(c, crb, crd, k8sapi.CreateOrUpdate)
}

// RemoveMesh removes all references to a deleted Mesh custom resource.
// It does not uninstall core components and dependencies, since that is handled
// by the apiserver when the Mesh custom resource is deleted.
func (i *Installer) RemoveMesh(mesh *v1alpha1.Mesh) {
	logger.Info("Uninstalling Mesh", "Name", mesh.Name)

	go i.RemoveMeshClient(mesh.Name)

	watch := make(map[string]struct{})
	watch[mesh.Spec.InstallNamespace] = struct{}{}

	i.Lock()
	delete(i.namespaces, mesh.Spec.InstallNamespace)
	for _, namespace := range mesh.Spec.WatchNamespaces {
		watch[namespace] = struct{}{}
		delete(i.namespaces, namespace)
	}
	i.Unlock()

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
					delete(deployment.Spec.Template.Labels, "greymatter.io/workload")
					k8sapi.Apply(i.client, &deployment, nil, k8sapi.CreateOrUpdate)
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
					delete(statefulset.Spec.Template.Labels, "greymatter.io/workload")
					k8sapi.Apply(i.client, &statefulset, nil, k8sapi.CreateOrUpdate)
				}
			}
		}
	}
}

// WatchedBy returns the name of the mesh a namespace is a member of, or an empty string.
func (i *Installer) WatchedBy(namespace string) string {
	i.RLock()
	defer i.RUnlock()

	return i.namespaces[namespace]
}

// Sidecar returns sidecar manifests for the mesh that a namespace is a membeer of.
// It is used by the webhook package for automatic sidecar injection.
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
