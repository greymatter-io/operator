package version

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ManifestGroup struct {
	Deployment *appsv1.Deployment
	Services   []*corev1.Service
	// TODO: ConfigMaps, PVCs, etc.
	// TODO: Inject certs, base64, etc. using Cue; see Redis options for example
	// Possibly use templating: https://cuetorials.com/first-steps/generate-all-the-things/
	// Tools for templates: https://github.com/Masterminds/sprig
}

func (ics InstallConfigs) Manifests() []ManifestGroup {
	return []ManifestGroup{}
}

func manifests(expose []string, ic ...InstallConfig) ManifestGroup {
	// Create deployment with labels
	// Create service with selectors
	return ManifestGroup{}
}
