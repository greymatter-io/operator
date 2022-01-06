package version

import (
	"fmt"
	"strings"
	"testing"

	"github.com/greymatter-io/operator/pkg/cueutils"

	"cuelang.org/go/cue"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"
)

type testOptions struct {
	name           string
	xdsCluster     string
	printYAML      bool
	options        []cue.Value
	checkManifests func(*testing.T, []ManifestGroup)
	checkSidecar   func(*testing.T, Sidecar)
}

func TestVersionManifests_1_7(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	testVersionManifests(t, loadVersion(t, "1.7"))
}

func TestVersionManifests_1_6(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	testVersionManifests(t, loadVersion(t, "1.6"))
}

func testVersionManifests(t *testing.T, v Version, to ...testOptions) {

	// Run all general tests for manifests
	t.Run("without options", func(t *testing.T) {
		// m := v.Manifests()
		// y, _ := yaml.Marshal(m[2].Deployment)
		// fmt.Println(string(y))
		// unimplemented
		// all expected manifests exist
	})

	baseOptions := []cue.Value{
		cueutils.Strings(map[string]string{
			"MeshName":         "mymesh",
			"ReleaseVersion":   v.name,
			"InstallNamespace": "myns",
			"Zone":             "myzone",
		}),
	}

	// Run tests with testOptions for settings available in all versions,
	// plus any additional testOptions specified in 'to'.
	for _, tc := range append([]testOptions{
		{
			name:    "With base options",
			options: baseOptions,

			checkManifests: func(t *testing.T, manifests []ManifestGroup) {
				t.Run("MeshName", func(t *testing.T) {
					catalogConfigMaps := manifests[4].ConfigMaps
					if len(catalogConfigMaps) == 0 {
						t.Fatal("expected catalog to have ConfigMaps")
					}
					if name := catalogConfigMaps[0].Name; name != "catalog-seed" {
						t.Fatalf("expected the first ConfigMap to be 'catalog-seed', got %s", name)
					}
					seedFile, ok := catalogConfigMaps[0].Data["seed.yaml"]
					if !ok {
						t.Fatal("ConfigMap data does not have file 'seed.yaml'")
					}
					if !strings.HasPrefix(seedFile, "mymesh") {
						t.Fatalf("seed file does not start with 'mymesh', got %s", seedFile)
					}
				})

				// TODO: ReleaseVersion

				t.Run("InstallNamespace", func(t *testing.T) {

					// All resources reference the InstallNamespace
					for _, group := range manifests {
						if group.Deployment != nil && group.Deployment.Namespace != "myns" {
							t.Errorf("expected Deployment namespace to be 'ns', got %s", group.Deployment.Namespace)
						}
						if group.StatefulSet != nil && group.StatefulSet.Namespace != "myns" {
							t.Errorf("expected StatefulSet namespace to be 'ns', got %s", group.StatefulSet.Namespace)
						}
						if group.Service != nil && group.Service.Namespace != "myns" {
							t.Errorf("expected Service namespace to be 'ns', got %s", group.Service.Namespace)
						}
						for _, cm := range group.ConfigMaps {
							if cm.Namespace != "myns" {
								t.Errorf("expected ConfigMap %s's namespace to be 'ns', got %s", cm.Name, cm.Namespace)
							}
						}
						for _, s := range group.Secrets {
							if s.Namespace != "myns" {
								t.Errorf("expected ConfigMap %s's namespace to be 'ns', got %s", s.Name, s.Namespace)
							}
						}
						if group.Ingress != nil && group.Ingress.Namespace != "myns" {
							t.Errorf("expected Ingress namespace to be 'ns', got %s", group.Ingress.Namespace)
						}
					}

					// Edge's XDS_HOST references the InstallNamespace
					xdsHost, ok := getEnvValue(manifests[0].Deployment.Spec.Template.Spec.Containers[0], "XDS_HOST")
					if !ok {
						t.Fatal("did not find 'XDS_HOST' env in edge container")
					}
					if !strings.Contains(xdsHost, "myns") {
						t.Fatalf("expected to find 'ns' in XDS_HOST env, got '%s'", xdsHost)
					}
				})

				t.Run("Zone", func(t *testing.T) {

					// Edge's XDS_ZONE references the Zone
					zone, ok := getEnvValue(manifests[0].Deployment.Spec.Template.Spec.Containers[0], "XDS_ZONE")
					if !ok {
						t.Fatal("did not find 'XDS_ZONE' env in edge container")
					}
					if zone != "myzone" {
						t.Fatalf("expected 'myzone' to be XDS_ZONE env, got '%s'", zone)
					}

					// Control & Control API's GM_CONTROL_API_ZONE_NAME references the Zone
					for _, container := range manifests[3].Deployment.Spec.Template.Spec.Containers {
						zone, ok := getEnvValue(container, "GM_CONTROL_API_ZONE_NAME")
						if !ok {
							t.Fatalf("did not find 'GM_CONTROL_API_ZONE_NAME' env in container %s", container.Name)
						}
						if zone != "myzone" {
							t.Fatalf("expected 'myzone' to be XDS_ZONE env, got '%s'", zone)
						}
					}

					// Catalog's seed file references the Zone in the mesh's default session
					seedFile, ok := manifests[4].ConfigMaps[0].Data["seed.yaml"]
					if !ok {
						t.Fatal("ConfigMap data does not have file 'seed.yaml'")
					}
					if !strings.Contains(seedFile, "zone: myzone") {
						t.Fatalf("seed file does not contain with 'zone: myzone', got %s", seedFile)
					}
				})

			},
		},
		{
			name: "StringSlices:WatchNamespaces",
			options: []cue.Value{
				cueutils.Strings(map[string]string{"InstallNamespace": "install"}),
				cueutils.StringSlices(map[string][]string{"WatchNamespaces": {"apples", "oranges", "apples"}}),
			},
			checkManifests: func(t *testing.T, manifests []ManifestGroup) {
				control := manifests[3].Deployment.Spec.Template.Spec.Containers[0]
				ns, ok := getEnvValue(control, "GM_CONTROL_KUBERNETES_NAMESPACES")
				if !ok {
					t.Fatal("did not find 'GM_CONTROL_KUBERNETES_NAMESPACES' env in control container")
				}
				namespaces := strings.Split(ns, ",")
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
			name: "UserTokens",
			options: []cue.Value{UserTokens(`[
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
			options: []cue.Value{JWTSecrets()},
			checkManifests: func(t *testing.T, manifests []ManifestGroup) {
				// unimplemented
				// check for expected secret and references to secret
			},
		},
		{
			name: "Ingress",
			options: []cue.Value{
				cueutils.Strings(map[string]string{
					"InstallNamespace": "myns",
					"IngressSubDomain": "myaddress.com",
				}),
			},
			checkManifests: func(t *testing.T, manifests []ManifestGroup) {
				edge := manifests[0]
				if edge.Ingress == nil {
					t.Fatal("Ingress was not created")
				}
			},
		},
	}, to...) {
		t.Run(tc.name, func(t *testing.T) {
			vc := v.Copy()
			vc.Unify(tc.options...)
			if err := vc.cue.Err(); err != nil {
				cueutils.LogError(logger, err)
				t.FailNow()
			}
			manifests := vc.Manifests()
			if tc.printYAML {
				y, _ := yaml.Marshal(manifests)
				fmt.Println(string(y))
			}
			tc.checkManifests(t, manifests)
		})
	}
}

func getEnvValue(container corev1.Container, key string) (string, bool) {
	var value string
	for _, e := range container.Env {
		if e.Name == key {
			value = e.Value
		}
	}
	return value, value != ""
}

func loadVersion(t *testing.T, name string) Version {
	versions, err := loadBaseWithVersions(nil)
	if err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	v, ok := versions[name]
	if !ok {
		t.Fatalf("did not find version %s", name)
	}

	if err := v.cue.Err(); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	return v
}
