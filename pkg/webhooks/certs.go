package webhooks

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/Masterminds/sprig"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InjectCA parses pre-existing registered webhooks and injects generated self-signed CA bundles
// so remote vanilla k8s deployments are possible without special CA operator injection.
func InjectCA(c client.Client) error {
	caBundle, err := storeAndReturnCertBundle("/tmp/k8s-webhook-server/serving-certs")
	if err != nil {
		logger.Error(err, "failed to create self-signed ca bundle")
		return err
	}

	mwc := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gm-mutating-webhook-configuration",
			Namespace: "gm-operator",
		},
	}
	if err = c.Get(context.TODO(), client.ObjectKeyFromObject(mwc), mwc); err != nil {
		logger.Error(err, "get", "result", "fail", "Namespace", "gm-operator", "MutatingWebhookConfiguration", "gm-mutating-webhook-configuration")
		return err
	}
	for i := range mwc.Webhooks {
		mwc.Webhooks[i].ClientConfig.CABundle = caBundle
	}
	if err := c.Update(context.TODO(), mwc); err != nil {
		logger.Error(err, "update", "result", "fail", "Namespace", "gm-operator", "MutatingWebhookConfiguration", "gm-mutating-webhook-configuration")
		return err
	}
	logger.Info("update", "result", "success", "Namespace", "gm-operator", "MutatingWebhookConfiguration", "gm-mutating-webhook-configuration")

	vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gm-validating-webhook-configuration",
			Namespace: "gm-operator",
		},
	}
	if err = c.Get(context.TODO(), client.ObjectKeyFromObject(vwc), vwc); err != nil {
		logger.Error(err, "get", "result", "fail", "Namespace", "gm-operator", "ValidatingWebhookConfiguration", "gm-validating-webhook-configuration")
		return err
	}
	for i := range vwc.Webhooks {
		vwc.Webhooks[i].ClientConfig.CABundle = caBundle
	}
	if err := c.Update(context.TODO(), vwc); err != nil {
		logger.Error(err, "update", "result", "fail", "Namespace", "gm-operator", "MutatingWebhookConfiguration", "gm-validating-webhook-configuration")
		return err
	}
	logger.Info("update", "result", "success", "Namespace", "gm-operator", "ValidatingWebhookConfiguration", "gm-validating-webhook-configuration")

	return nil
}

func storeAndReturnCertBundle(certMountPath string) ([]byte, error) {
	certs, err := genCerts(certMountPath)
	if err != nil {
		return []byte{}, err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/", certMountPath), 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}
	err = writeStringToFile(fmt.Sprintf("%s/tls.crt", certMountPath), certs[1])
	if err != nil {
		return nil, fmt.Errorf("failed to write server cert to file: %w", err)
	}
	err = writeStringToFile(fmt.Sprintf("%s/tls.key", certMountPath), certs[2])
	if err != nil {
		return nil, fmt.Errorf("failed to write server key to file: %w", err)
	}

	return []byte(certs[0]), nil
}

func writeStringToFile(filepath string, data string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(data))
	return err
}

func genCerts(certMountPath string) ([]string, error) {
	tmpl, err := template.New("certs").Funcs(sprig.FuncMap()).Parse(`
		{{- $ca := genCA "greymatter.io" 3650 -}}
		{{- $server := genSignedCert "gm-webhook-service.gm-operator.svc" (list) (list "gm-webhook-service.gm-operator.svc.cluster.local") 3650 $ca -}}
		{{- $ca.Cert | b64enc -}}~~~
		{{- $server.Cert -}}~~~
		{{- $server.Key -}}
	`)
	if err != nil {
		return []string{}, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		return []string{}, err
	}

	var trimmed []string
	for _, s := range strings.Split(buf.String(), "~~~") {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}

	return trimmed, nil
}
