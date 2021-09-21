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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OperatorConfig enables defining configuration settings for a Grey Matter Operator.
type OperatorConfig struct {
	metav1.TypeMeta                        `json:",inline"`
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// TODO: Define settings that we need to configure for each Operator
	// and create logical groupings for them.
	ImagePullSecretName string `json:"imagePullSecretName"`
}

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
}
