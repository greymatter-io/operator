package values

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ManifestGroup struct {
	Deployment *appsv1.Deployment
	Services   []*corev1.Service
}

func (v Values) GenerateManifests() []ManifestGroup {
	var manifests []ManifestGroup
	return manifests
}
