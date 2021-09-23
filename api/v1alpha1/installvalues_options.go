package v1alpha1

import corev1 "k8s.io/api/core/v1"

func (installValues *InstallValues) Overlay(opts ...func(*InstallValues)) {
	for _, opt := range opts {
		opt(installValues)
	}
}

// A InstallValues option that adds Proxy values to Edge values.
// This keeps the InstallValuesConfig succinct since duplicate values don't need
// to be defined for both Proxy and Edge. Edge values should just be overrides.
func WithEdgeValuesFromProxy(installValues *InstallValues) {
	values := Values{}

	values.Image = installValues.Proxy.Image
	if installValues.Edge.Image != "" {
		values.Image = installValues.Edge.Image
	}

	values.Resources = installValues.Proxy.Resources
	if installValues.Edge.Resources != nil {
		values.Resources = installValues.Edge.Resources
	}

	if len(installValues.Proxy.Labels)+len(installValues.Edge.Labels) > 0 {
		values.Labels = make(map[string]string)
	}
	for k, v := range installValues.Proxy.Labels {
		values.Labels[k] = v
	}
	for k, v := range installValues.Edge.Labels {
		values.Labels[k] = v
	}

	if len(installValues.Proxy.Ports)+len(installValues.Edge.Ports) > 0 {
		values.Ports = make(map[string]corev1.ContainerPort)
	}
	for k, v := range installValues.Proxy.Ports {
		values.Ports[k] = v
	}
	for k, v := range installValues.Edge.Ports {
		values.Ports[k] = v
	}

	if len(installValues.Proxy.Env)+len(installValues.Edge.Env) > 0 {
		values.Env = make(map[string]string)
	}
	for k, v := range installValues.Proxy.Env {
		values.Env[k] = v
	}
	for k, v := range installValues.Edge.Env {
		values.Env[k] = v
	}

	if len(installValues.Proxy.EnvFrom)+len(installValues.Edge.EnvFrom) > 0 {
		values.EnvFrom = make(map[string]corev1.EnvVarSource)
	}
	for k, v := range installValues.Proxy.EnvFrom {
		values.EnvFrom[k] = v
	}
	for k, v := range installValues.Edge.EnvFrom {
		values.EnvFrom[k] = v
	}

	if len(installValues.Proxy.Volumes)+len(installValues.Edge.Volumes) > 0 {
		values.Volumes = make(map[string]corev1.VolumeSource)
	}
	for k, v := range installValues.Proxy.Volumes {
		values.Volumes[k] = v
	}
	for k, v := range installValues.Edge.Volumes {
		values.Volumes[k] = v
	}

	if len(installValues.Proxy.VolumeMounts)+len(installValues.Edge.VolumeMounts) > 0 {
		values.VolumeMounts = make(map[string]corev1.VolumeMount)
	}
	for k, v := range installValues.Proxy.VolumeMounts {
		values.VolumeMounts[k] = v
	}
	for k, v := range installValues.Edge.VolumeMounts {
		values.VolumeMounts[k] = v
	}

	installValues.Edge = values
}

// A InstallValues option that injects SPIRE configuration into Proxy values.
func WithSPIRE(installValues *InstallValues) {
	installValues.Proxy.Overlay(
		WithVolume("spire-socket", corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/run/spire/socket",
				Type: func() *corev1.HostPathType {
					pathType := corev1.HostPathDirectoryOrCreate
					return &pathType
				}(),
			},
		}),
		WithVolumeMount("spire-socket", corev1.VolumeMount{
			MountPath: "/run/spire/socket",
		}),
		WithEnv("SPIRE_PATH", "/run/spire/socket/agent.sock"),
	)
}

// A InstallValues option that injects configuration for a Redis provider.
// If the Redis configuration is empty, adds Values for configuring an internal Redis.
func WithRedis(redisHost, redisPort string) func(*InstallValues) {
	return func(installValues *InstallValues) {
		// inject configs into install values for control api, catalog, jwt security
	}
}
