package version

import (
	"fmt"
	"testing"

	"cuelang.org/go/cue/errors"
	"github.com/ghodss/yaml"
)

func TestVersions(t *testing.T) {
	versions, err := Load()
	if err != nil {
		for _, e := range errors.Errors(err) {
			t.Error(e)
		}
		t.Fatal()
	}

	for name, version := range versions {
		if version.cue.Err(); err != nil {
			for _, e := range errors.Errors(err) {
				t.Error(e)
			}
			t.Fatal()
		}

		t.Run(name, func(t *testing.T) {
			t.Run("manifests", func(t *testing.T) {
				// fmt.Println(version.cue.LookupPath(cue.ParsePath("edge")))
				// TODO: Check values in manifests
				// manifests := version.Manifests()
				// y, _ := yaml.Marshal(manifests)
				// fmt.Println(string(y))
			})

			t.Run("sidecar", func(t *testing.T) {
				// TODO: Check values in sidecar
				// 	sidecar := version.Sidecar()
				// y, _ := yaml.Marshal(sidecar)
				// fmt.Println(string(y))
			})

			for _, tc := range []struct {
				name           string
				options        []InstallOption
				checkManifests func([]ManifestGroup) error
				checkSidecar   func(Sidecar) error
			}{
				{
					name:    "Namespace option",
					options: []InstallOption{Namespace("ns")},
					checkManifests: func(manifests []ManifestGroup) error {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
						return nil
					},
					checkSidecar: func(sidecar Sidecar) error {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
						return nil
					},
				},
				{
					name:    "SPIRE option",
					options: []InstallOption{SPIRE},
					checkManifests: func(manifests []ManifestGroup) error {
						// unimplemented
						y, _ := yaml.Marshal(manifests)
						fmt.Println(string(y))
						return nil
					},
					checkSidecar: func(sidecar Sidecar) error {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
						return nil
					},
				},
				{
					name:    "Redis internal option",
					options: []InstallOption{Namespace("ns"), Redis(nil)},
					checkManifests: func(manifests []ManifestGroup) error {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
						return nil
					},
					checkSidecar: func(sidecar Sidecar) error {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
						return nil
					},
				},
				{
					name:    "Redis external option",
					options: []InstallOption{Redis(&ExternalRedisConfig{URL: "redis://:pass@extserver:6379/2"})},
					checkManifests: func(manifests []ManifestGroup) error {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
						return nil
					},
					checkSidecar: func(sidecar Sidecar) error {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
						return nil
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					vCopy := version.Copy()
					vCopy.Apply(tc.options...)
					if vCopy.cue.Err(); err != nil {
						for _, e := range errors.Errors(err) {
							t.Error(e)
						}
						t.Fatal()
					}

					t.Run("manifests", func(t *testing.T) {
						manifests := vCopy.Manifests()
						if err := tc.checkManifests(manifests); err != nil {
							t.Fatal(err)
						}
					})

					t.Run("sidecar", func(t *testing.T) {
						sidecar := vCopy.Sidecar()
						if err := tc.checkSidecar(sidecar); err != nil {
							t.Fatal(err)
						}
					})
				})
			}
		})
	}
}
