package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
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

	// If there's no mesh, don't assist deployment
	if wd.Mesh == nil || wd.Mesh.Name == "" {
		return admission.ValidationResponse(true, "allowed")
	}
	// If the pod isn't in a watched namespace, don't assist deployment
	watched := false
	for _, ns := range wd.Mesh.Spec.WatchNamespaces {
		if req.Namespace == ns {
			watched = true
			break
		}
	}
	if !watched {
		return admission.ValidationResponse(true, "allowed")
	}

	pod := &corev1.Pod{}
	if err := wd.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	annotations := pod.Annotations
	if injectSidecarTo, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]; !injectSidecar || injectSidecarTo == "" {
		logger.Info("No inject-sidecar-to annotation, skipping", "name", req.Name, "annotations", annotations)
		return admission.ValidationResponse(true, "allowed")
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
	// If there's no mesh, don't assist deployment
	if wd.Mesh == nil || wd.Mesh.Name == "" {
		return admission.ValidationResponse(true, "allowed")
	}
	meshName := wd.Mesh.Name // wd.WatchedBy(req.Namespace)

	// If the workload isn't in a watched namespace or the install namespace, don't assist deployment
	watched := false
	for _, ns := range wd.Mesh.Spec.WatchNamespaces {
		if req.Namespace == ns {
			watched = true
			break
		}
	}
	if req.Namespace == wd.Mesh.Spec.InstallNamespace {
		watched = true
	}
	if !watched {
		return admission.ValidationResponse(true, "allowed")
	}

	var rawUpdate json.RawMessage
	var err error

	switch req.Kind.Kind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		if req.Operation != admissionv1.Delete { // if new or updated Deployment
			wd.Decode(req, deployment)
			if deployment.Spec.Template.Annotations == nil {
				deployment.Spec.Template.Annotations = make(map[string]string)
			}
			deployment.Spec.Template.Annotations[wellknown.ANNOTATION_LAST_APPLIED] = time.Now().String()
			deployment.Spec.Template = mesh_install.AddClusterLabels(deployment.Spec.Template, meshName, req.Name)
			rawUpdate, err = json.Marshal(deployment)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to Deployment", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)

			annotations := deployment.Spec.Template.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.ConfigureSidecar(wd.OperatorCUE, req.Name, annotations)
				}()
			}

		} else { // if this Deployment is being deleted...
			wd.DecodeRaw(req.OldObject, deployment)

			annotations := deployment.Spec.Template.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.UnconfigureSidecar(wd.OperatorCUE, req.Name, annotations)
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
			statefulset.Spec.Template = mesh_install.AddClusterLabels(statefulset.Spec.Template, meshName, req.Name)
			rawUpdate, err = json.Marshal(statefulset)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to StatefulSet", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)

			annotations := statefulset.Spec.Template.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.ConfigureSidecar(wd.OperatorCUE, req.Name, annotations)
				}()
			}

		} else { // if this StatefulSet is being deleted...
			wd.DecodeRaw(req.OldObject, statefulset)

			annotations := statefulset.Spec.Template.Annotations
			_, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
			if injectSidecar {
				go func() {
					wd.UnconfigureSidecar(wd.OperatorCUE, req.Name, annotations)
				}()
			}
			return admission.ValidationResponse(true, "allowed")
		}
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, rawUpdate)
}
