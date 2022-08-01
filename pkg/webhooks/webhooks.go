package webhooks

import (
	"context"
	"os"
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	"github.com/greymatter-io/operator/pkg/gmapi"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	"github.com/greymatter-io/operator/pkg/mesh_install"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger = ctrl.Log.WithName("webhooks")
)

const (
	defaultCSRHost = "gm-webhook.gm-operator.svc"
)

type Loader struct {
	client.Client
	*mesh_install.Installer
	*gmapi.CLI
	*cfsslsrv.CFSSLServer
	getServer func() *webhook.Server
	caBundle  []byte
	cert      []byte
	key       []byte
}

func New(
	cl *client.Client,
	i *mesh_install.Installer,
	c *gmapi.CLI,
	cs *cfsslsrv.CFSSLServer,
	get func() *webhook.Server) (*Loader, error) {

	wl := &Loader{Client: *cl, Installer: i, CLI: c, CFSSLServer: cs, getServer: get}

	if !i.Config.GenerateWebhookCerts {
		logger.Info("webhook server cert generation disabled; expecting webhook server certs to be mounted from external source")
		return wl, nil
	}

	var err error

	wl.caBundle = wl.GetRootCA()

	wl.cert, wl.key, err = wl.RequestCert(csr.CertificateRequest{
		CN:         "admission",
		Hosts:      []string{defaultCSRHost},
		KeyRequest: &csr.KeyRequest{A: "ecdsa", S: 256},
	})
	if err != nil {
		logger.Error(err, "failed to retrieve certs for webhook server")
		return nil, err
	}

	logger.Info("Retrieved signed certs from CFSSL server")

	return wl, nil
}

func (wl *Loader) Start(ctx context.Context) error {

	// If webhook cert generation is disabled, just register the webhook handlers and exit
	if !wl.Config.GenerateWebhookCerts {
		wl.register()
		return nil
	}

	// Patch the opaque secret with our previously loaded signed certs
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gm-webhook-cert",
			Namespace: "gm-operator",
		},
	}
	k8sapi.Apply(&wl.Client, secret, nil, k8sapi.MkPatchAction(func(obj client.Object) client.Object {
		s := obj.(*corev1.Secret)
		if s.StringData == nil {
			s.StringData = make(map[string]string)
		}
		s.StringData["tls.crt"] = string(wl.cert)
		s.StringData["tls.key"] = string(wl.key)
		return s
	}))

	// Patch the mutatingwebhookconfiguration with our previously loaded cabundle
	mwc := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-mutate-config"},
	}
	k8sapi.Apply(&wl.Client, mwc, nil, k8sapi.MkPatchAction(func(obj client.Object) client.Object {
		m := obj.(*admissionregistrationv1.MutatingWebhookConfiguration)
		for i := range m.Webhooks {
			m.Webhooks[i].ClientConfig.CABundle = wl.caBundle
		}
		return m
	}))

	// Since we've just patched our webhook secret, check the mounted file for changes.
	// This lets us wait for the certwatcher to identify cert "rotation" before registering webhooks.
	logger.Info("Waiting for certwatcher to detect new webhook TLS certs")
	start := time.Now()
	var byteCount int64
	for byteCount == 0 {
		fileInfo, _ := os.Stat("/tmp/k8s-webhook-server/serving-certs/tls.crt")
		byteCount = fileInfo.Size()
		time.Sleep(time.Second * 2)
	}
	logger.Info("New webhook TLS certs detected", "Elapsed", time.Since(start).String())
	wl.register()

	return nil
}

func (wl *Loader) register() {
	server := wl.getServer()
	server.Register("/mutate-workload", &admission.Webhook{Handler: &workloadDefaulter{Installer: wl.Installer, CLI: wl.CLI}})
}
