package webhooks

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// ref: https://www.velotio.com/engineering-blog/managing-tls-certificate-for-kubernetes-admission-webhook
func CreateCertBundle(namespace, service, certMountPath string) ([]byte, error) {
	var caPEM, serverCertPEM, serverPrivKeyPEM bytes.Buffer

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// CA config
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2021),
		Subject:               pkix.Name{Organization: []string{"greymatter.io"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Self signed CA certificate
	b, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// PEM encode CA cert
	if err := pem.Encode(&caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: b,
	}); err != nil {
		return nil, fmt.Errorf("failed to encode CA certificate: %w", err)
	}

	namespacedName := fmt.Sprintf("%s.%s", service, namespace)
	commonName := fmt.Sprintf("%s.svc", namespacedName)
	dnsNames := []string{service, namespacedName, commonName}

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"greymatter.io"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to create server private key: %w", err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &serverPrivKey.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to sign server certificate: %w", err)
	}

	// PEM encode the  server cert and key
	_ = pem.Encode(&serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	_ = pem.Encode(&serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = os.MkdirAll(fmt.Sprintf("%s/", certMountPath), 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}
	err = writeFile(fmt.Sprintf("%s/tls.crt", certMountPath), &serverCertPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to write server cert to file: %w", err)
	}
	err = writeFile(fmt.Sprintf("%s/tls.key", certMountPath), &serverPrivKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to write server key to file: %w", err)
	}

	return caPEM.Bytes(), nil
}

func writeFile(filepath string, data *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data.Bytes())
	if err != nil {
		return err
	}
	return nil
}
