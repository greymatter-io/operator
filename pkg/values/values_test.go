package values

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

//go:embed fixture.yaml
var fixture string

func TestWithSPIRE(t *testing.T) {
	values := loadFixture()
	values.Apply(SPIRE)

	t.Run("applies values to Proxy", func(t *testing.T) {
		if _, ok := values.Proxy.Volumes["spire-socket"]; !ok {
			t.Error("expected to find 'spire-socket' volume in Proxy")
		}
		if _, ok := values.Proxy.VolumeMounts["spire-socket"]; !ok {
			t.Error("expected to find 'spire-socket' volumeMount in Proxy")
		}
		if _, ok := values.Proxy.Envs["SPIRE_PATH"]; !ok {
			t.Error("expected to find 'SPIRE_PATH' env in Proxy")
		}
	})

	t.Run("overlays are marshalable into YAML", func(t *testing.T) {
		y, err := yaml.Marshal(values.Proxy)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(y), "SPIRE_PATH: /run/spire/socket/agent.sock") {
			t.Error("did not find substring 'SPIRE_PATH: /run/spire/socket/agent.sock' in Proxy YAML")
		}
	})
}

func TestRedis(t *testing.T) {
	values := loadFixture()

	t.Run("updates values with a nil ExternalRedisConfig", func(t *testing.T) {
		values.Apply(Redis(nil, "namespace"))
		if values.Redis.Envs["REDIS_PASSWORD"] == "" {
			t.Errorf("Redis.Envs was not assigned a generated REDIS_PASSWORD")
		}
	})
}

func loadFixture() *Values {
	values := &Values{}
	yaml.Unmarshal([]byte(fixture), values)
	return values
}
