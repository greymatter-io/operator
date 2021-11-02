package webhooks

import (
	"strings"
	"testing"
)

func TestGenCerts(t *testing.T) {
	output, err := genCerts("")
	if err != nil {
		t.Fatal(err)
	}

	if len(output) != 3 {
		t.Fatal("output does not have len == 3")
	}

	if !strings.HasPrefix(output[0], "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUR") {
		t.Error("output[0] is not a base 64 encoded CA")
	}

	if !strings.HasPrefix(output[1], "-----BEGIN CERTIFICATE-----") {
		t.Error("output[1] is not a certificate")
	}

	if !strings.HasPrefix(output[2], "-----BEGIN RSA PRIVATE KEY-----") {
		t.Error("output[2] is not a RSA private key")
	}
}
