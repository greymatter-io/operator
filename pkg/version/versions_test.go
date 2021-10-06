package version

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestVersions(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	versions, err := loadBaseWithVersions()
	if err != nil {
		logCueErrors(err)
		t.Fatal()
	}

	for name, version := range versions {
		if version.cue.Err(); err != nil {
			logCueErrors(err)
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
				// sidecar := version.Sidecar()
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
					name:    "InstallNamespace option",
					options: []InstallOption{InstallNamespace("ns")},
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
					name:    "SPIRE option",
					options: []InstallOption{SPIRE},
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
					name:    "Redis internal option",
					options: []InstallOption{InstallNamespace("ns"), Redis("")},
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
					options: []InstallOption{Redis("redis://:pass@extserver:6379/2")},
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
					v := version.Copy()
					v.Apply(tc.options...)

					// if tc.name == "InstallNamespace option" {
					// 	fmt.Println(v.cue)
					// }

					if err := v.cue.Err(); err != nil {
						logCueErrors(err)
						t.Fatal()
					}
					t.Run("manifests", func(t *testing.T) {
						if err := tc.checkManifests(v.Manifests()); err != nil {
							t.Fatal(err)
						}
					})
					t.Run("sidecar", func(t *testing.T) {
						if err := tc.checkSidecar(v.Sidecar()); err != nil {
							t.Fatal(err)
						}
					})
				})
			}
		})
	}
}
