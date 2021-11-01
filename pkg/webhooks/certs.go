package webhooks

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig"
)

// NOTE: This code does not run, but is retained here for reference and to use later for generating certs.
//lint:ignore U1000 save for reference
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
