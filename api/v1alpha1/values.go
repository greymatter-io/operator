package v1alpha1

import corev1 "k8s.io/api/core/v1"

type Values struct {
	// Docker image name.
	Image string `json:"image,omitempty"`
	// Command to override container entry point
	Command string `json:"command,omitempty"`
	// Arguments to append to command when overriting container entry point
	Arguments []string `json:"args,omitempty"`
	// Compute resources required by the container.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Labels to add to the Deployment/StatefulSet and its Template.Spec
	Labels map[string]string `json:"labels,omitempty"`
	// *Map* of ports to expose from the container.
	Ports map[string]corev1.ContainerPort `json:"ports,omitempty"`
	// *Map* of *value* (string) environment variables to set in the container.
	Envs map[string]string `json:"envs,omitempty"`
	// *Map* of *valueFrom* environment variables to set in the container.
	EnvsFrom map[string]corev1.EnvVarSource `json:"envsFrom,omitempty"`
	// *Map* of volumes that should be mounted by the container.
	Volumes map[string]corev1.VolumeSource `json:"volumes,omitempty"`
	// *Map* of pod volumes to mount into the container's filesystem.
	VolumeMounts map[string]corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Persistent Volume Claim Template
	PersistentVolumeClaimTemplate *corev1.PersistentVolumeClaimTemplate `json:"persistentVolumeClaimTemplate,omitempty"`
}

func (v *Values) With(opts ...func(*Values)) *Values {
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func Image(img string) func(*Values) {
	return func(values *Values) {
		if img != "" {
			values.Image = img
		}
	}
}

func Command(cmd string) func(*Values) {
	return func(values *Values) {
		if len(cmd) > 0 {
			values.Command = cmd
		}
	}
}

func Args(args []string) func(*Values) {
	return func(values *Values) {
		if len(args) > 0 {
			values.Arguments = args
		}
	}
}

func Resources(r *corev1.ResourceRequirements) func(*Values) {
	return func(values *Values) {
		if r != nil {
			values.Resources = r
		}
	}
}

func Label(k, v string) func(*Values) {
	return func(values *Values) {
		if values.Labels == nil {
			values.Labels = make(map[string]string)
		}
		values.Labels[k] = v
	}
}

func Labels(labels map[string]string) func(*Values) {
	return func(values *Values) {
		if values.Labels == nil {
			values.Labels = make(map[string]string)
		}
		for k, v := range labels {
			values.Labels[k] = v
		}
	}
}

func Port(k string, v corev1.ContainerPort) func(*Values) {
	return func(values *Values) {
		if values.Ports == nil {
			values.Ports = make(map[string]corev1.ContainerPort)
		}
		values.Ports[k] = v
	}
}

func Ports(ports map[string]corev1.ContainerPort) func(*Values) {
	return func(values *Values) {
		if values.Ports == nil {
			values.Ports = make(map[string]corev1.ContainerPort)
		}
		for k, v := range ports {
			values.Ports[k] = v
		}
	}
}

func Env(k, v string) func(*Values) {
	return func(values *Values) {
		if values.Envs == nil {
			values.Envs = make(map[string]string)
		}
		values.Envs[k] = v
	}
}

func Envs(envs map[string]string) func(*Values) {
	return func(values *Values) {
		if values.Envs == nil {
			values.Envs = make(map[string]string)
		}
		for k, v := range envs {
			values.Envs[k] = v
		}
	}
}

func EnvFrom(k string, v corev1.EnvVarSource) func(*Values) {
	return func(values *Values) {
		if values.EnvsFrom == nil {
			values.EnvsFrom = make(map[string]corev1.EnvVarSource)
		}
		values.EnvsFrom[k] = v
	}
}

func EnvsFrom(envsFrom map[string]corev1.EnvVarSource) func(*Values) {
	return func(values *Values) {
		if values.EnvsFrom == nil {
			values.EnvsFrom = make(map[string]corev1.EnvVarSource)
		}
		for k, v := range envsFrom {
			values.EnvsFrom[k] = v
		}
	}
}

func Volume(k string, v corev1.VolumeSource) func(*Values) {
	return func(values *Values) {
		if values.Volumes == nil {
			values.Volumes = make(map[string]corev1.VolumeSource)
		}
		values.Volumes[k] = v
	}
}

func Volumes(volumes map[string]corev1.VolumeSource) func(*Values) {
	return func(values *Values) {
		if values.Volumes == nil {
			values.Volumes = make(map[string]corev1.VolumeSource)
		}
		for k, v := range volumes {
			values.Volumes[k] = v
		}
	}
}

func VolumeMount(k string, v corev1.VolumeMount) func(*Values) {
	return func(values *Values) {
		if values.VolumeMounts == nil {
			values.VolumeMounts = make(map[string]corev1.VolumeMount)
		}
		values.VolumeMounts[k] = v
	}
}

func VolumeMounts(volumeMounts map[string]corev1.VolumeMount) func(*Values) {
	return func(values *Values) {
		if values.VolumeMounts == nil {
			values.VolumeMounts = make(map[string]corev1.VolumeMount)
		}
		for k, v := range volumeMounts {
			values.VolumeMounts[k] = v
		}
	}
}
