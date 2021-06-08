package reconcilers

import (
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment struct {
	GmService gmcore.Service
	ObjectKey types.NamespacedName
}

func (d Deployment) Kind() string {
	return "Deployment"
}

func (d Deployment) Key() types.NamespacedName {
	return d.ObjectKey
}

func (d Deployment) Object() client.Object {
	return &appsv1.Deployment{}
}

func (d Deployment) Build(mesh *installv1.Mesh) client.Object {
	configs := gmcore.Configs(mesh.Spec.Version)
	svc := d.GmService

	matchLabels := map[string]string{
		"greymatter.io/control": d.ObjectKey.Name,
	}

	podLabels := map[string]string{
		"greymatter.io/control":         d.ObjectKey.Name,
		"greymatter.io/component":       configs[svc].Component,
		"greymatter.io/service-version": configs[svc].ImageTag,
	}
	if svc != gmcore.Control && d.ObjectKey.Name != "edge" {
		podLabels["greymatter.io/sidecar-version"] = configs[gmcore.Proxy].ImageTag
	}

	objectLabels := map[string]string{
		"app.kubernetes.io/name":       d.ObjectKey.Name,
		"app.kubernetes.io/version":    configs[svc].ImageTag,
		"app.kubernetes.io/part-of":    "greymatter",
		"app.kubernetes.io/managed-by": "gm-operator",
		"app.kubernetes.io/created-by": "gm-operator",
	}
	for k, v := range podLabels {
		objectLabels[k] = v
	}

	envsMap := configs[svc].MkEnvsMap(mesh, d.ObjectKey.Name)
	var envs []corev1.EnvVar
	for k, v := range envsMap {
		envs = append(envs, corev1.EnvVar{Name: k, Value: v})
	}

	svcContainer := corev1.Container{
		Name:            "service",
		Image:           fmt.Sprintf("docker.greymatter.io/release/gm-%s:%s", svc, configs[svc].ImageTag),
		Env:             envs,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports:           configs[svc].ContainerPorts,
	}
	if configs[svc].Resources != nil {
		svcContainer.Resources = *configs[svc].Resources
	}

	var containers []corev1.Container

	if d.ObjectKey.Name != "edge" {
		containers = append(containers, svcContainer)
	}

	if svc != gmcore.Control {
		proxyEnvsMap := configs[gmcore.Proxy].MkEnvsMap(mesh, d.ObjectKey.Name)
		var proxyEnvs []corev1.EnvVar
		for k, v := range proxyEnvsMap {
			proxyEnvs = append(proxyEnvs, corev1.EnvVar{Name: k, Value: v})
		}
		proxyContainer := corev1.Container{
			Name:            "sidecar",
			Image:           fmt.Sprintf("docker.greymatter.io/release/gm-proxy:%s", configs[gmcore.Proxy].ImageTag),
			Env:             proxyEnvs,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports:           configs[gmcore.Proxy].ContainerPorts,
			Resources:       *configs[gmcore.Proxy].Resources,
		}
		if d.ObjectKey.Name == "edge" {
			proxyContainer.Name = "edge"
		}
		containers = append(containers, proxyContainer)
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.ObjectKey.Name,
			Namespace: mesh.Namespace,
			Labels:    objectLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: matchLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: mesh.Spec.ImagePullSecret}},
					Containers:       containers,
				},
			},
		},
	}

	if svc == gmcore.Control {
		deployment.Spec.Template.Spec.ServiceAccountName = "control-pods"
	}

	return deployment
}

func (d Deployment) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	configs := gmcore.Configs(mesh.Spec.Version)
	svc := d.GmService

	labels := obj.GetLabels()
	if lbl := labels["greymatter.io/service-version"]; lbl != configs[svc].ImageTag {
		return false, nil
	}
	if lbl := labels["greymatter.io/sidecar-version"]; svc != gmcore.Control &&
		d.ObjectKey.Name != "edge" &&
		lbl != configs[gmcore.Proxy].ImageTag {
		return false, nil
	}

	return true, nil
}

func (d Deployment) Mutate(mesh *installv1.Mesh, obj client.Object) client.Object {
	return obj
}
