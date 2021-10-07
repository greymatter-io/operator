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
				checkManifests func(*testing.T, []ManifestGroup)
				checkSidecar   func(*testing.T, Sidecar)
			}{
				{
					name:    "InstallNamespace",
					options: []InstallOption{InstallNamespace("ns")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name:    "Zone",
					options: []InstallOption{Zone("myzone")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name:    "ImagePullSecretName",
					options: []InstallOption{Zone("mysecret")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name:    "MeshPort",
					options: []InstallOption{MeshPort(10999)},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						edge := manifests[0].Deployment.Spec.Template.Spec.Containers[0]
						var proxyPort *corev1.ContainerPort
						for _, p := range edge.Ports {
							if p.Name == "proxy" {
								proxyPort = &p
							}
						}
						if proxyPort == nil {
							t.Fatal("No proxy port found in edge")
						}
						if proxyPort.ContainerPort != 10999 {
							t.Errorf("Expected proxy port to be 10999 but got %d", proxyPort.ContainerPort)
						}
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						var proxyPort *corev1.ContainerPort
						for _, p := range sidecar.Container.Ports {
							if p.Name == "proxy" {
								proxyPort = &p
							}
						}
						if proxyPort == nil {
							t.Fatal("No proxy port found in edge")
						}
						if proxyPort.ContainerPort != 10999 {
							t.Errorf("Expected proxy port to be 10999 but got %d", proxyPort.ContainerPort)
						}
					},
				},
				{
					name:    "SPIRE",
					options: []InstallOption{SPIRE},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name:    "Redis internal",
					options: []InstallOption{InstallNamespace("ns"), Redis("")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name:    "Redis external",
					options: []InstallOption{Redis("redis://:pass@extserver:6379/2")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name: "UserTokens",
					options: []InstallOption{UserTokens(`[
						{
							"label": "CN=engineer,OU=engineering,O=Decipher,=Alexandria,=Virginia,C=US",
							"values": {
								"email": ["engineering@greymatter.io"],
								"org": ["www.greymatter.io"],
								"privilege": ["root"]
							}
						}
					]`)},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						y, _ := yaml.Marshal(manifests)
						fmt.Println(string(y))
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
				{
					name:    "JWTSecrets",
					options: []InstallOption{JWTSecrets},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					v := version.Copy()
					v.Apply(tc.options...)
					if err := v.cue.Err(); err != nil {
						logCueErrors(err)
						t.Fatal()
					}
					t.Run("manifests", func(t *testing.T) {
						tc.checkManifests(t, v.Manifests())
					})
					t.Run("sidecar", func(t *testing.T) {
						tc.checkSidecar(t, v.Sidecar())
					})
				})
			}
		})
	}
}
