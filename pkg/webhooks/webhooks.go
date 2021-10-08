// Package webhooks exposes functions called from admission webhook handlers.
package webhooks

import (
	"github.com/greymatter-io/operator/pkg/installer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func Register(mgr ctrl.Manager, i *installer.Installer) {
	mgr.GetWebhookServer().Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{inst: i}})
	mgr.GetWebhookServer().Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{inst: i}})
}
