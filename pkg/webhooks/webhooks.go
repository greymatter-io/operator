// Package webhooks exposes functions called from admission webhook handlers.
package webhooks

import (
	"github.com/greymatter-io/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type inst interface {
	ApplyMesh(*v1alpha1.Mesh, bool)
	RemoveMesh(string)
	InjectSidecar([]corev1.Container, string, string) []corev1.Container
}

func Register(mgr ctrl.Manager, i inst) {
	mgr.GetWebhookServer().Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{inst: i}})
	mgr.GetWebhookServer().Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{inst: i}})
}
