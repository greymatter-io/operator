package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	getServer func() *webhook.Server
	caBundle  []byte
	cert      string
	key       string
}

func New(cl client.Client, i *installer.Installer, c *cli.CLI, get func() *webhook.Server) *Loader {
	wl := &Loader{Client: cl, Installer: i, CLI: c, getServer: get}
	return wl
}

func (wl *Loader) register() {
	server := wl.getServer()
	server.Register("/mutate-mesh", &admission.Webhook{Handler: &meshDefaulter{Installer: wl.Installer}})
	server.Register("/validate-mesh", &admission.Webhook{Handler: &meshValidator{Installer: wl.Installer, Client: wl.Client}})
	server.Register("/mutate-workload", &admission.Webhook{Handler: &workloadDefaulter{Installer: wl.Installer, CLI: wl.CLI}})
}

func (wl *Loader) LoadCerts() error {
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

func (wl *Loader) Start(ctx context.Context) error {
	secret := &corev1.Secret{}
	key := client.ObjectKey{Name: "gm-webhook-server-cert", Namespace: "gm-operator"}
	if err := wl.Client.Get(context.TODO(), key, secret); err != nil {
		logger.Error(err, "get", "result", "fail", "Secret", key)
		return err
	}

	patch := client.MergeFrom(secret.DeepCopy())
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	secret.StringData["tls.crt"] = wl.cert
	secret.StringData["tls.key"] = wl.key
	if err := wl.Client.Patch(context.TODO(), secret, patch); err != nil {
		logger.Error(err, "patch", "result", "fail", "Secret", key)
		return err
	}

	mwc := &admissionregistrationv1.MutatingWebhookConfiguration{}
	key = client.ObjectKey{Name: "gm-mutating-webhook-configuration"}
	if err := wl.Client.Get(context.TODO(), key, mwc); err != nil {
		logger.Error(err, "get", "result", "fail", "MutatingWebhookConfiguration", key)
		return err
	}

	patch = client.MergeFrom(mwc.DeepCopy())
	for i := range mwc.Webhooks {
		mwc.Webhooks[i].ClientConfig.CABundle = wl.caBundle
	}
	if err := wl.Client.Patch(context.TODO(), mwc, patch); err != nil {
		logger.Error(err, "patch", "result", "fail", "MutatingWebhookConfiguration", key)
		return err
	}

	vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	key = client.ObjectKey{Name: "gm-validating-webhook-configuration"}
	if err := wl.Client.Get(context.TODO(), key, vwc); err != nil {
		logger.Error(err, "get", "result", "fail", "ValidatingWebhookConfiguration", key)
		return err
	}
	patch = client.MergeFrom(vwc.DeepCopy())
	for i := range vwc.Webhooks {
		vwc.Webhooks[i].ClientConfig.CABundle = wl.caBundle
	}
	if err := wl.Client.Patch(context.TODO(), vwc, patch); err != nil {
		logger.Error(err, "patch", "result", "fail", "ValidatingWebhookConfiguration", key)
		return err
	}

	wl.register()

	return nil
}

type respStruct struct {
	Result struct {
		Cert string `json:"certificate"`
		Key  string `json:"private_key"`
	} `json:"result"`
}

// TODO: Clean this up
func genCerts() ([]string, error) {
	client := http.Client{Timeout: time.Second}

	// Request root CA without specifying a signer label, since we expect only one root (for now; maybe support multi-root later)
	req, err := http.NewRequest("POST", "http://127.0.0.1:8888/api/v1/cfssl/info", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := &respStruct{}
	if err := json.Unmarshal(respBody, body); err != nil {
		return nil, err
	}

	caBundle := body.Result.Cert

	// Request a new signed cert for our webhook server endpoints
	csr := bytes.NewReader([]byte(`{
		"request": {
			"CN":"admission",
			"hosts":["gm-webhook-service.gm-operator.svc"],
			"key":{"algo":"rsa","size":2048},
			"names": [{"C":"US","ST":"VA","O":"Grey Matter"}]
		},
		"profile": "server"
	}`))
	req, err = http.NewRequest("POST", "http://127.0.0.1:8888/api/v1/cfssl/newcert", csr)
	if err != nil {
		return nil, err
	}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body = &respStruct{}
	if err := json.Unmarshal(respBody, body); err != nil {
		return nil, err
	}
	signed := body.Result

	return []string{
		caBundle,
		string(signed.Cert),
		string(signed.Key),
	}, nil
}
