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
	Deployment *appsv1.Deployment `json:"deployment"`
	Services   []*corev1.Service  `json:"services"`
	// TODO: ConfigMaps, PVCs, etc.
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
	codec.Encode(v.cue, &m)
	return m.Manifests
}

func (v Version) Sidecar() *corev1.Container {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var c struct {
		Sidecar *corev1.Container `json:"sidecar"`
	}
	codec.Encode(v.cue, &c)
	return c.Sidecar
}

func (v *Version) Apply(opts ...InstallOption) {
	for _, opt := range opts {
		opt(v)
	}
}

type InstallOption func(*Version)

var (
	spire         *cue.Value
	redisInternal *cue.Value
	redisExternal *cue.Value
)

func Namespace(namespace string) InstallOption {
	return func(v *Version) {
		injected := Cue(fmt.Sprintf(`Namespace: "%s"`, namespace))
		if err := injected.Err(); err != nil {
			logger.Error(err, "failed to inject apply option", "option", "Namespace")
			return
		}
		v.cue = v.cue.Unify(injected)
	}
}

// An InstallOption for injecting SPIRE configuration into Proxy values.
func SPIRE(v *Version) {
	if spire == nil {
		value, err := loadOption("spire.cue")
		if err != nil {
			return
		}
		spire = &value
	}
	v.cue = v.cue.Unify(*spire)
}

// An InstallOption for injecting configuration for an internal Redis deployment.
func InternalRedis(v *Version) {
	if redisInternal == nil {
		value, err := loadOption("redis_internal.cue")
		if err != nil {
			return
		}
		redisInternal = &value
	}
	b := make([]byte, 10)
	rand.Read(b)
	password := base64.URLEncoding.EncodeToString(b)
	injected := Cue(fmt.Sprintf(`password: "%s"`, password))
	if err := injected.Err(); err != nil {
		logger.Error(err, "failed to inject apply option", "option", "internal Redis")
		return
	}
	v.cue = v.cue.Unify(*redisInternal).Unify(injected)
}

// TODO: Instead of `url`, require host, port, password, dbs. No username option.
type ExternalRedisConfig struct {
	URL string `json:"url"`
	// +optional
	CertSecretName string `json:"cert_secret_name"`
}

// An InstallOption for injecting configuration for an external Redis server.
func ExternalRedis(cfg *ExternalRedisConfig) InstallOption {
	// TODO: In the Mesh validating webhook, ensure the user provided URL is parseable.
	// This actually might be OBE if we require the user to supply values separately rather than as a URL.
	redisOptions, _ := redis.ParseURL(cfg.URL)
	hostPort := redisOptions.Addr
	split := strings.Split(hostPort, ":")
	host, port := split[0], split[1]
	password := redisOptions.Password
	db := fmt.Sprintf("%d", redisOptions.DB)

	return func(v *Version) {
		if redisExternal == nil {
			value, err := loadOption("redis_external.cue")
			if err != nil {
				return
			}
			redisExternal = &value
		}
		injected := Cue(fmt.Sprintf(`
			host: "%s"
			port: "%s"
			password: "%s"
			db: "%s"
		`, host, port, password, db))
		if err := injected.Err(); err != nil {
			logger.Error(err, "failed to inject apply option", "option", "internal Redis")
			return
		}
		v.cue = v.cue.Unify(*redisExternal).Unify(injected)
	}
}

func loadOption(fileName string) (cue.Value, error) {
	data, err := filesystem.ReadFile(fmt.Sprintf("options/%s", fileName))
	if err != nil {
		logger.Error(err, "failed to load option", "file", fileName)
		return cue.Value{}, err
	}
	value := Cue(string(data))
	if err := value.Err(); err != nil {
		logger.Error(err, "failed to parse option", "file", fileName)
	}
	return value, nil
}
