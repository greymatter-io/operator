// Package webhooks exposes functions called from admission webhook handlers.
package webhooks

import (
	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger = ctrl.Log.WithName("webhooks")
)

func Register(mgr ctrl.Manager, i *installer.Installer, c *cli.CLI) {
	mgr.GetWebhookServer().Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{Installer: i}})
	mgr.GetWebhookServer().Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{Installer: i, CLI: c}})
	mgr.GetWebhookServer().Register("/mutate-workload", &admission.Webhook{Handler: &workloadDefaulter{Installer: i, CLI: c}})
}
