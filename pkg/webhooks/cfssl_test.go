package webhooks

import (
	"testing"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCFSSL(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	cs, err := NewCFSSLServer(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := helpers.ParseCertificatesPEM(cs.GetCABundle()); err != nil {
		t.Fatal("invalid CA PEM block", err)
	}
	if _, err := helpers.ParsePrivateKeyPEM(cs.caKey); err != nil {
		t.Fatal("invalid CA key PEM block", err)
	}

	if err := cs.Start(); err != nil {
		t.Fatal(err)
	}

	cert, key, err := cs.RequestCert(csr.CertificateRequest{
		CN:         "dummy",
		Hosts:      []string{"some.svc"},
		KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := helpers.ParseCertificatesPEM(cert); err != nil {
		t.Fatal("invalid CA PEM block", err)
	}
	if _, err := helpers.ParsePrivateKeyPEM(key); err != nil {
		t.Fatal("invalid CA key PEM block", err)
	}
}
