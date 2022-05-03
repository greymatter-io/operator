package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/greymatter-io/operator/pkg/gmapi"
	"github.com/greymatter-io/operator/pkg/mesh_install"
	"github.com/greymatter-io/operator/pkg/wellknown"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type workloadDefaulter struct {
	*mesh_install.Installer
	*gmapi.CLI
	*admission.Decoder
}

// InjectDecoder implements admission.DecoderInjector.
// A decoder will be automatically injected for decoding deployments, statefulsets, and pods.
func (wd *workloadDefaulter) InjectDecoder(d *admission.Decoder) error {
	wd.Decoder = d
	return nil
}

// Handle implements admission.Handler.
// It will be invoked when creating, updating, or deleting deployments and statefulsets,
// or when creating or updating pods.
func (wd *workloadDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Kind.Kind == "Pod" {
		return wd.handlePod(req)
	}
	return wd.handleWorkload(req)
}

func (wd *workloadDefaulter) handlePod(req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		return admission.ValidationResponse(true, "allowed")
	}

	pod := &corev1.Pod{}
	if err := wd.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Check for a cluster label; if not found, this pod does not belong to a Mesh.
	clusterLabel, ok := pod.Labels[wellknown.LABEL_CLUSTER]
	if !ok {
		return admission.ValidationResponse(true, "allowed")
	}
	// Check for an existing proxy port; if found, this pod already has a sidecar.
	for _, container := range pod.Spec.Containers {
		for _, p := range container.Ports {
			if p.Name == "proxy" {
				return admission.ValidationResponse(true, "allowed")
			}
		}
	}

	container, volumes, err := wd.OperatorCUE.UnifyAndExtractSidecar(clusterLabel)
	if err != nil {
		return admission.ValidationResponse(true, "allowed")
	}

	pod.Spec.Containers = append(pod.Spec.Containers, container)
	pod.Spec.Volumes = append(pod.Spec.Volumes, volumes...)
	logger.Info("injected sidecar", "name", clusterLabel, "kind", "Pod", "generateName", pod.GenerateName+"*", "namespace", req.Namespace)

	// Inject a reference to the image pull secret
	var hasImagePullSecret bool
	for _, secret := range pod.Spec.ImagePullSecrets {
		if secret.Name == "gm-docker-secret" {
			hasImagePullSecret = true
		}
	}
	if !hasImagePullSecret {
		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: "gm-docker-secret"})
	}

	rawUpdate, err := json.Marshal(pod)
	if err != nil {
		logger.Error(err, "Failed to decode corev1.Pod", "Name", req.Name, "Namespace", req.Namespace)
		return admission.ValidationResponse(false, "failed to decode")
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, rawUpdate)
}

// TODO: Modification should happen using a CUE package.
func (wd *workloadDefaulter) handleWorkload(req admission.Request) admission.Response {
	meshName := wd.Mesh.Name                           // wd.WatchedBy(req.Namespace)
	if meshName == "" || wd.Installer.Mesh.UID == "" { // If the mesh isn't actually applied, don't assist deployments
		return admission.ValidationResponse(true, "allowed")
	}

	var rawUpdate json.RawMessage
	var err error

	switch req.Kind.Kind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		if req.Operation != admissionv1.Delete { // if new or updated Deployment
			wd.Decode(req, deployment)
			if deployment.Annotations == nil {
				deployment.Annotations = make(map[string]string)
			}
			deployment.Annotations[wellknown.ANNOTATION_LAST_APPLIED] = time.Now().String()
			deployment.Spec.Template = addClusterLabels(deployment.Spec.Template, meshName, req.Name)
			rawUpdate, err = json.Marshal(deployment)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to Deployment", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)

			annotations := deployment.ObjectMeta.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.addToMeshSidecarList(req.Name)
					wd.ConfigureSidecar(wd.OperatorCUE, req.Name, deployment.ObjectMeta)
				}()
			}

		} else { // if this Deployment is being deleted...
			wd.DecodeRaw(req.OldObject, deployment)

			annotations := deployment.ObjectMeta.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.removeFromMeshSidecarList(req.Name)
					wd.UnconfigureSidecar(wd.OperatorCUE, req.Name, deployment.ObjectMeta)
				}()
			}
			return admission.ValidationResponse(true, "allowed")
		}

	case "StatefulSet":
		statefulset := &appsv1.StatefulSet{}
		if req.Operation != admissionv1.Delete { // if new or updated StatefulSet
			wd.Decode(req, statefulset)
			if statefulset.Annotations == nil {
				statefulset.Annotations = make(map[string]string)
			}
			statefulset.Annotations[wellknown.ANNOTATION_LAST_APPLIED] = time.Now().String()
			statefulset.Spec.Template = addClusterLabels(statefulset.Spec.Template, meshName, req.Name)
			rawUpdate, err = json.Marshal(statefulset)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to StatefulSet", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)

			annotations := statefulset.ObjectMeta.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.addToMeshSidecarList(req.Name)
					wd.ConfigureSidecar(wd.OperatorCUE, req.Name, statefulset.ObjectMeta)
				}()
			}

		} else { // if this StatefulSet is being deleted...
			wd.DecodeRaw(req.OldObject, statefulset)

			annotations := statefulset.ObjectMeta.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.removeFromMeshSidecarList(req.Name)
					wd.UnconfigureSidecar(wd.OperatorCUE, req.Name, statefulset.ObjectMeta)
				}()
			}
			return admission.ValidationResponse(true, "allowed")
		}
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, rawUpdate)
}

func (wd *workloadDefaulter) addToMeshSidecarList(name string) {
	// apply the new sidecar list to the Mesh status
	logger.Info("Mesh SidecarList at time of addition", "list", wd.Mesh.Status.SidecarList, "addition", name) // DEBUG
	// TODO don't append it if it's already there
	for _, sidecarName := range wd.Mesh.Status.SidecarList {
		if sidecarName == name {
			return
		}
	}
	// save the sidecar list to the deployed CR
	meshCopy := wd.Mesh.DeepCopy()
	patch := client.MergeFrom(meshCopy)
	wd.Mesh.Status.SidecarList = append(wd.Mesh.Status.SidecarList, name)
	err := (*wd.K8sClient).Status().Patch(context.TODO(), wd.Mesh, patch)
	if err != nil {
		logger.Error(err, "error while attempting to update the status subresource of mesh", "mesh name", wd.Mesh.Name, "Status", wd.Mesh.Status)
	}

	// Update the mesh inside the OperatorCUE with the new sidecar_list
	freshLoadOperatorCUE, _ := cuemodule.LoadAll(wd.CueRoot)
	wd.OperatorCUE = freshLoadOperatorCUE
	wd.OperatorCUE.UnifyWithMesh(wd.Mesh)
}

func (wd *workloadDefaulter) removeFromMeshSidecarList(name string) {
	var filtered []string
	for _, sidecarName := range wd.Mesh.Status.SidecarList {
		if sidecarName != name {
			filtered = append(filtered, sidecarName)
		}
	}
	meshCopy := wd.Mesh.DeepCopy()
	patch := client.MergeFrom(meshCopy)
	wd.Mesh.Status.SidecarList = filtered
	err := (*wd.K8sClient).Status().Patch(context.TODO(), wd.Mesh, patch)
	if err != nil {
		logger.Error(err, "error while attempting to update the status subresource of mesh", "mesh name", wd.Mesh.Name, "Status", wd.Mesh.Status)
	}

	// Update the mesh inside the OperatorCUE with the new sidecar_list
	freshLoadOperatorCUE, _ := cuemodule.LoadAll(wd.CueRoot)
	wd.OperatorCUE = freshLoadOperatorCUE
	wd.OperatorCUE.UnifyWithMesh(wd.Mesh)
}

func addClusterLabels(tmpl corev1.PodTemplateSpec, meshName, clusterName string) corev1.PodTemplateSpec {
	if tmpl.Labels == nil {
		tmpl.Labels = make(map[string]string)
	}
	// For service discovery
	tmpl.Labels[wellknown.LABEL_CLUSTER] = clusterName
	// For Spire identification
	tmpl.Labels[wellknown.LABEL_WORKLOAD] = fmt.Sprintf("%s.%s", meshName, clusterName)
	return tmpl
}
