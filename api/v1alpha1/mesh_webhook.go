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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

//+kubebuilder:webhook:path=/mutate-mesh,mutating=true,failurePolicy=fail,sideEffects=None,groups=greymatter.io,resources=meshes,verbs=create;update,versions=v1alpha1,name=mutate-mesh-webhook.greymatter.io,admissionReviewVersions={v1,v1beta1}
//+kubebuilder:webhook:path=/validate-mesh,mutating=false,failurePolicy=fail,sideEffects=None,groups=greymatter.io,resources=meshes,verbs=create;update;delete,versions=v1alpha1,name=validate-mesh-webhook.greymatter.io,admissionReviewVersions={v1,v1beta1}

// Installer callbacks declared in the Webhook setup function which will be called by each Webhook event
// These are initialized here as no-ops for testing.
var (
	applyInstall   = func(*Mesh, bool) {}
	applyUninstall = func(string) {}
)

func (r *Mesh) SetupWebhooks(mgr ctrl.Manager, install func(*Mesh, bool), uninstall func(string)) error {
	applyInstall = install
	applyUninstall = uninstall

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

///+kubebuilder:webhook:path=/mutate-greymatter-io-v1alpha1-mesh,mutating=true,failurePolicy=fail,sideEffects=None,groups=greymatter.io,resources=meshes,verbs=create;update,versions=v1alpha1,name=mmesh.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Mesh{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Mesh) Default() {
	fmt.Printf("Default: %#v\n", r.Spec)
}

///+kubebuilder:webhook:path=/validate-greymatter-io-v1alpha1-mesh,mutating=false,failurePolicy=fail,sideEffects=None,groups=greymatter.io,resources=meshes,verbs=create;update;delete,versions=v1alpha1,name=vmesh.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Mesh{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Mesh) ValidateCreate() error {
	fmt.Printf("ValidateCreate: %#v\n", r.Spec)
	go applyInstall(r, true)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Mesh) ValidateUpdate(old runtime.Object) error {
	fmt.Printf("ValidateUpdate: %#v\n", r.Spec)
	// TODO: Identify watch namespaces we need to delete imagePullSecrets for.
	// Maybe "applyInstall" needs to take both the old and new mesh to do a diff.
	go applyInstall(r, false)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Mesh) ValidateDelete() error {
	fmt.Printf("ValidateDelete: %#v\n", r.Spec)
	go applyUninstall(r.Name)
	return nil
}
