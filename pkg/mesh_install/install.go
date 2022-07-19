package mesh_install

import (
	"context"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/gmapi"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	"github.com/greymatter-io/operator/pkg/wellknown"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	appsv1 "k8s.io/api/apps/v1"
)

// ApplyMesh installs and updates Grey Matter core components and dependencies for a single mesh.
func (i *Installer) ApplyMesh() {
	freshLoadOperatorCUE, mesh, err := cuemodule.LoadAll(i.CueRoot)
	if err != nil {
		logger.Error(err, "failed to load CUE during Apply")
		return
	}

	i.OperatorCUE = freshLoadOperatorCUE

	if i.Mesh == nil {
		logger.Info("Installing Mesh", "Name", mesh)
	} else {
		logger.Info("Updating Mesh", "Name", i.Mesh.Name)
	}

	// Create Namespace and image pull secret if this Mesh is new.
	if i.Mesh == nil {
		namespace := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name: mesh.Spec.InstallNamespace,
			},
		}
		k8sapi.Apply(i.K8sClient, namespace, nil, k8sapi.GetOrCreate)
		secret := i.imagePullSecret.DeepCopy()
		secret.Namespace = mesh.Spec.InstallNamespace
		k8sapi.Apply(i.K8sClient, secret, i.owner, k8sapi.GetOrCreate)
	}

	for _, watchedNS := range mesh.Spec.WatchNamespaces {
		// Create all watched namespaces, if they don't already exist
		namespace := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name: watchedNS,
			},
		}
		k8sapi.Apply(i.K8sClient, namespace, nil, k8sapi.GetOrCreate)
		// Copy the imagePullSecret into all watched namespaces
		secret := i.imagePullSecret.DeepCopy()
		secret.Namespace = watchedNS
		k8sapi.Apply(i.K8sClient, secret, i.owner, k8sapi.GetOrCreate)
	}

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
			k8sapi.Apply(i.K8sClient, &deployment, i.owner, k8sapi.CreateOrUpdate)
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
			k8sapi.Apply(i.K8sClient, &statefulset, i.owner, k8sapi.CreateOrUpdate)
		}
	}

	// Extract 'em
	manifestObjects, err := i.OperatorCUE.ExtractCoreK8sManifests()
	if err != nil {
		logger.Error(err, "failed to extract k8s manifests")
		return
	}

	// Apply the k8s manifests we just extracted
	logger.Info("Reapplying k8s manifests")
	for _, manifest := range manifestObjects {
		logger.Info("Applying manifest object:",
			"Name", manifest.GetName(),
			"Repr", manifest)

		k8sapi.Apply(i.K8sClient, manifest, i.owner, k8sapi.CreateOrUpdate)
	}

	if i.Mesh == nil {
		i.ConfigureMeshClient(mesh) // Synchronously applies the Grey Matter configuration once Control and Catalog are up
	} else {
		logger.Info("Reapplying mesh configs")
		i.EnsureClient("ApplyMesh")
		go gmapi.ApplyCoreMeshConfigs(i.Client, i.OperatorCUE)
	}

	i.Mesh = mesh // set this mesh as THE mesh managed by the operator
}
