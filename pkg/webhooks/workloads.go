package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type workloadDefaulter struct {
	*installer.Installer
	*cli.CLI
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

	pod.Spec.Containers = append(pod.Spec.Containers, sidecar.Container)
	pod.Spec.Volumes = append(pod.Spec.Volumes, sidecar.Volumes...)
	logger.Info("injected sidecar", "kind", "Pod", "generateName", pod.GenerateName+"*", "namespace", req.Namespace)

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

func (wd *workloadDefaulter) handleWorkload(req admission.Request) admission.Response {
	mesh := wd.WatchedBy(req.Namespace)
	if mesh == "" {
		return admission.ValidationResponse(true, "allowed")
	}

	var rawUpdate json.RawMessage
	var err error

	switch req.Kind.Kind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		if req.Operation != admissionv1.Delete {
			wd.Decode(req, deployment)
			deployment.Spec.Template = addClusterLabel(deployment.Spec.Template, req.Name)
			rawUpdate, err = json.Marshal(deployment)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to Deployment", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)
			go wd.ConfigureService(mesh, req.Name, deployment.Spec.Template.Spec.Containers)
		} else {
			wd.DecodeRaw(req.OldObject, deployment)
			go wd.RemoveService(mesh, req.Name, deployment.Spec.Template.Spec.Containers)
			return admission.ValidationResponse(true, "allowed")
		}

	case "StatefulSet":
		statefulset := &appsv1.StatefulSet{}
		if req.Operation != admissionv1.Delete {
			wd.Decode(req, statefulset)
			statefulset.Spec.Template = addClusterLabel(statefulset.Spec.Template, req.Name)
			rawUpdate, err = json.Marshal(statefulset)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to StatefulSet", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)
			go wd.ConfigureService(mesh, req.Name, statefulset.Spec.Template.Spec.Containers)
		} else {
			wd.DecodeRaw(req.OldObject, statefulset)
			go wd.RemoveService(mesh, req.Name, statefulset.Spec.Template.Spec.Containers)
			return admission.ValidationResponse(true, "allowed")
		}
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, rawUpdate)
}

func addClusterLabel(tmpl corev1.PodTemplateSpec, name string) corev1.PodTemplateSpec {
	if tmpl.Labels == nil {
		tmpl.Labels = make(map[string]string)
	}
	tmpl.Labels["greymatter.io/cluster"] = name
	return tmpl
}
