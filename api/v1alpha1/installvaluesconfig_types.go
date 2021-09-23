/*
Copyright Decipher Technology Studios 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstallValues defines the desired state of InstallValuesConfig
type InstallValues struct {
	// Values for injecting proxy containers into deployments/statefulsets.
	Proxy Values `json:"proxy"`
	// Values for defining a Grey Matter Edge deployment.
	Edge Values `json:"edge"`
	// Values for defining a Grey Matter Control container in the control deployment.
	Control Values `json:"control"`
	// Values for defining a Grey Matter Control API container in the control deployment.
	ControlAPI Values `json:"controlApi"`
	// Values for defining a Grey Matter Catalog deployment.
	Catalog Values `json:"catalog"`
	// Values for defining a Grey Matter Dashboard deployment.
	Dashboard Values `json:"dashboard"`
	// Values for defining a Grey Matter JWT Security Service deployment.
	JWTSecurity Values `json:"jwtSecurity"`
	// Values for defining a Redis deployment. Optional.
	Redis Values `json:"redis"`
	// Values for defining a Prometheus deployment. Optional.
	Prometheus Values `json:"prometheus"`
}

type Values struct {
	// Docker image name.
	Image string `json:"image,omitempty"`
	// Compute resources required by the container.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Labels to add to the Deployment/StatefulSet and its Template.Spec
	Labels map[string]string `json:"labels,omitempty"`
	// *Map* of ports to expose from the container.
	Ports map[string]corev1.ContainerPort `json:"ports,omitempty"`
	// *Map* of *value* (string) environment variables to set in the container.
	Env map[string]string `json:"env,omitempty"`
	// *Map* of *valueFrom* environment variables to set in the container.
	EnvFrom map[string]corev1.EnvVarSource `json:"envFrom,omitempty"`
	// *Map* of volumes that should be mounted by the container.
	Volumes map[string]corev1.VolumeSource `json:"volumes,omitempty"`
	// *Map* of pod volumes to mount into the container's filesystem.
	VolumeMounts map[string]corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

//+kubebuilder:object:root=true

// InstallValuesConfig is the Schema for the installvaluesconfigs API
type InstallValuesConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	InstallValues     `json:",inline"`
}

//+kubebuilder:object:root=true

// InstallValuesConfigList contains a list of InstallValuesConfig
type InstallValuesConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InstallValuesConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InstallValuesConfig{}, &InstallValuesConfigList{})
}
