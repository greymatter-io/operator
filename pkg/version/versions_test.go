package version

import (
	"testing"

	"cuelang.org/go/cue/errors"
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
		t.Run(name, func(t *testing.T) {
			t.Run("vanilla", func(t *testing.T) {
				t.Run("manifests", func(t *testing.T) {
					// TODO: Check values in manifests
					// manifests := version.Manifests()
				})

				// t.Run("sidecar", func(t *testing.T) {
				// 	sidecar := version.Sidecar()
				// 	y, _ := yaml.Marshal(sidecar)
				// 	fmt.Println(string(y))
				// })
			})

			for _, tc := range []struct {
				name    string
				options []InstallOption
				errors  func([]ManifestGroup) error
			}{
				{
					name:    "Namespace option",
					options: []InstallOption{Namespace("ns")},
					errors: func(mg []ManifestGroup) error {
						return nil
					},
				},
				{
					name:    "SPIRE option",
					options: []InstallOption{SPIRE},
					errors: func(mg []ManifestGroup) error {
						return nil
					},
				},
				{
					name:    "InternalRedis option",
					options: []InstallOption{Namespace("ns"), InternalRedis},
					errors: func(mg []ManifestGroup) error {
						return nil
					},
				},
				{
					name:    "ExternalRedis option",
					options: []InstallOption{ExternalRedis(&ExternalRedisConfig{URL: "redis://:pass@extserver:6379/2"})},
					errors: func(mg []ManifestGroup) error {
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
						manifests := version.Manifests()
						if err := tc.errors(manifests); err != nil {
							t.Fatal(err)
							// y, _ := yaml.Marshal(group)
							// fmt.Println(string(y))
						}
					})

					// t.Run("sidecar", func(t *testing.T) {
					// 	sidecar := version.Sidecar()
					// 	y, _ := yaml.Marshal(sidecar)
					// 	fmt.Println(string(y))
					// })
				})
			}
		})
	}
}
