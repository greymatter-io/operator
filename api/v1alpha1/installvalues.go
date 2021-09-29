package v1alpha1

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	redis "github.com/go-redis/redis/v8"
	corev1 "k8s.io/api/core/v1"
)

// InstallValues are values used for installing a Grey Matter mesh.
type InstallValues struct {
	// Values for injecting proxy containers into deployments/statefulsets.
	Proxy *Values `json:"proxy"`
	// Values for defining a Grey Matter Edge deployment.
	Edge *Values `json:"edge"`
	// Values for defining a Grey Matter Control container in the control deployment.
	Control *Values `json:"control"`
	// Values for defining a Grey Matter Control API container in the control deployment.
	ControlAPI *Values `json:"controlApi"`
	// Values for defining a Grey Matter Catalog deployment.
	Catalog *Values `json:"catalog"`
	// Values for defining a Grey Matter Dashboard deployment.
	Dashboard *Values `json:"dashboard"`
	// Values for defining a Grey Matter JWT Security Service deployment.
	JWTSecurity *Values `json:"jwtSecurity"`
	// Values for defining a Redis deployment. Optional.
	Redis *Values `json:"redis"`
	// Values for defining a Prometheus deployment. Optional.
	Prometheus *Values `json:"prometheus"`
}

func (installValues *InstallValues) With(opts ...func(*InstallValues)) *InstallValues {
	for _, opt := range opts {
		opt(installValues)
	}
	return installValues
}

// A InstallValues option that adds Proxy values to Edge values.
// This keeps an InstallationConfig succinct since duplicate values don't need
// to be defined for both Proxy and Edge. Edge values should just be overrides.
func WithEdgeValuesFromProxy(installValues *InstallValues) {
	installValues.Edge = (&Values{}).With(
		// First apply all non-nil Proxy values
		Image(installValues.Proxy.Image),
		Resources(installValues.Proxy.Resources),
		Labels(installValues.Proxy.Labels),
		Ports(installValues.Proxy.Ports),
		Envs(installValues.Proxy.Envs),
		EnvsFrom(installValues.Proxy.EnvsFrom),
		Volumes(installValues.Proxy.Volumes),
		VolumeMounts(installValues.Proxy.VolumeMounts),
		// Then apply all non-nil Edge values
		Image(installValues.Edge.Image),
		Resources(installValues.Edge.Resources),
		Labels(installValues.Edge.Labels),
		Ports(installValues.Edge.Ports),
		Envs(installValues.Edge.Envs),
		EnvsFrom(installValues.Edge.EnvsFrom),
		Volumes(installValues.Edge.Volumes),
		VolumeMounts(installValues.Edge.VolumeMounts),
	)
}

// A InstallValues option that injects SPIRE configuration into Proxy values.
func SPIRE(installValues *InstallValues) {
	installValues.Proxy.With(
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

// A InstallValues option that injects configuration for a Redis provider.
// If the Redis configuration is empty, adds Values for configuring an internal Redis.
func Redis(rc *RedisConfig, namespace string) func(*InstallValues) {

	// If redisConfig is nil then we need to add in redis values to create a redis deployment
	// Otherwise we need to

	return func(installValues *InstallValues) {

		//Initialize default redis values to be reasigned based on redis config
		var redisHost string
		var redisPassword string
		//TODO: add ablility to use separate db for catalog and control-api
		var redisDB string
		var redisPort string

		if installValues.Redis != nil {
			redisHost = fmt.Sprintf("internal-greymatter-redis.%s.cluster.local", namespace)
			redisDB = "0"
			redisPort = "6379"

			// Generate an 8 character random password
			b := make([]byte, 8)
			rand.Read(b)
			redisPassword = base64.URLEncoding.EncodeToString(b)

			installValues.Redis.With(Env("REDIS_PASSWORD", redisPassword))
		}

		//if a redisConfig is provided then do not use the defaults
		if rc != nil && rc.Url != "" {
			installValues.Redis = nil

			//parse the given url
			//TODO: in webhook validate the url is parseable
			redisOptions, _ := redis.ParseURL(rc.Url)

			// username := redisOptions.Username
			redisPassword = redisOptions.Password
			hostPort := redisOptions.Addr
			split := strings.Split(hostPort, ":")
			redisHost, redisPort = split[0], split[1]
			redisDB = fmt.Sprintf("%d", redisOptions.DB)

		}

		fmt.Printf("%s %s %s %s", redisHost, redisPort, redisPassword, redisDB)

		//modify controlapi values
		installValues.ControlAPI.With(
			Envs(map[string]string{
				"GM_CONTROL_API_REDIS_HOST": redisHost,
				"GM_CONTROL_API_REDIS_PORT": redisPort,
				"GM_CONTROL_API_REDIS_PASS": redisPassword,
				"GM_CONTROL_API_REDIS_DB":   redisDB,
			}),
		)
		//modify catalog values
		installValues.Catalog.With(
			Envs(map[string]string{
				"REDIS_HOST": redisHost,
				"REDIS_PORT": redisPort,
				"REDIS_PASS": redisPassword,
				"REDIS_DB":   redisDB,
			}),
		)
		//modify jwtSecurity values
		installValues.JWTSecurity.With(
			Envs(map[string]string{
				"REDIS_HOST": redisHost,
				"REDIS_PORT": redisPort,
				"REDIS_PASS": redisPassword,
				"REDIS_DB":   redisDB,
			}),
		)

	}
}
