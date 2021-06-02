package common

import (
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentReconciler struct{}

func (dr DeploymentReconciler) Object() client.Object {
	return &appsv1.Deployment{}
}

func (dr DeploymentReconciler) Build(mesh *installv1.Mesh, svc gmcore.SvcName) (client.Object, error) {
	configs := gmcore.Configs(mesh.Spec.Version)

	matchLabels := map[string]string{
		"greymatter.io/control": string(svc),
	}

	podLabels := map[string]string{
		"greymatter.io/control":         string(svc),
		"greymatter.io/component":       configs[svc].Component,
		"greymatter.io/service-version": configs[svc].ImageTag,
	}
	if svc != gmcore.Control {
		podLabels["greymatter.io/sidecar-version"] = configs[gmcore.Proxy].ImageTag
	}

	objectLabels := map[string]string{
		"app.kubernetes.io/name":       string(svc),
		"app.kubernetes.io/version":    configs[svc].ImageTag,
		"app.kubernetes.io/part-of":    "greymatter",
		"app.kubernetes.io/managed-by": "gm-operator",
		"app.kubernetes.io/created-by": "gm-operator",
	}
	for k, v := range podLabels {
		objectLabels[k] = v
	}

	envsMap := configs[svc].MkEnvsMap(mesh, svc)
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

	containers := []corev1.Container{svcContainer}

	if svc != gmcore.Control {
		proxyEnvsMap := configs[gmcore.Proxy].MkEnvsMap(mesh, svc)
		var proxyEnvs []corev1.EnvVar
		for k, v := range proxyEnvsMap {
			proxyEnvs = append(proxyEnvs, corev1.EnvVar{Name: k, Value: v})
		}
		containers = append(
			containers,
			corev1.Container{
				Name:            "sidecar",
				Image:           fmt.Sprintf("docker.greymatter.io/release/gm-proxy:%s", configs[gmcore.Proxy].ImageTag),
				Env:             proxyEnvs,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Ports:           configs[gmcore.Proxy].ContainerPorts,
				Resources:       *configs[gmcore.Proxy].Resources,
			},
		)
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(svc),
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

	return deployment, nil
}

func (dr DeploymentReconciler) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	configs := gmcore.Configs(mesh.Spec.Version)

	svc, err := gmcore.ServiceName(obj.GetName())
	if err != nil {
		return false, err
	}

	labels := obj.GetLabels()
	if lbl := labels["greymatter.io/service-version"]; lbl != configs[svc].ImageTag {
		return false, nil
	}
	if lbl := labels["greymatter.io/sidecar-version"]; svc != gmcore.Control && lbl != configs[gmcore.Proxy].ImageTag {
		return false, nil
	}

	return true, nil
}
