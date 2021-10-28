package version

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/greymatter-io/operator/pkg/cueutils"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

// The manifests applied for a Grey Matter component or dependency.
type ManifestGroup struct {
	Deployment  *appsv1.Deployment  `json:"deployment"`
	StatefulSet *appsv1.StatefulSet `json:"statefulset"`
	Service     *corev1.Service     `json:"service"`
	ConfigMaps  []*corev1.ConfigMap `json:"configMaps"`
	Secrets     []*corev1.Secret    `json:"secrets"`
	Ingress     *netv1.Ingress      `json:"ingress"`
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
	Container    corev1.Container `json:"container"`
	Volumes      []corev1.Volume  `json:"volumes"`
	StaticConfig json.RawMessage  `json:"staticConfig"`
}

// Returns a function that extracts sidecar manifests from a Version's cue.Value.
func (v Version) SidecarTemplate() func(string) Sidecar {
	return func(xdsCluster string) Sidecar {
		//lint:ignore SA1019 will update to Context in next Cue version
		codec := gocodec.New(&cue.Runtime{}, nil)
		var s struct {
			Sidecar `json:"sidecar"`
		}

		injected := v.cue.Unify(injectXDSCluster(xdsCluster))
		codec.Encode(injected, &s)

		if len(string(s.Sidecar.StaticConfig)) > 0 {
			s.Sidecar.Container.Env = append(s.Sidecar.Container.Env, corev1.EnvVar{
				Name:  "ENVOY_CONFIG",
				Value: base64.RawStdEncoding.EncodeToString(s.Sidecar.StaticConfig),
			})
		}

		return s.Sidecar
	}
}

func injectXDSCluster(xdsCluster string) cue.Value {
	b := make([]byte, 10)
	rand.Read(b)
	node := strings.TrimSuffix(base64.URLEncoding.EncodeToString(b), "==")

	switch xdsCluster {
	case "control":
		return cueutils.FromStrings(fmt.Sprintf(`sidecar: {
			xdsCluster: "%s"
			node: "%s"
			controlHost: "127.0.0.1"
		}`, xdsCluster, node))

	case "catalog", "gm-redis":
		return cueutils.FromStrings(fmt.Sprintf(`sidecar: {
			xdsCluster: "%s"
			node: "%s"
		}`, xdsCluster, node))

	default:
		return cueutils.FromStrings(fmt.Sprintf(`sidecar: {
			xdsCluster: "%s"
		}`, xdsCluster))
	}
}
