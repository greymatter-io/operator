package version

import (
	"strings"
	"testing"

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
						edge := manifests[0]
						// y, _ := yaml.Marshal(edge)
						// fmt.Println(string(y))

						edgeContainer := edge.Deployment.Spec.Template.Spec.Containers[0]

						if len(edgeContainer.Ports) == 0 {
							t.Fatal("No ports found in edge")
						}

						var proxyPort *corev1.ContainerPort
						for _, p := range edgeContainer.Ports {
							if p.Name == "proxy" {
								proxyPort = &p
							}
						}
						if proxyPort == nil {
							t.Fatal("No proxy port found in edge")
						}
						actualProxyPort := proxyPort.ContainerPort
						// Should not have 0
						if actualProxyPort == 0 {
							t.Fatal("Proxy Port is set to 0.  Was not updated")
						}
						// Should not have default value 10808
						if actualProxyPort == 10808 {
							t.Fatalf("Proxy Port is set to [%d] the default value and was not updated to [10999]", actualProxyPort)
						}
						// Should have the value we expect (10999)
						if actualProxyPort != 10999 {
							t.Fatalf("Proxy Port is set to [%d] and was not updated to [10999]", actualProxyPort)
						}

					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))

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
						actualProxyPort := proxyPort.ContainerPort
						// Should not have 0
						if actualProxyPort == 0 {
							t.Fatal("Proxy Port is set to 0.  Was not updated")
						}
						// Should not have default value 10808
						if actualProxyPort == 10808 {
							t.Fatalf("Proxy Port is set to [%d] the default value and was not updated to [10999]", actualProxyPort)
						}
						// Should have the value we expect (10999)
						if actualProxyPort != 10999 {
							t.Fatalf("Proxy Port is set to [%d] and was not updated to [10999]", actualProxyPort)
						}

					},
				},
				{
					name:    "Watch Namespace Option- no additional namespaces",
					options: []InstallOption{WatchNamespaces("mygreymatter", "")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {

						controlContainer := manifests[1].Deployment.Spec.Template.Spec.Containers[0]

						var watchNamespacesEnvVar *corev1.EnvVar
						for _, e := range controlContainer.Env {
							if e.Name == "GM_CONTROL_KUBERNETES_NAMESPACES" {
								watchNamespacesEnvVar = &e
							}
						}

						expectedValue := "mygreymatter"
						envarValue := watchNamespacesEnvVar.Value

						s := strings.Split(envarValue, ",")
						if len(s) > 1 {
							t.Fatalf("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] includes too Many Namespaces.  actual: [%s]  expected: [%s]", envarValue, "mygreymatter")
						}

						// envar not set
						if watchNamespacesEnvVar == nil {
							t.Fatal("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] is not set in Control container")
						}
						// envar blank
						if envarValue == "" {
							t.Fatal("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] has no value specified in Control container")
						}

						// envar is not as expected
						if envarValue != expectedValue {
							t.Fatalf("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] actual: [%s] expected: [%s]", envarValue, expectedValue)
						}
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
					},
				},
				{
					name:    "Watch Namespace Option- additional namespaces",
					options: []InstallOption{WatchNamespaces("mygreymatter", "apples,oranges,mygreymatter")},
					checkManifests: func(t *testing.T, manifests []ManifestGroup) {

						controlContainer := manifests[1].Deployment.Spec.Template.Spec.Containers[0]

						var watchNamespacesEnvVar *corev1.EnvVar
						for _, e := range controlContainer.Env {
							if e.Name == "GM_CONTROL_KUBERNETES_NAMESPACES" {
								watchNamespacesEnvVar = &e
							}
						}

						expectedValue := "mygreymatter,apples,oranges"
						envarValue := watchNamespacesEnvVar.Value

						s := strings.Split(envarValue, ",")
						if len(s) > 3 {
							t.Fatalf("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] includes too Many Namespaces.  actual: [%s]  expected: [%s]", envarValue, "mygreymatter")
						}

						// envar not set
						if watchNamespacesEnvVar == nil {
							t.Fatal("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] is not set in Control container")
						}
						// envar blank
						if envarValue == "" {
							t.Fatal("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] has no value specified in Control container")
						}

						expSlice := strings.Split(expectedValue, ",")
						envarValueSlice := strings.Split(envarValue, ",")
						// envar is not as expected
						if stringSlicesEqual(expSlice, envarValueSlice) {
							t.Fatalf("Environment Variable [GM_CONTROL_KUBERNETES_NAMESPACES] actual: [%s] expected: [%s]", envarValue, expectedValue)
						}
					},
					checkSidecar: func(t *testing.T, sidecar Sidecar) {
						// unimplemented
						// y, _ := yaml.Marshal(sidecar)
						// fmt.Println(string(y))
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

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
