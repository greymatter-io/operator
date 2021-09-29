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
	installValues.With(SPIRE)

	t.Run("applies values to Proxy", func(t *testing.T) {
		if _, ok := installValues.Proxy.Volumes["spire-socket"]; !ok {
			t.Error("expected to find 'spire-socket' volume in Proxy")
		}
		if _, ok := installValues.Proxy.VolumeMounts["spire-socket"]; !ok {
			t.Error("expected to find 'spire-socket' volumeMount in Proxy")
		}
		if _, ok := installValues.Proxy.Envs["SPIRE_PATH"]; !ok {
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

func TestRedis(t *testing.T) {
	installValues := loadFixture()

	t.Run("Test install values are loaded are modified (with redis config)", func(t *testing.T) {
		if installValues.Redis.Envs["REDIS_PASSWORD"] != "" {
			t.Errorf("Expected to find REDIS_PASSWORD is not empty")
		}

	})

}

func loadFixture() *InstallValues {
	cfg := &InstallationConfig{}
	yaml.Unmarshal([]byte(fixture), cfg)
	return &cfg.InstallValues
}
