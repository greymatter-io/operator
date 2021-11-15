package webhooks

import (
	"context"
	"os"
	"time"

	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	logger = ctrl.Log.WithName("webhooks")
)

type Loader struct {
	client.Client
	*installer.Installer
	*cli.CLI
	getServer      func() *webhook.Server
	disableCertGen bool
	caBundle       []byte
	cert           string
	key            string
}

func New(cl client.Client, i *installer.Installer, c *cli.CLI, get func() *webhook.Server, disableCertGen bool) (*Loader, error) {
	wl := &Loader{Client: cl, Installer: i, CLI: c, getServer: get, disableCertGen: disableCertGen}

	if !wl.disableCertGen {
		// Initialize and launch CFSSL server. This will eventualy move out of the webhooks package,
		// but can stay here for now since we're just using it to issue our webhook certs.
		caBundle, err := serveCFSSL()
		if err != nil {
			logger.Error(err, "Failed to launch CFSSL server")
			return nil, err
		}
		wl.caBundle = caBundle

		certs, err := requestWebhookCerts()
		if err != nil {
			logger.Error(err, "failed to retrieve webhook certs")
			return nil, err
		}

		wl.cert = certs[0]
		wl.key = certs[1]
	}

	return wl, nil
}

func (wl *Loader) Start(ctx context.Context) error {

	// If webhook cert generation is disabled, just register the webhook handlers and exit
	if wl.disableCertGen {
		wl.register()
		return nil
	}

	// Patch the opaque secret with our previously loaded signed certs
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gm-controller-manager-service-cert",
			Namespace: "gm-operator",
		},
	}
	patch := func(obj client.Object) client.Object {
		s := obj.(*corev1.Secret)
		if s.StringData == nil {
			s.StringData = make(map[string]string)
		}
		s.StringData["tls.crt"] = wl.cert
		s.StringData["tls.key"] = wl.key
		return s
	}
	if err := applyPatch(wl.Client, secret, patch); err != nil {
		return err
	}

	// Patch the mutatingwebhookconfiguration with our previously loaded cabundle
	mwc := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-mutating-webhook-configuration"},
	}
	patch = func(obj client.Object) client.Object {
		m := obj.(*admissionregistrationv1.MutatingWebhookConfiguration)
		for i := range m.Webhooks {
			m.Webhooks[i].ClientConfig.CABundle = wl.caBundle
		}
		return m
	}
	if err := applyPatch(wl.Client, mwc, patch); err != nil {
		return err
	}

	// Patch the validatingwebhookconfiguration with our previously loaded cabundle
	vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-validating-webhook-configuration"},
	}
	patch = func(obj client.Object) client.Object {
		v := obj.(*admissionregistrationv1.ValidatingWebhookConfiguration)
		for i := range v.Webhooks {
			v.Webhooks[i].ClientConfig.CABundle = wl.caBundle
		}
		return v
	}
	if err := applyPatch(wl.Client, vwc, patch); err != nil {
		return err
	}

	// Since we've just patched our webhook secret, check the mounted file for changes.
	// This lets us wait for the certwatcher to identify cert "rotation" before registering webhooks.
	logger.Info("Waiting for certwatcher to detect TLS certificate update")
	var byteCount int64
	for byteCount == 0 {
		fileInfo, _ := os.Stat("/tmp/k8s-webhook-server/serving-certs/tls.crt")
		byteCount = fileInfo.Size()
		time.Sleep(time.Second * 3)
	}

	wl.register()

	return nil
}

func applyPatch(c client.Client, obj client.Object, patch func(client.Object) client.Object) error {
	var kind string
	if gvk, err := apiutil.GVKForObject(obj.(runtime.Object), c.Scheme()); err != nil {
		kind = "Object"
	} else {
		kind = gvk.Kind
	}

	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(context.TODO(), key, obj); err != nil {
		logger.Error(err, "get", "result", "fail", kind, key)
	}

	mp := client.MergeFrom(obj.DeepCopyObject().(client.Object))
	obj = patch(obj)
	if err := c.Patch(context.TODO(), obj, mp); err != nil {
		logger.Error(err, "patch", "result", "fail", kind, key)
	}

	logger.Info("patch", "result", "success", kind, key)

	return nil
}

func (wl *Loader) register() {
	server := wl.getServer()
	server.Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{Installer: wl.Installer}})
	server.Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{Installer: wl.Installer, Client: wl.Client}})
	server.Register("/mutate-workload", &admission.Webhook{Handler: &workloadDefaulter{Installer: wl.Installer, CLI: wl.CLI}})
}
