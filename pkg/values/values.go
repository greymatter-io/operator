package values

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	redis "github.com/go-redis/redis/v8"
	corev1 "k8s.io/api/core/v1"
)

//+kubebuilder:object:generate=true

// Values contain ContainerValues for each Grey Matter core service and dependencies.
type Values struct {
	// Values for injecting proxy containers into deployments/statefulsets.
	Proxy *ContainerValues `json:"proxy"`
	// Values for defining a Grey Matter Edge deployment.
	Edge *ContainerValues `json:"edge"`
	// Values for defining a Grey Matter Control container in the control deployment.
	Control *ContainerValues `json:"control"`
	// Values for defining a Grey Matter Control API container in the control deployment.
	ControlAPI *ContainerValues `json:"control_api"`
	// Values for defining a Grey Matter Catalog deployment.
	Catalog *ContainerValues `json:"catalog"`
	// Values for defining a Grey Matter Dashboard deployment.
	Dashboard *ContainerValues `json:"dashboard"`
	// Values for defining a Grey Matter JWT Security Service deployment.
	JWTSecurity *ContainerValues `json:"jwt_security"`
	// Values for defining a Redis deployment. Optional.
	Redis *ContainerValues `json:"redis"`
	// Values for defining a Prometheus deployment. Optional.
	Prometheus *ContainerValues `json:"prometheus"`
}

type InstallOpt func(*Values)

func (v *Values) Apply(opts ...InstallOpt) {
	for _, opt := range opts {
		opt(v)
	}
}

// A ValuesOpt that adds Proxy values to Edge values.
// This keeps an Values file succinct since duplicate values don't need
// to be defined for both Proxy and Edge. Edge values should just be overrides.
func EdgeValuesFromProxy(v *Values) {
	edge := &ContainerValues{}
	v.Edge.Apply(
		// First apply all non-nil Proxy values
		Image(v.Proxy.Image),
		Resources(v.Proxy.Resources),
		Labels(v.Proxy.Labels),
		Ports(v.Proxy.Ports),
		Envs(v.Proxy.Envs),
		EnvsFrom(v.Proxy.EnvsFrom),
		Volumes(v.Proxy.Volumes),
		VolumeMounts(v.Proxy.VolumeMounts),
		// Then apply all non-nil Edge values
		Image(v.Edge.Image),
		Resources(v.Edge.Resources),
		Labels(v.Edge.Labels),
		Ports(v.Edge.Ports),
		Envs(v.Edge.Envs),
		EnvsFrom(v.Edge.EnvsFrom),
		Volumes(v.Edge.Volumes),
		VolumeMounts(v.Edge.VolumeMounts),
	)
	v.Edge = edge
}

// A ValuesOpt that injects SPIRE configuration.
func SPIRE(v *Values) {
	v.Proxy.Apply(
		Volume("spire-socket", corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/run/spire/socket",
				Type: func() *corev1.HostPathType {
					pathType := corev1.HostPathDirectoryOrCreate
					return &pathType
				}(),
			},
		}),
		VolumeMount("spire-socket", corev1.VolumeMount{
			MountPath: "/run/spire/socket",
		}),
		Env("SPIRE_PATH", "/run/spire/socket/agent.sock"),
	)
}

// TODO: Should this live somewhere else?
// TODO: Instead of `url`, require host, port, password, dbs. No username option.
type ExternalRedisConfig struct {
	URL string `json:"url"`
	// +optional
	CertSecretName string `json:"cert_secret_name"`
}

// A ValuesOpt that applies configuration for a Redis server.
// If ExternalRedisConfig is not nil, it will apply values derived from it pointing to an external Redis server.
// If ExternalRedisConfig is nil, it will apply values stored in v.Redis pointing to an internal Redis server.
func Redis(cfg *ExternalRedisConfig, namespace string) InstallOpt {
	return func(v *Values) {
		var host string
		var port string
		var password string
		var db string

		if cfg != nil && cfg.URL != "" {
			// Since an ExternalRedisConfig is provided, set v.Redis to nil so the Installer does not install an internal Redis.
			v.Redis = nil

			// TODO: In the Mesh validating webhook, ensure the user provided URL is parseable.
			// This actually might be OBE if we require the user to supply values separately rather than as a URL.
			redisOptions, _ := redis.ParseURL(cfg.URL)

			password = redisOptions.Password
			hostPort := redisOptions.Addr
			split := strings.Split(hostPort, ":")
			host, port = split[0], split[1]
			// TODO: Enable specifying separate databases
			db = fmt.Sprintf("%d", redisOptions.DB)
		} else {
			host = fmt.Sprintf("greymatter-redis.%s.svc.cluster.local", namespace)
			port = "6379"
			db = "0"

			// Generate and inject an 8 character random password
			b := make([]byte, 8)
			rand.Read(b)
			password = base64.URLEncoding.EncodeToString(b)
			v.Redis.Apply(Env("REDIS_PASSWORD", password))
		}

		v.ControlAPI.Apply(
			Envs(map[string]string{
				"GM_CONTROL_API_PERSISTER_TYPE": "redis",
				"GM_CONTROL_API_REDIS_HOST":     host,
				"GM_CONTROL_API_REDIS_PORT":     port,
				"GM_CONTROL_API_REDIS_PASS":     password,
				"GM_CONTROL_API_REDIS_DB":       db,
			}),
		)
		v.Catalog.Apply(
			Envs(map[string]string{
				"REDIS_HOST": host,
				"REDIS_PORT": port,
				"REDIS_PASS": password,
				"REDIS_DB":   db,
			}),
		)
		v.JWTSecurity.Apply(
			Envs(map[string]string{
				"REDIS_HOST": host,
				"REDIS_PORT": port,
				"REDIS_PASS": password,
				"REDIS_DB":   db,
			}),
		)
	}
}
