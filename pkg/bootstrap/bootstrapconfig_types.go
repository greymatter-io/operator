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

package bootstrap

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// BootstrapConfig specifies global settings for the operator.
// Fields are camelCased rather than snake_cased for compatibility with cfg.ControllerManagerConfigSpec.
type BootstrapConfig struct {
	metav1.TypeMeta                        `json:",inline"`
	cfg.ControllerManagerConfigurationSpec `json:",inline"`
	// By default, the operator uses a local CFSSL API server for signing its webhook TLS certs.
	// Set this to true when deploying to Openshift or in cases where users want to manually sign the webhook TLS certs.
	DisableWebhookCertGeneration bool `json:"disableWebhookCertGeneration"`
	// The name of the cluster ingress (configs.openshift.io/v1.Ingress) in an Openshift environment.
	// This is used to determine the domain of the cluster. Defaults to "cluster" if not set.
	// This is ignored in non-Openshift environments.
	ClusterIngressName string `json:"clusterIngressName"`
}
