package version

import (
	corev1 "k8s.io/api/core/v1"
)

// Values contain ContainerValues for each Grey Matter core service and dependencies.
type Values struct {
	// Values for injecting proxy containers into deployments/statefulsets.
	Proxy ContainerValues `json:"proxy"`
	// Values for defining a Grey Matter Edge deployment.
	Edge ContainerValues `json:"edge"`
	// Values for defining a Grey Matter Control container in the control deployment.
	Control ContainerValues `json:"control"`
	// Values for defining a Grey Matter Control API container in the control deployment.
	ControlAPI ContainerValues `json:"control_api"`
	// Values for defining a Grey Matter Catalog deployment.
	Catalog ContainerValues `json:"catalog"`
	// Values for defining a Grey Matter Dashboard deployment.
	Dashboard ContainerValues `json:"dashboard"`
	// Values for defining a Grey Matter JWT Security Service deployment.
	JWTSecurity ContainerValues `json:"jwt_security"`
	// Values for defining a Redis deployment. Optional.
	Redis ContainerValues `json:"redis"`
	// Values for defining a Prometheus deployment.
	Prometheus ContainerValues `json:"prometheus"`
}

type ContainerValues struct {
	// Docker image name.
	Image string `json:"image,omitempty"`
	// Command to override container entry point.
	Command string `json:"command,omitempty"`
	// Arguments to append to command when overriting container entry point.
	Args []string `json:"args,omitempty"`
	// Compute resources required by the container.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Labels to add to the Deployment/StatefulSet and its Template.Spec.
	Labels map[string]string `json:"labels,omitempty"`
	// *Map* of ports to expose from the container.
	Ports map[string]int32 `json:"ports,omitempty"`
	// *Map* of *value* (string) environment variables to set in the container.
	Envs map[string]string `json:"envs,omitempty"`
	// *Map* of *valueFrom* environment variables to set in the container.
	EnvsFrom map[string]corev1.EnvVarSource `json:"envsFrom,omitempty"`
	// *Map* of volumes that should be mounted by the container.
	Volumes map[string]corev1.VolumeSource `json:"volumes,omitempty"`
	// *Map* of pod volumes to mount into the container's filesystem.
	VolumeMounts map[string]corev1.VolumeMount `json:"volumeMounts,omitempty"`
}
