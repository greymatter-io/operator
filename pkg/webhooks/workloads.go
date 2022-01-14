package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

	// If a Pod is being deleted, or the Pod's namespace does not belong to a Mesh, skip.
	if req.Operation == admissionv1.Delete || wd.WatchedBy(req.Namespace) == "" {
		return admission.ValidationResponse(true, "allowed")
	}

	// Decode the inlined Pod object in the request.
	// This should never error since the Kubelet has already validated the object as a Pod.
	pod := &corev1.Pod{}
	if err := wd.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Check the Pod for a cluster label. If not found, this pod does not belong to a Mesh.
	xdsCluster, ok := pod.Labels["greymatter.io/cluster"]
	if !ok {
		return admission.ValidationResponse(true, "allowed")
	}

	// Get the Mesh-specific sidecar container manifest and any volumes which should be injected in the Pod.
	sidecar, ok := wd.Sidecar(req.Namespace, xdsCluster)
	if !ok {
		logger.Error(fmt.Errorf("failed to inject sidecar"), "failed to compile container config", "Pod", pod.GenerateName+"*")
		return admission.ValidationResponse(true, "allowed")
	}

	// Determine whether the incoming Pod spec already has a container with an image that has "gm-proxy"  in it.
	// If it does, it has a sidecar and we want to identify its index in the slice of containers.
	sidecarContainerIdx := -1
	for i, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "gm-proxy") {
			sidecarContainerIdx = i
		}
	}

	// If a sidecar container already exists in the Pod spec, populate its config. Otherwise, inject a sidecar container.
	// Note that this overwrite means users cannot manually configure the sidecar container in a Pod spec/template.
	// We can change this implementation later if we want some values to be configurable (e.g. env vars, volumeMount).
	if sidecarContainerIdx > -1 {
		pod.Spec.Containers[sidecarContainerIdx] = sidecar.Container
		logger.Info("configured sidecar", "kind", "Pod", "generateName", pod.GenerateName+"*", "namespace", req.Namespace)
	} else {
		pod.Spec.Containers = append(pod.Spec.Containers, sidecar.Container)
		logger.Info("injected sidecar", "kind", "Pod", "generateName", pod.GenerateName+"*", "namespace", req.Namespace)
	}

	// Inject sidecar volumes into the Pod spec.
	// In order to avoid duplicate volumes, make a set of volume names.
	// Any user-defined volumes will be added to the standard list of sidecar volumes.
	// Note that users can also override volumes that share the same name if they choose.
	volumes := make(map[string]struct{})
	for _, vol := range pod.Spec.Volumes {
		volumes[vol.Name] = struct{}{}
	}
	for _, vol := range sidecar.Volumes {
		if _, ok := volumes[vol.Name]; !ok {
			pod.Spec.Volumes = append(pod.Spec.Volumes, vol)
		}
	}

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
			if deployment.Annotations == nil {
				deployment.Annotations = make(map[string]string)
			}
			deployment.Annotations["greymatter.io/last-applied"] = time.Now().String()
			deployment.Spec.Template = addClusterLabels(deployment.Spec.Template, mesh, req.Name)
			rawUpdate, err = json.Marshal(deployment)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to Deployment", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added/found cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)
			go wd.ConfigureService(mesh, req.Name, deployment.Annotations, deployment.Spec.Template.Spec.Containers)
		} else {
			wd.DecodeRaw(req.OldObject, deployment)
			go wd.RemoveService(mesh, req.Name, deployment.Annotations, deployment.Spec.Template.Spec.Containers)
			return admission.ValidationResponse(true, "allowed")
		}

	case "StatefulSet":
		statefulset := &appsv1.StatefulSet{}
		if req.Operation != admissionv1.Delete {
			wd.Decode(req, statefulset)
			if statefulset.Annotations == nil {
				statefulset.Annotations = make(map[string]string)
			}
			statefulset.Annotations["greymatter.io/last-applied"] = time.Now().String()
			statefulset.Spec.Template = addClusterLabels(statefulset.Spec.Template, mesh, req.Name)
			rawUpdate, err = json.Marshal(statefulset)
			if err != nil {
				logger.Error(err, "Failed to add cluster label to StatefulSet", "Name", req.Name, "Namespace", req.Namespace)
				return admission.ValidationResponse(false, "failed to add cluster label")
			}
			logger.Info("added/found cluster label", "kind", req.Kind.Kind, "name", req.Name, "namespace", req.Namespace)
			go wd.ConfigureService(mesh, req.Name, statefulset.Annotations, statefulset.Spec.Template.Spec.Containers)
		} else {
			wd.DecodeRaw(req.OldObject, statefulset)
			go wd.RemoveService(mesh, req.Name, statefulset.Annotations, statefulset.Spec.Template.Spec.Containers)
			return admission.ValidationResponse(true, "allowed")
		}
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, rawUpdate)
}

func addClusterLabels(tmpl corev1.PodTemplateSpec, mesh, name string) corev1.PodTemplateSpec {
	if tmpl.Labels == nil {
		tmpl.Labels = make(map[string]string)
	}
	tmpl.Labels["greymatter.io/cluster"] = name
	tmpl.Labels["greymatter.io/workload"] = fmt.Sprintf("%s.%s", mesh, name)
	return tmpl
}
