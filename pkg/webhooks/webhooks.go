// Package webhooks exposes functions called from admission webhook handlers.
package webhooks

import (
	"context"
	"fmt"

	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger        = ctrl.Log.WithName("webhooks")
	knownWebhooks = []string{"gm-validating-webhook-configuration", "gm-mutating-webhook-configuration"}
)

func Register(mgr ctrl.Manager, i *installer.Installer, c *cli.CLI, cc client.Client) {
	mgr.GetWebhookServer().Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{Installer: i}})
	mgr.GetWebhookServer().Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{Installer: i, Client: cc}})
	mgr.GetWebhookServer().Register("/mutate-workload", &admission.Webhook{Handler: &workloadDefaulter{Installer: i, CLI: c}})
}

// InjectCA parses pre-existing registered webhooks and injects generated self-signed CA bundles
// so remote vanilla k8s deployments are possible without special CA operator injection.
func InjectCA(c client.Client) error {
	caBundle, err := CreateCertBundle("gm-operator", "gm-webhook-service", "/tmp/k8s-webhook-server/serving-certs")
	if err != nil {
		return fmt.Errorf("failed to create self-signed ca bundle: %w", err)
	}

	// Retrieve the currently registered webhooks
	cfg := admissionregistrationv1.MutatingWebhookConfiguration{}
	for _, webhook := range knownWebhooks {
		if err := c.Get(context.Background(), client.ObjectKey{
			Namespace: "gm-operator",
			Name:      webhook,
		}, &cfg); err == nil {
			for _, webhook := range cfg.Webhooks {
				if webhook.ClientConfig.CABundle == nil {
					webhook.ClientConfig.CABundle = caBundle
				}
			}

			// Return the updated webhooks back to the API server
			err = c.Update(context.Background(), &cfg)
			if err != nil {
				return fmt.Errorf("failed while updating CA bundles in pre-registered webhooks: %v", err)
			}
		}
	}

	return nil
}
