package version

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestVersions(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	versions, err := Load()
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
				checkManifests func(*testing.T, []ManifestGroup)
				checkSidecar   func(*testing.T, Sidecar)
			}{
				{
					name:    "Namespace option",
					options: []InstallOption{Namespace("ns")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
					},
				},
				{
					name:    "SPIRE option",
					options: []InstallOption{SPIRE},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
					},
				},
				{
					name:    "Redis internal option",
					options: []InstallOption{Namespace("ns"), Redis("")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
					},
				},
				{
					name:    "Redis external option",
					options: []InstallOption{Redis("redis://:pass@extserver:6379/2")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// y, _ := yaml.Marshal(manifests)
						// fmt.Println(string(y))
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
					},
				},
				{
					name:    "MeshPort option",
					options: []InstallOption{MeshPort(10999)},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {

					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						y, _ := yaml.Marshal(sidecar)
						fmt.Println(string(y))

						if len(sidecar.Container.Ports) == 0 {
							t.Fatal("No ports found in sidecar")
						}

						var proxyPort *corev1.ContainerPort
						for _, p := range sidecar.Container.Ports {
							if p.Name == "proxy" {
								proxyPort = &p
							}
						}

						if proxyPort == nil {
							t.Fatal("No proxy port found in sidecar")
						}

						// Test proxyPort
						// Should not have 0
						// Should not have default value 10808
						// Should have the value we expect (10999)

						proxyExists := false
						proxyUniqueCheck := 0
						for _, p := range sidecar.Container.Ports {
							// fmt.Printf("\n --test sidecar ---> [name: %s; port: %d] \n", p.Name, p.ContainerPort)
							// This check can happen when outputs.cue is modified instead of inputs.cue
							if p.Name == "proxy" && p.ContainerPort == 10999 {
								proxyExists = true
								proxyUniqueCheck++
							}
							if p.Name == "proxy" && p.ContainerPort != 10999 {
								t.Error("Container port named proxy found however it does not have the correct port number")
							}
							if p.ContainerPort == 10909 || p.Name != "proxy" {
								t.Error("A duplicate port numbers found matching the proxy port specified")
							}
						}
						if !proxyExists {
							t.Error("Proxy port not found")
						}
						if proxyUniqueCheck > 1 {
							t.Error("Too Many Proxys with the name proxy and port 10909 were injected")
						}

					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					vCopy := version.Copy()
					vCopy.Apply(tc.options...)
					if err := vCopy.cue.Err(); err != nil {
						logCueErrors(err)
						t.Fatal()
					}
					t.Run("manifests", func(t *testing.T) {
						tc.checkManifests(t, vCopy.Manifests())
					})
					t.Run("sidecar", func(t *testing.T) {
						tc.checkSidecar(t, vCopy.Sidecar())
					})
				})
			}
		})
	}
}
