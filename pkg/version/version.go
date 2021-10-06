package version

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/go-redis/redis/v8"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Version struct {
	cue cue.Value
}

func (v Version) Copy() Version {
	return Version{v.cue}
}

type ManifestGroup struct {
	Deployment *appsv1.Deployment  `json:"deployment"`
	Services   []*corev1.Service   `json:"services"`
	ConfigMaps []*corev1.ConfigMap `json:"configMaps"`
	// TODO: PVCs, etc.
	// TODO: Inject certs, base64, etc. using Cue; see Redis options for example
	// Possibly use templating: https://cuetorials.com/first-steps/generate-all-the-things/
	// Tools for templates: https://github.com/Masterminds/sprig
}

func (v Version) Manifests() []ManifestGroup {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var m struct {
		Manifests []ManifestGroup `json:"manifests"`
	}
	// Encode Cue value into Go struct
	codec.Encode(v.cue, &m)
	return m.Manifests
}

type Sidecar struct {
	Container *corev1.Container `json:"container"`
	Volumes   []corev1.Volume   `json:"volumes"`
}

func (v Version) Sidecar() Sidecar {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var s struct {
		Sidecar `json:"sidecar"`
	}
	// Encode Cue value into Go struct
	codec.Encode(v.cue, &s)
	return s.Sidecar
}

func (v *Version) Apply(opts ...InstallOption) {
	for _, opt := range opts {
		opt(v)
	}
}

type InstallOption func(*Version)

// An InstallOption for injecting an InstallNamespace value.
func InstallNamespace(namespace string) InstallOption {
	return func(v *Version) {
		v.cue = v.cue.Unify(Cue(fmt.Sprintf(`InstallNamespace: "%s"`, namespace)))
	}
}

// An InstallOption for injectign a Zone value.
func Zone(zone string) InstallOption {
	return func(v *Version) {
		v.cue = v.cue.Unify(Cue(fmt.Sprintf(`Zone: "%s"`, zone)))
	}
}

// An InstallOption for injecting an ImagePullSecretName value.
// Note that this value is not sourced from the Mesh CR spec, but by the Installer.
func ImagePullSecretName(imagePullSecretName string) InstallOption {
	return func(v *Version) {
		v.cue = v.cue.Unify(Cue(fmt.Sprintf(`ImagePullSecretName: "%s"`, imagePullSecretName)))
	}
}

// An InstallOption for injecting SPIRE configuration.
func SPIRE(v *Version) {
	v.cue = v.cue.Unify(Cue(`Spire: true`))
}

// An InstallOption for injecting Redis configuration for either an external
// Redis server (if the config is not nil) or otherwise an internal Redis deployment.
func Redis(externalURL string) InstallOption {
	return func(v *Version) {
		if externalURL == "" {
			b := make([]byte, 10)
			rand.Read(b)
			password := base64.URLEncoding.EncodeToString(b)
			v.cue = v.cue.Unify(Cue(fmt.Sprintf(`Redis: password: "%s"`, password)))
			return
		}

		// TODO: In the Mesh validating webhook, ensure the user provided URL is parseable.
		// This actually might be OBE if we require the user to supply values separately rather than as a URL.
		redisOptions, _ := redis.ParseURL(externalURL)
		hostPort := redisOptions.Addr
		split := strings.Split(hostPort, ":")
		host, port := split[0], split[1]
		password := redisOptions.Password
		db := fmt.Sprintf("%d", redisOptions.DB)
		v.cue = v.cue.Unify(Cue(fmt.Sprintf(
			`Redis: {
				host: "%s"
				port: "%s"
				password: "%s"
				db: "%s"
			}`,
			host, port, password, db)),
		)
	}
}

func UserTokens(users string) InstallOption {
	return func(v *Version) {
		v.cue = v.cue.Unify(Cue(fmt.Sprintf(`UserTokens: "%s"`, users)))
	}
}
