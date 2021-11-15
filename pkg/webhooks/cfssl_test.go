package webhooks

import (
	"strings"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCFSSL(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	ca, err := serveCFSSL()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(ca), "-----BEGIN CERTIFICATE-----") {
		t.Error("ca is not a valid certificate")
	}

	certs, err := requestWebhookCerts()
	if err != nil {
		t.Fatal(err)
	}

	if len(certs) != 2 {
		t.Fatal("certs does not have len == 2")
	}

	cert, key := certs[0], certs[1]

	if !strings.HasPrefix(cert, "-----BEGIN CERTIFICATE-----") {
		t.Error("certs[0] is not a certificate")
	}

	if !strings.HasPrefix(key, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Error("certs[1] is not a RSA private key")
	}
}
