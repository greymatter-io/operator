package v1alpha1

import (
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

func TestWithSPIRE(t *testing.T) {
	sv := &SystemValues{}
	sv.Overlay(WithSPIRE)
	if _, ok := sv.Proxy.Volumes["spire-socket"]; !ok {
		t.Fatal("expected to find 'spire-socket' volume in Proxy")
	}
	if _, ok := sv.Proxy.VolumeMounts["spire-socket"]; !ok {
		t.Fatal("expected to find 'spire-socket' volumeMount in Proxy")
	}
	if _, ok := sv.Proxy.Env["SPIRE_PATH"]; !ok {
		t.Fatal("expected to find 'SPIRE_PATH' env in Proxy")
	}

	y, err := yaml.Marshal(sv.Proxy)
	if err != nil {
		t.Fatal(err)
	}

	if count := strings.Count(string(y), "SPIRE_PATH: /run/spire/socket/agent.sock"); count != 1 {
		t.Error("did not find substring '/run/spire/socket/agent.sock' in Proxy YAML")
	}
}
