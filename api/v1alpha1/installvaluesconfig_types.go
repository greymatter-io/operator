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
)

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
