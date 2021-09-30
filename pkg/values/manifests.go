package values

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ManifestGroup struct {
	Deployment *appsv1.Deployment
	Services   []*corev1.Service
	// TODO: ConfigMaps, PVCs, etc.
}

func (v Values) GenerateManifests() []ManifestGroup {
	return []ManifestGroup{
		{
			Deployment: v.Redis.Deployment(),
			Services:   v.Redis.Services(),
		},
		{
			Deployment: v.ControlAPI.Deployment(),
			Services:   v.ControlAPI.Services(),
		},
		{
			// TODO: Add an init container that pings Control API
			Deployment: v.Control.Deployment(),
			Services:   v.Control.Services(),
		},
		{
			Deployment: v.Edge.Deployment(),
			Services:   v.Edge.Services(),
		},
		{
			Deployment: v.Catalog.Deployment(),
			Services:   v.Catalog.Services(),
		},
		{
			Deployment: v.Dashboard.Deployment(),
			Services:   v.Dashboard.Services(),
		},
		{
			Deployment: v.JWTSecurity.Deployment(),
			Services:   v.JWTSecurity.Services(),
		},
		{
			Deployment: v.Prometheus.Deployment(),
			Services:   v.Prometheus.Services(),
		},
	}
}

func (cv ContainerValues) Deployment() *appsv1.Deployment {
	return &appsv1.Deployment{}
}

func (cv ContainerValues) Services() []*corev1.Service {
	return []*corev1.Service{}
}
