package cfsslsrv

import (
	"crypto/x509"
	"testing"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCFSSL(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	cs, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	roots, err := helpers.ParseCertificatesPEM(cs.GetRootCA())
	if err != nil {
		t.Fatal("invalid CA PEM block", err)
	}
	if _, err := helpers.ParsePrivateKeyPEM(cs.caKey); err != nil {
		t.Fatal("invalid CA key PEM block", err)
	}

	if err := cs.Start(); err != nil {
		t.Fatal(err)
	}

	ca, caKey, err := cs.RequestIntermediateCA(csr.CertificateRequest{
		CN:         "Grey Matter Intermediate CA",
		KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
		Names: []csr.Name{
			{C: "US", ST: "VA", L: "Alexandria", O: "Grey Matter"},
		},
		Hosts: []string{"greymatter.io"},
	})
	if err != nil {
		t.Fatal(err)
	}

	intermediates, err := helpers.ParseCertificatesPEM(ca)
	if err != nil {
		t.Fatal("invalid intermediate CA PEM block", err)
	}
	if _, err := helpers.ParsePrivateKeyPEM(caKey); err != nil {
		t.Fatal("invalid intermediate CA key PEM block", err)
	}

	root := roots[0]
	intermediate := intermediates[0]

	r := x509.NewCertPool()
	r.AddCert(root)

	if _, err := intermediate.Verify(x509.VerifyOptions{Roots: r}); err != nil {
		t.Fatal("failed to verify intermediate", err)
	}

	cert, key, err := cs.RequestCert(csr.CertificateRequest{
		CN:         "dummy",
		Hosts:      []string{"dummy.svc"},
		KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
	})
	if err != nil {
		t.Fatal(err)
	}

	certs, err := helpers.ParseCertificatesPEM(cert)
	if err != nil {
		t.Fatal("invalid cert PEM block", err)
	}
	if _, err := helpers.ParsePrivateKeyPEM(key); err != nil {
		t.Fatal("invalid cert key PEM block", err)
	}

	c := certs[0]
	if _, err := c.Verify(x509.VerifyOptions{
		Roots: r,
	}); err != nil {
		t.Fatal("failed to verify cert", err)
	}
}
