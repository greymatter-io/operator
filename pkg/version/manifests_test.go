package version

import (
	"fmt"
	"testing"

	"cuelang.org/go/cue/errors"
	"github.com/ghodss/yaml"
)

func TestManifests(t *testing.T) {
	versions, err := Load()
	if err != nil {
		for _, e := range errors.Errors(err) {
			t.Error(e)
		}
		t.Fatal()
	}

	for name, version := range versions {
		t.Run(name, func(t *testing.T) {
			// t.Run("cue", func(t *testing.T) {
			// 	fmt.Println(version.cue.LookupPath(cue.ParsePath("manifests")))
			// })

			t.Run("manifests", func(t *testing.T) {
				manifests := version.Manifests()
				for _, group := range manifests {
					y, _ := yaml.Marshal(group)
					fmt.Println(string(y))
				}
			})

			// t.Run("sidecar", func(t *testing.T) {
			// 	sidecar := version.Sidecar()
			// 	y, _ := yaml.Marshal(sidecar)
			// 	fmt.Println(string(y))
			// })
		})
	}
}
