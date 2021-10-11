// Package webhooks exposes functions called from admission webhook handlers.
package webhooks

import (
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/version"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger = ctrl.Log.WithName("pkg.webhooks")
)

type inst interface {
	ApplyMesh(*v1alpha1.Mesh, bool)
	RemoveMesh(string)
	IsMeshMember(string) bool
	Sidecar(string, string) (version.Sidecar, bool)
}

func Register(mgr ctrl.Manager, i inst) {
	mgr.GetWebhookServer().Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{inst: i}})
	mgr.GetWebhookServer().Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{inst: i}})
	mgr.GetWebhookServer().Register("/mutate-workload", &admission.Webhook{Handler: &workloadDefaulter{inst: i}})
}
