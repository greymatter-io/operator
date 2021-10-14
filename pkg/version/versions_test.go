package version

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/greymatter-io/operator/pkg/cueutils"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestVersions(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	versions, err := loadBaseWithVersions()
	if err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	for name, v := range versions {
		if v.cue.Err(); err != nil {
			cueutils.LogError(logger, err)
			t.FailNow()
		}

		t.Run(name, func(t *testing.T) {
			t.Run("manifests", func(t *testing.T) {
				v.Manifests()
				// unimplemented
				// all expected manifests exist
			})

			t.Run("sidecar", func(t *testing.T) {
				v.SidecarTemplate()("mock")
				// unimplemented
				// all expected manifests exist
			})

			for _, tc := range []struct {
				name           string
				options        []InstallOption
				checkManifests func(*testing.T, []ManifestGroup)
				checkSidecar   func(*testing.T, Sidecar)
			}{
				{
					name:    "MeshName, InstallNamespace, Zone",
					options: []InstallOption{MeshName("mymesh"), InstallNamespace("ns"), Zone("myzone")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// each manifest references install namespace
						y, _ := yaml.Marshal(manifests[4])
						fmt.Println(string(y))
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// each manifest references install namespace
					},
				},
				{
					name:    "WatchNamespaces",
					options: []InstallOption{WatchNamespaces("install", "install", "apples", "oranges", "apples")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						control := manifests[3].Deployment.Spec.Template.Spec.Containers[0]
						var namespaces []string
						for _, e := range control.Env {
							if e.Name == "GM_CONTROL_KUBERNETES_NAMESPACES" {
								namespaces = strings.Split(e.Value, ",")
							}
						}
						if count := len(namespaces); count != 3 {
							t.Fatalf("Expected len(namespaces) to be 3 but got %d: %v", count, namespaces)
						}
						set := make(map[string]struct{})
						for _, namespace := range namespaces {
							set[namespace] = struct{}{}
						}
						for _, namespace := range []string{"install", "apples", "oranges"} {
							if _, ok := set[namespace]; !ok {
								t.Errorf("Expected namespaces to contain %s: got %v", namespace, namespaces)
							}
						}
					},
				},
				{
					name:    "ImagePullSecretName",
					options: []InstallOption{Zone("mysecret")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// core service deployments reference image pull secret name
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// sidecar.ImagePullSecretRef should reference image pull secret name
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
						// edge should have SPIRE settings
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// sidecar should have SPIRE settings, plus a volume
					},
				},
				{
					name:    "Redis internal",
					options: []InstallOption{InstallNamespace("ns"), Redis("")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// check for expected values
					},
				},
				{
					name:    "Redis external",
					options: []InstallOption{Redis("redis://:pass@extserver:6379/2")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// check for expected values
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
						// check for expected configMap and reference to configMap
					},
				},
				{
					name:    "JWTSecrets",
					options: []InstallOption{JWTSecrets},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {
						// unimplemented
						// check for expected secret and references to secret
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					vc := v.Copy()
					vc.Apply(tc.options...)
					if err := vc.cue.Err(); err != nil {
						cueutils.LogError(logger, err)
						t.FailNow()
					}
					if tc.checkManifests != nil {
						t.Run("manifests", func(t *testing.T) {
							tc.checkManifests(t, vc.Manifests())
						})
					}
					if tc.checkSidecar != nil {
						t.Run("sidecar", func(t *testing.T) {
							tc.checkSidecar(t, vc.SidecarTemplate()("mock"))
						})
					}
				})
			}
		})
	}
}
