package version

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// The manifests applied for a Grey Matter component or dependency.
type ManifestGroup struct {
	Deployment  *appsv1.Deployment  `json:"deployment"`
	StatefulSet *appsv1.StatefulSet `json:"statefulset"`
	Services    []*corev1.Service   `json:"services"`
	ConfigMaps  []*corev1.ConfigMap `json:"configMaps"`
	Secrets     []*corev1.Secret    `json:"secrets"`
}

// Extracts manifests from a Version's cue.Value.
func (v Version) Manifests() []ManifestGroup {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var m struct {
		Manifests []ManifestGroup `json:"manifests"`
	}
	codec.Encode(v.cue, &m)
	return m.Manifests
}

// The manifests applied for Grey Matter sidecar injection.
type Sidecar struct {
	Container corev1.Container `json:"container"`
	Volumes   []corev1.Volume  `json:"volumes"`
}

// Returns a function that extracts sidecar manifests from a Version's cue.Value.
func (v Version) SidecarTemplate() func(string) Sidecar {
	return func(xdsCluster string) Sidecar {
		//lint:ignore SA1019 will update to Context in next Cue version
		codec := gocodec.New(&cue.Runtime{}, nil)
		var s struct {
			Sidecar `json:"sidecar"`
		}
		codec.Encode(v.cue.Unify(Cue(
			fmt.Sprintf(`sidecar: xdsCluster: "%s"`, xdsCluster),
		)), &s)
		return s.Sidecar
	}
}
