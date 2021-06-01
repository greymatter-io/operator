package common

import (
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MkDeployment(mesh *installv1.Mesh, svc GmCore) (*appsv1.Deployment, error) {
	replicas := int32(1)

	versionedConfigs, ok := GmCoreConfigs[mesh.Spec.Version]
	if !ok {
		return nil, fmt.Errorf("invalid Mesh.Spec.Version '%s'", mesh.Spec.Version)
	}

	meshLabels := map[string]string{
		"app.kubernetes.io/name":       string(svc),
		"app.kubernetes.io/version":    versionedConfigs[svc].imageTag,
		"app.kubernetes.io/part-of":    "greymatter",
		"app.kubernetes.io/managed-by": "gm-operator",
		"app.kubernetes.io/created-by": "gm-operator",
	}
	if svc != Control {
		meshLabels["proxy-version"] = versionedConfigs[Proxy].imageTag
	}

	labels := map[string]string{
		"greymatter":            versionedConfigs[svc].component,
		"greymatter.io/control": string(svc),
	}

	envsMap := versionedConfigs[svc].mkEnvsMap(mesh)
	var envs []corev1.EnvVar
	for k, v := range envsMap {
		envs = append(envs, corev1.EnvVar{Name: k, Value: v})
	}

	svcContainer := corev1.Container{
		Name:            string(svc),
		Image:           fmt.Sprintf("docker.greymatter.io/release/gm-%s:%s", svc, versionedConfigs[svc].imageTag),
		Env:             envs,
		ImagePullPolicy: ImagePullPolicy,
		Ports:           versionedConfigs[svc].containerPorts,
	}
	if versionedConfigs[svc].resources != nil {
		svcContainer.Resources = *versionedConfigs[svc].resources
	}

	containers := []corev1.Container{svcContainer}

	if svc != Control {
		proxyEnvsMap := versionedConfigs[svc].mkEnvsMap(mesh)
		var proxyEnvs []corev1.EnvVar
		for k, v := range proxyEnvsMap {
			proxyEnvs = append(proxyEnvs, corev1.EnvVar{Name: k, Value: v})
		}
		containers = append(
			containers,
			corev1.Container{
				Name:            "sidecar",
				Image:           fmt.Sprintf("docker.greymatter.io/release/gm-proxy:%s", versionedConfigs[Proxy].imageTag),
				Env:             proxyEnvs,
				ImagePullPolicy: ImagePullPolicy,
				Ports:           versionedConfigs[Proxy].containerPorts,
				Resources:       *versionedConfigs[Proxy].resources,
			},
		)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(svc),
			Namespace: mesh.Namespace,
			Labels:    meshLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: mesh.Spec.ImagePullSecret}},
					Containers:       containers,
				},
			},
		},
	}

	if svc == Control {
		deployment.Spec.Template.Spec.ServiceAccountName = "control-pods"
	}

	return deployment, nil
}
