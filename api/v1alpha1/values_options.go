package v1alpha1

import corev1 "k8s.io/api/core/v1"

func (v *Values) Overlay(opts ...func(*Values)) {
	for _, opt := range opts {
		opt(v)
	}
}

func WithImage(img string) func(*Values) {
	return func(values *Values) {
		values.Image = img
	}
}

func WithResources(r corev1.ResourceRequirements) func(*Values) {
	return func(values *Values) {
		values.Resources = &r
	}
}

func WithLabel(k, v string) func(*Values) {
	return func(values *Values) {
		if values.Labels == nil {
			values.Labels = make(map[string]string)
		}
		values.Labels[k] = v
	}
}

func WithPort(k string, v corev1.ContainerPort) func(*Values) {
	return func(values *Values) {
		if values.Ports == nil {
			values.Ports = make(map[string]corev1.ContainerPort)
		}
		values.Ports[k] = v
	}
}

func WithEnv(k, v string) func(*Values) {
	return func(values *Values) {
		if values.Env == nil {
			values.Env = make(map[string]string)
		}
		values.Env[k] = v
	}
}

func WithEnvFrom(k string, v corev1.EnvVarSource) func(*Values) {
	return func(values *Values) {
		if values.EnvFrom == nil {
			values.EnvFrom = make(map[string]corev1.EnvVarSource)
		}
		values.EnvFrom[k] = v
	}
}

func WithVolume(k string, v corev1.VolumeSource) func(*Values) {
	return func(values *Values) {
		if values.Volumes == nil {
			values.Volumes = make(map[string]corev1.VolumeSource)
		}
		values.Volumes[k] = v
	}
}

func WithVolumeMount(k string, v corev1.VolumeMount) func(*Values) {
	return func(values *Values) {
		if values.VolumeMounts == nil {
			values.VolumeMounts = make(map[string]corev1.VolumeMount)
		}
		values.VolumeMounts[k] = v
	}
}
