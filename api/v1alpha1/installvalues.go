package v1alpha1

import corev1 "k8s.io/api/core/v1"

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
func Redis(redisHost, redisPort string) func(*InstallValues) {
	return func(installValues *InstallValues) {
		// inject configs into install values for control api, catalog, jwt security
	}
}
