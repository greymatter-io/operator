package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const certMountPath = "/tmp/k8s-webhook-server/serving-certs"

type CertsLoader struct {
	client.Client
	caBundle []byte
	cert     []byte
	key      []byte
}

func (cl *CertsLoader) Load() error {
	certs, err := genCerts()
	if err != nil {
		logger.Error(err, "failed to generate certs")
		return err
	}

	if err = os.MkdirAll(fmt.Sprintf("%s/", certMountPath), 0666); err != nil {
		logger.Error(err, "failed to create cert directory")
		return err
	}
	if err = writeFile(fmt.Sprintf("%s/tls.crt", certMountPath), certs[1]); err != nil {
		logger.Error(err, "failed to write server cert to file")
		return err
	}
	if err = writeFile(fmt.Sprintf("%s/tls.key", certMountPath), certs[2]); err != nil {
		logger.Error(err, "failed to write server key to file")
		return err
	}

	cl.caBundle = []byte(certs[0])
	return nil
}

func writeFile(filepath string, data string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write([]byte(data)); err != nil {
		return err
	}
	return nil
}

func (cl *CertsLoader) Start(ctx context.Context) error {
	mwc := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gm-mutating-webhook-configuration",
		},
	}
	key := client.ObjectKeyFromObject(mwc)
	if err := cl.Client.Get(context.TODO(), key, mwc); err != nil {
		logger.Error(err, "get", "result", "fail", "MutatingWebhookConfiguration", key)
		return err
	}

	patch := client.MergeFrom(mwc.DeepCopy())
	for i := range mwc.Webhooks {
		mwc.Webhooks[i].ClientConfig.CABundle = cl.caBundle
	}
	if err := cl.Client.Patch(context.TODO(), mwc, patch); err != nil {
		logger.Error(err, "patch", "result", "fail", "MutatingWebhookConfiguration", key)
		return err
	}

	vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gm-validating-webhook-configuration",
		},
	}
	key = client.ObjectKeyFromObject(vwc)
	if err := cl.Client.Get(context.TODO(), key, vwc); err != nil {
		logger.Error(err, "get", "result", "fail", "ValidatingWebhookConfiguration", key)
		return err
	}
	patch = client.MergeFrom(vwc.DeepCopy())
	for i := range vwc.Webhooks {
		vwc.Webhooks[i].ClientConfig.CABundle = cl.caBundle
	}
	if err := cl.Client.Patch(context.TODO(), vwc, patch); err != nil {
		logger.Error(err, "patch", "result", "fail", "ValidatingWebhookConfiguration", key)
		return err
	}

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
