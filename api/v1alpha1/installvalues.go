package v1alpha1

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
// This keeps a InstallValuesConfig succinct since duplicate values don't need
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
func Redis(rc RedisConfig, msh Mesh) func(*InstallValues) {

	// If redisConfig is nil then we need to add in redis values to create a redis deployment
	// Otherwise we need to

	return func(installValues *InstallValues) {

		//Initialize default redis values to be reasigned based on redis config
		redis_host := ""
		redis_port := "6379"
		redis_password := "redis" //TODO: randomize this internal only password
		redis_db := "0"

		if rc.Url == "" {

			svc_name := "redis"
			redis_host = fmt.Sprintf("%s.%s.svc.cluster.local", svc_name, msh.Namespace)
			redis_port = "6379"
			redis_port_int, _ := strconv.Atoi(redis_port)
			redis_password = "redis"
			redis_db = "0"

			// Add redis values if we need to create one
			installValues.Redis.With(
				Image("bitnami/redis:5.0.12"),
				Command("redis-server"),
				Args([]string{"--appendonly", "yes", "--requirepass", "$(REDIS_PASSWORD)", "--dir", "/data"}),
				Envs(map[string]string{
					"REDIS_PASSWORD": redis_password,
				}),
				Port("internal-greymatter-redis", corev1.ContainerPort{
					Name:          svc_name,
					ContainerPort: int32(redis_port_int),
				}),
				Resources(&corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    resource.MustParse("200m"),
						"memory": resource.MustParse("500Mi"),
					},
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("100m"),
						"memory": resource.MustParse("128Mi"),
					},
				}),
				//TODO: add /data volume and volume mount
			)

			// add volume and volume mount if tls secret exists

		} else {
			//Parse given redis url and assign values to previously initialized variables
			u, err := url.Parse(rc.Url)
			if err != nil {
				panic(err)
			}
			redis_host, redis_port, _ = net.SplitHostPort(u.Host)
			redis_password = "redis"
			redis_db = strings.ReplaceAll(u.Path, "/", "")
		}

		//modify controlapi values
		installValues.ControlAPI.With(
			Envs(map[string]string{
				"GM_CONTROL_API_REDIS_HOST": redis_host,
				"GM_CONTROL_API_REDIS_PORT": string(redis_port),
				"GM_CONTROL_API_REDIS_PASS": redis_password,
				"GM_CONTROL_API_REDIS_DB":   string(redis_db),
			}),
		)
		//modify catalog values
		installValues.Catalog.With(
			Envs(map[string]string{
				"REDIS_HOST": redis_host,
				"REDIS_PORT": string(redis_port),
				"REDIS_PASS": redis_password,
				"REDIS_DB":   string(redis_db),
			}),
		)
		//modify jwtSecurity values
		installValues.JWTSecurity.With(
			Envs(map[string]string{
				"REDIS_HOST": redis_host,
				"REDIS_PORT": string(redis_port),
				"REDIS_PASS": redis_password,
				"REDIS_DB":   string(redis_db),
			}),
		)

	}
}
