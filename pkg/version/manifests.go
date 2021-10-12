package version

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/greymatter-io/operator/pkg/cueutils"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

// The manifests applied for a Grey Matter component or dependency.
type ManifestGroup struct {
	Deployment  *appsv1.Deployment  `json:"deployment"`
	StatefulSet *appsv1.StatefulSet `json:"statefulset"`
	Service     *corev1.Service     `json:"service"`
	ConfigMaps  []*corev1.ConfigMap `json:"configMaps"`
	Secrets     []*corev1.Secret    `json:"secrets"`
	Route       *routev1.Route      `json:"route"`
	Ingress     *extv1beta1.Ingress `json:"ingress"`
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
		codec.Encode(v.cue.Unify(cueutils.FromStrings(
			fmt.Sprintf(`sidecar: xdsCluster: "%s"`, xdsCluster),
		)), &s)
		return s.Sidecar
	}
}
