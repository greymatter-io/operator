package version

import (
	"strings"
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/cueutils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type test struct {
	name string
	run  func(*testing.T, *Version)
}

type manifestArgs struct {
	mesh  *v1alpha1.Mesh
	opts  []Opt
	tests []test
}

func TestBasicMesh(t *testing.T) {
	testManifestOutputs(t,
		manifestArgs{
			mesh: &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{Name: "default-mesh"},
				Spec: v1alpha1.MeshSpec{
					ReleaseVersion:   "latest",
					InstallNamespace: "myns",
					WatchNamespaces:  []string{"ns2", "ns3"},
					Zone:             "default-zone",
				},
			},
			opts: nil,
			tests: []test{
				{
					name: "mesh.metadata.name",
					run: func(t *testing.T, v *Version) {
						manifests := v.Manifests()
						seedFile, ok := manifests[4].ConfigMaps[0].Data["seed.yaml"]
						if !ok {
							t.Fatal("ConfigMap data does not have file 'seed.yaml'")
						}
						if !strings.HasPrefix(seedFile, "default-mesh") {
							t.Fatalf("seed file does not start with 'default-mesh', got %s", seedFile)
						}
					},
				},
				{
					name: "mesh.spec.release_version",
					run: func(t *testing.T, v *Version) {
						var vs struct {
							Versions map[string]string `json:"versions"`
						}
						if err := cueutils.Extract(v.cue, &vs); err != nil {
							t.Fatal(err)
						}
						for svc, img := range vs.Versions {
							if img == "" {
								t.Errorf("service '%s' is missing an OCI container string", svc)
							}
						}
					},
				},
				{
					name: "mesh.spec.install_namespace",
					run: func(t *testing.T, v *Version) {
						manifests := v.Manifests()
						for _, group := range manifests {
							if group.Deployment != nil && group.Deployment.Namespace != "myns" {
								t.Errorf("expected Deployment namespace to be 'myns', got %s", group.Deployment.Namespace)
							}
							if group.StatefulSet != nil && group.StatefulSet.Namespace != "myns" {
								t.Errorf("expected StatefulSet namespace to be 'myns', got %s", group.StatefulSet.Namespace)
							}
							if group.Service != nil && group.Service.Namespace != "myns" {
								t.Errorf("expected Service namespace to be 'myns', got %s", group.Service.Namespace)
							}
							for _, cm := range group.ConfigMaps {
								if cm.Namespace != "myns" {
									t.Errorf("expected ConfigMap %s's namespace to be 'myns', got %s", cm.Name, cm.Namespace)
								}
							}
							for _, s := range group.Secrets {
								if s.Namespace != "myns" {
									t.Errorf("expected ConfigMap %s's namespace to be 'myns', got %s", s.Name, s.Namespace)
								}
							}
							if group.Ingress != nil && group.Ingress.Namespace != "myns" {
								t.Errorf("expected Ingress namespace to be 'myns', got %s", group.Ingress.Namespace)
							}
						}
						// Edge's XDS_HOST references the InstallNamespace
						xdsHost, ok := getEnvValue(manifests[0].Deployment.Spec.Template.Spec.Containers[0], "XDS_HOST")
						if !ok {
							t.Fatal("did not find 'XDS_HOST' env in edge container")
						}
						if !strings.Contains(xdsHost, "myns") {
							t.Fatalf("expected to find 'myns' in XDS_HOST env, got '%s'", xdsHost)
						}
					},
				},
				{
					name: "mesh.spec.watch_namespaces",
					run: func(t *testing.T, v *Version) {
						manifests := v.Manifests()
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
						for _, namespace := range []string{"myns", "ns2", "ns3"} {
							if _, ok := set[namespace]; !ok {
								t.Errorf("Expected namespaces to contain %s: got %v", namespace, namespaces)
							}
						}
					},
				},
				{
					name: "mesh.spec.zone",
					run: func(t *testing.T, v *Version) {
						manifests := v.Manifests()
						// Edge's XDS_ZONE references the Zone
						zone, ok := getEnvValue(manifests[0].Deployment.Spec.Template.Spec.Containers[0], "XDS_ZONE")
						if !ok {
							t.Fatal("did not find 'XDS_ZONE' env in edge container")
						}
						if zone != "default-zone" {
							t.Fatalf("expected 'default-zone' to be XDS_ZONE env, got '%s'", zone)
						}
						// Control & Control API's GM_CONTROL_API_ZONE_NAME references the Zone
						for _, container := range manifests[3].Deployment.Spec.Template.Spec.Containers {
							zone, ok := getEnvValue(container, "GM_CONTROL_API_ZONE_NAME")
							if !ok {
								t.Fatalf("did not find 'GM_CONTROL_API_ZONE_NAME' env in container %s", container.Name)
							}
							if zone != "default-zone" {
								t.Fatalf("expected 'default-zone' to be XDS_ZONE env, got '%s'", zone)
							}
						}
						// Catalog's seed file references the Zone in the mesh's default session
						seedFile, ok := manifests[4].ConfigMaps[0].Data["seed.yaml"]
						if !ok {
							t.Fatal("ConfigMap data does not have file 'seed.yaml'")
						}
						if !strings.Contains(seedFile, "zone: default-zone") {
							t.Fatalf("seed file does not contain with 'zone: default-zone', got %s", seedFile)
						}
					},
				},
			},
		},
	)
}

func testManifestOutputs(t *testing.T, a manifestArgs) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	base, err := cuemodule.LoadPackageForTest("base")
	if err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}
	version, err := New(base, a.mesh, a.opts...)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range a.tests {
		t.Run(test.name, func(t *testing.T) {
			test.run(t, version)
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
