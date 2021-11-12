package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		if err := wl.loadCerts(); err != nil {
		    return nil, err
		}
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

	// Register handlers with the webhook server
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

func (wl *Loader) loadCerts() error {
	certs, err := genCerts()
	if err != nil {
		logger.Error(err, "failed to generate certs")
		return err
	}

	wl.caBundle = []byte(certs[0])
	wl.cert = certs[1]
	wl.key = certs[2]

	return nil
}

type cfsslResp struct {
	Result struct {
		Cert string `json:"certificate"`
		Key  string `json:"private_key"`
	} `json:"result"`
}

func genCerts() ([]string, error) {
	c := http.Client{Timeout: time.Second}

	// Request root CA without specifying a signer label, since we expect only one root (for now; maybe support multi-root later)
	info, err := getCFSSLResp(c, "info", "{}")
	if err != nil {
		return nil, err
	}
	caBundle := info.Result.Cert

	// Request a new signed cert for our webhook server endpoints
	newcert, err := getCFSSLResp(c, "newcert", `{
		"request": {
			"CN":"admission",
			"hosts":["gm-webhook-service.gm-operator.svc"],
			"key":{"algo":"rsa","size":2048},
			"names": [{"C":"US","ST":"VA","O":"Grey Matter"}]
		},
		"profile": "server"
	}`)
	if err != nil {
		return nil, err
	}
	signed := newcert.Result

	return []string{
		caBundle,
		string(signed.Cert),
		string(signed.Key),
	}, nil
}

func getCFSSLResp(c http.Client, path, data string) (*cfsslResp, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:8888/api/v1/cfssl/%s", path), bytes.NewReader([]byte(data)))
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := &cfsslResp{}
	if err := json.Unmarshal(respBody, body); err != nil {
		return nil, err
	}
	return body, nil
}
