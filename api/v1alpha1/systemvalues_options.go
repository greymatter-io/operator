package v1alpha1

import corev1 "k8s.io/api/core/v1"

func (sv *SystemValues) Overlay(opts ...func(*SystemValues)) {
	for _, opt := range opts {
		opt(sv)
	}
}

// A SystemValues option that adds Proxy values to Edge values.
// This keeps the SystemValuesConfig succinct since duplicate values don't need
// to be defined for both Proxy and Edge. Edge values should just be overrides.
func WithEdgeValuesFromProxy(sv *SystemValues) {
	values := Values{}

	values.Image = sv.Proxy.Image
	if sv.Edge.Image != "" {
		values.Image = sv.Edge.Image
	}

	values.Resources = sv.Proxy.Resources
	if sv.Edge.Resources != nil {
		values.Resources = sv.Edge.Resources
	}

	if len(sv.Proxy.Labels)+len(sv.Edge.Labels) > 0 {
		values.Labels = make(map[string]string)
	}
	for k, v := range sv.Proxy.Labels {
		values.Labels[k] = v
	}
	for k, v := range sv.Edge.Labels {
		values.Labels[k] = v
	}

	if len(sv.Proxy.Ports)+len(sv.Edge.Ports) > 0 {
		values.Ports = make(map[string]corev1.ContainerPort)
	}
	for k, v := range sv.Proxy.Ports {
		values.Ports[k] = v
	}
	for k, v := range sv.Edge.Ports {
		values.Ports[k] = v
	}

	if len(sv.Proxy.Env)+len(sv.Edge.Env) > 0 {
		values.Env = make(map[string]string)
	}
	for k, v := range sv.Proxy.Env {
		values.Env[k] = v
	}
	for k, v := range sv.Edge.Env {
		values.Env[k] = v
	}

	if len(sv.Proxy.EnvFrom)+len(sv.Edge.EnvFrom) > 0 {
		values.EnvFrom = make(map[string]corev1.EnvVarSource)
	}
	for k, v := range sv.Proxy.EnvFrom {
		values.EnvFrom[k] = v
	}
	for k, v := range sv.Edge.EnvFrom {
		values.EnvFrom[k] = v
	}

	if len(sv.Proxy.Volumes)+len(sv.Edge.Volumes) > 0 {
		values.Volumes = make(map[string]corev1.VolumeSource)
	}
	for k, v := range sv.Proxy.Volumes {
		values.Volumes[k] = v
	}
	for k, v := range sv.Edge.Volumes {
		values.Volumes[k] = v
	}

	if len(sv.Proxy.VolumeMounts)+len(sv.Edge.VolumeMounts) > 0 {
		values.VolumeMounts = make(map[string]corev1.VolumeMount)
	}
	for k, v := range sv.Proxy.VolumeMounts {
		values.VolumeMounts[k] = v
	}
	for k, v := range sv.Edge.VolumeMounts {
		values.VolumeMounts[k] = v
	}

	sv.Edge = values
}

// A SystemValues option that injects SPIRE configuration into Proxy values.
func WithSPIRE(sv *SystemValues) {
	sv.Proxy.Overlay(
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
