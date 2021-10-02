package version

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"
)

func TestInstallConfigs(t *testing.T) {
	versions, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	for name, version := range versions {
		t.Run(name, func(t *testing.T) {
			values := version.InstallConfigs()
			y, _ := yaml.Marshal(values)
			fmt.Println(string(y))
		})
	}
}
