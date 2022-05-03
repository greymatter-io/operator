package mesh_install

import (
	"context"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	"github.com/greymatter-io/operator/pkg/wellknown"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	appsv1 "k8s.io/api/apps/v1"
)

// ApplyMesh installs and updates Grey Matter core components and dependencies for a single mesh.
func (i *Installer) ApplyMesh(prev, mesh *v1alpha1.Mesh) {
	if prev == nil {
		logger.Info("Installing Mesh", "Name", mesh.Name)
	} else {
		logger.Info("Updating Mesh", "Name", mesh.Name)
	}

	// Create Namespace and image pull secret if this Mesh is new.
	if prev == nil {
		namespace := &v1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: mesh.Spec.InstallNamespace},
		}
		k8sapi.Apply(i.K8sClient, namespace, mesh, k8sapi.GetOrCreate)
		secret := i.imagePullSecret.DeepCopy()
		secret.Namespace = mesh.Spec.InstallNamespace
		k8sapi.Apply(i.K8sClient, secret, mesh, k8sapi.GetOrCreate)
		// TODO reverse-Chesterton's fence: I don't understand why this _wasn't_ done in the old operator
		for _, watched_ns := range mesh.Spec.WatchNamespaces {
			secret := i.imagePullSecret.DeepCopy()
			secret.Namespace = watched_ns
			k8sapi.Apply(i.K8sClient, secret, mesh, k8sapi.GetOrCreate)
		}

		// If we're applying a new mesh, pull the initial sidecar list for redis metrics ingress from the default
		// mesh in the initially-loaded CUE. This is important only when we're not already auto-applying the mesh in
		// the CUE, but doesn't hurt either way
		meshCopy := mesh.DeepCopy()
		patch := client.MergeFrom(meshCopy)
		mesh.Status.SidecarList = i.Mesh.Status.SidecarList
		err := (*i.K8sClient).Status().Patch(context.TODO(), mesh, patch)
		if err != nil {
			logger.Error(err, "error while attempting to update the status subresource of mesh in ApplyMesh", "mesh name", mesh.Name, "Status", mesh.Status)
		}
	}

	// The idea is a) one operator per mesh, and b) the sidecar template comes from unification with global OperatorCUE

	// Label existing deployments and statefulsets in this Mesh's namespaces
	deployments := &appsv1.DeploymentList{}
	(*i.K8sClient).List(context.TODO(), deployments)
	for _, deployment := range deployments.Items {
		watched := false
		for _, ns := range mesh.Spec.WatchNamespaces {
			if deployment.Namespace == ns {
				watched = true
				break
			}
		}
		if watched || deployment.Namespace == mesh.Spec.InstallNamespace {
			if deployment.Annotations == nil {
				deployment.Annotations = make(map[string]string)
			}
			deployment.Annotations[wellknown.ANNOTATION_LAST_APPLIED] = time.Now().String()
			k8sapi.Apply(i.K8sClient, &deployment, nil, k8sapi.CreateOrUpdate)
		}
	}
	statefulsets := &appsv1.StatefulSetList{}
	(*i.K8sClient).List(context.TODO(), statefulsets)
	for _, statefulset := range statefulsets.Items {
		watched := false
		for _, ns := range mesh.Spec.WatchNamespaces {
			if statefulset.Namespace == ns {
				watched = true
				break
			}
		}
		if watched || statefulset.Namespace == mesh.Spec.InstallNamespace {
			if statefulset.Annotations == nil {
				statefulset.Annotations = make(map[string]string)
			}
			statefulset.Annotations[wellknown.ANNOTATION_LAST_APPLIED] = time.Now().String()
			k8sapi.Apply(i.K8sClient, &statefulset, nil, k8sapi.CreateOrUpdate)
		}
	}

	// If we're updating an existing mesh, we need to reload the CUE before unification to avoid a situation
	// where the old concrete values conflict with the new ones
	if prev != nil {
		freshLoadOperatorCUE, _ := cuemodule.LoadAll(i.CueRoot)
		i.OperatorCUE = freshLoadOperatorCUE
	}
	// Do unification between the Mesh and K8s CUE here before extraction, and save the unified values
	err := i.OperatorCUE.UnifyWithMesh(mesh)
	if err != nil {
		logger.Error(err,
			"error while attempting to unify provided Mesh resource with Grey Matter mesh configs CUE",
			"Mesh", mesh)
		return
	}
	i.Mesh = mesh // set this mesh as THE mesh managed by the operator

	// Once that's done, we can get the Grey Matter configurator going concurrently
	go i.ConfigureMeshClient(mesh) // Applies the Grey Matter configuration once Control and Catalog are up

	// Extract 'em
	manifestObjects, err := i.OperatorCUE.ExtractCoreK8sManifests()
	if err != nil {
		logger.Error(err, "failed to extract k8s manifests")
		return
	}

	// Apply the k8s manifests we just extracted
	for _, manifest := range manifestObjects {
		logger.Info("Applying manifest object:",
			"Name", manifest.GetName(),
			"Repr", manifest)

		k8sapi.Apply(i.K8sClient, manifest, mesh, k8sapi.CreateOrUpdate)
	}

}

// RemoveMesh removes all references to a deleted Mesh custom resource.
// It does not uninstall core components and dependencies, since that is handled
// by the apiserver when the Mesh custom resource is deleted.
func (i *Installer) RemoveMesh(mesh *v1alpha1.Mesh) {
	logger.Info("Uninstalling Mesh", "Name", mesh.Name)

	go i.RemoveMeshClient()

	// Reload the starter Mesh CUE so it can be unified with a new one in the future
	freshLoadOperatorCUE, freshLoadMesh := cuemodule.LoadAll(i.CueRoot)
	i.OperatorCUE = freshLoadOperatorCUE
	i.Mesh = freshLoadMesh

	// Remove label for existing deployments and statefulsets
	deployments := &appsv1.DeploymentList{}
	(*i.K8sClient).List(context.TODO(), deployments)
	for _, deployment := range deployments.Items {
		watched := false
		for _, ns := range mesh.Spec.WatchNamespaces {
			if deployment.Namespace == ns {
				watched = true
				break
			}
		}
		if watched {
			dirty := false
			if deployment.Spec.Template.Labels == nil {
				dirty = true
				deployment.Spec.Template.Labels = make(map[string]string)
			}
			if _, ok := deployment.Spec.Template.Labels[wellknown.LABEL_CLUSTER]; ok {
				dirty = true
				delete(deployment.Spec.Template.Labels, wellknown.LABEL_CLUSTER)
			}
			if _, ok := deployment.Spec.Template.Labels[wellknown.LABEL_WORKLOAD]; ok {
				dirty = true
				delete(deployment.Spec.Template.Labels, wellknown.LABEL_WORKLOAD)
			}
			if dirty {
				k8sapi.Apply(i.K8sClient, &deployment, nil, k8sapi.CreateOrUpdate)
			}
		}
	}

	statefulsets := &appsv1.StatefulSetList{}
	(*i.K8sClient).List(context.TODO(), statefulsets)
	for _, statefulset := range statefulsets.Items {
		watched := false
		for _, ns := range mesh.Spec.WatchNamespaces {
			if statefulset.Namespace == ns {
				watched = true
				break
			}
		}
		if watched {
			dirty := false
			if statefulset.Spec.Template.Labels == nil {
				dirty = true
				statefulset.Spec.Template.Labels = make(map[string]string)
			}
			if _, ok := statefulset.Spec.Template.Labels[wellknown.LABEL_CLUSTER]; ok {
				dirty = true
				delete(statefulset.Spec.Template.Labels, wellknown.LABEL_CLUSTER)
			}
			if _, ok := statefulset.Spec.Template.Labels[wellknown.LABEL_WORKLOAD]; ok {
				dirty = true
				delete(statefulset.Spec.Template.Labels, wellknown.LABEL_WORKLOAD)
			}
			if dirty {
				k8sapi.Apply(i.K8sClient, &statefulset, nil, k8sapi.CreateOrUpdate)
			}
		}
	}

}
