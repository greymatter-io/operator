package v1alpha1

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

//go:embed fixture.yaml
var fixture string

func TestWithSPIRE(t *testing.T) {
	installValues := loadFixture()
	installValues.Overlay(WithSPIRE)

	t.Run("applies values to Proxy", func(t *testing.T) {
		if _, ok := installValues.Proxy.Volumes["spire-socket"]; !ok {
			t.Error("expected to find 'spire-socket' volume in Proxy")
		}
		if _, ok := installValues.Proxy.VolumeMounts["spire-socket"]; !ok {
			t.Error("expected to find 'spire-socket' volumeMount in Proxy")
		}
		if _, ok := installValues.Proxy.Env["SPIRE_PATH"]; !ok {
			t.Error("expected to find 'SPIRE_PATH' env in Proxy")
		}
	})

	t.Run("overlays are marshalable into YAML", func(t *testing.T) {
		y, err := yaml.Marshal(installValues.Proxy)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(y), "SPIRE_PATH: /run/spire/socket/agent.sock") {
			t.Error("did not find substring 'SPIRE_PATH: /run/spire/socket/agent.sock' in Proxy YAML")
		}
	})
}

func TestWithRedis(t *testing.T) {
	installValues := loadFixture()
	installValues.Overlay(WithRedis("host", "port")) // TODO
}

func loadFixture() *InstallValues {
	installValues := &InstallValuesConfig{}
	yaml.Unmarshal([]byte(fixture), installValues)
	return &installValues.InstallValues
}
