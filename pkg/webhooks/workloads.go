package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/greymatter-io/operator/pkg/installer"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type workloadDefaulter struct {
	*installer.Installer
	*admission.Decoder
}

// Implements admission.DecoderInjector.
// A decoder will be automatically injected.
func (wd *workloadDefaulter) InjectDecoder(d *admission.Decoder) error {
	wd.Decoder = d
	return nil
}

func (wd *workloadDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Kind.Kind == "Pod" {
		return wd.handlePod(req)
	}
	return wd.handleWorkload(req)
}

func (wd *workloadDefaulter) handleWorkload(req admission.Request) admission.Response {

	// TEMP: Later, handle deletes by removing mesh configs
	if req.Operation == admissionv1.Delete {
		return admission.ValidationResponse(true, "allowed")
	}

	if !wd.IsMeshMember(req.Namespace) {
		return admission.ValidationResponse(true, "allowed")
	}

	logger.Info("add cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)

	var rawUpdate json.RawMessage
	var err error

	switch req.Kind.Kind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		if req.Operation != admissionv1.Delete {
			wd.DecodeRaw(req.Object, deployment)
			if deployment.Spec.Template.Labels == nil {
				deployment.Spec.Template.Labels = make(map[string]string)
			}
			if _, ok := deployment.Spec.Template.Labels["greymatter.io/cluster"]; ok {
				return admission.ValidationResponse(true, "allowed")
			}
			deployment.Spec.Template.Labels["greymatter.io/cluster"] = req.Name
			// TODO: Add mesh configs
		} // else {}
		// TODO: Handle deletes (remove mesh configs)
		rawUpdate, err = json.Marshal(deployment)
		if err != nil {
			logger.Error(err, "Failed to decode appsv1.Deployment", "Name", req.Name, "Namespace", req.Namespace)
			return admission.ValidationResponse(false, "failed to decode")
		}

	case "StatefulSet":
		statefulset := &appsv1.StatefulSet{}
		if req.Operation != admissionv1.Delete {
			wd.DecodeRaw(req.Object, statefulset)
			if statefulset.Spec.Template.Labels == nil {
				statefulset.Spec.Template.Labels = make(map[string]string)
			}
			if _, ok := statefulset.Spec.Template.Labels["greymatter.io/cluster"]; ok {
				return admission.ValidationResponse(true, "allowed")
			}
			statefulset.Spec.Template.Labels["greymatter.io/cluster"] = req.Name
			// TODO: Add mesh configs
		} // else {}
		// TODO: Handle deletes (remove mesh configs)
		rawUpdate, err = json.Marshal(statefulset)
		if err != nil {
			logger.Error(err, "Failed to decode appsv1.StatefulSet", "Name", req.Name, "Namespace", req.Namespace)
			return admission.ValidationResponse(false, "failed to decode")
		}
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, rawUpdate)
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
	xdsCluster, ok := pod.Labels["greymatter.io/cluster"]
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

	// Get the sidecar container and any volumes to add to the Pod.
	sidecar, ok := wd.Sidecar(req.Namespace, xdsCluster)
	if !ok {
		return admission.ValidationResponse(true, "allowed")
	}

	logger.Info("inject sidecar", "kind", "Pod", "generateName", pod.GenerateName+"*", "namespace", req.Namespace)
	pod.Spec.Containers = append(pod.Spec.Containers, sidecar.Container)
	pod.Spec.Volumes = append(pod.Spec.Volumes, sidecar.Volumes...)

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
