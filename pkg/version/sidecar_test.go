package version

import (
	"fmt"
	"testing"

	"github.com/greymatter-io/operator/pkg/assert"
	"github.com/greymatter-io/operator/pkg/cueutils"

	"cuelang.org/go/cue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"
)

func TestVersionSidecar_1_7(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	testVersionSidecar(t, loadVersion(t, "1.7"))
}

func TestVersionSidecar_1_6(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	testVersionSidecar(t, loadVersion(t, "1.6"))
}

func testVersionSidecar(t *testing.T, v Version, to ...testOptions) {

	// Run all general tests for sidecar
	t.Run("without options", func(t *testing.T) {
		v.SidecarTemplate()("mock")
		// unimplemented
		// all expected sidecar values exist
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
			name:       "With base options",
			xdsCluster: "mock",
			options:    baseOptions,
			checkSidecar: func(t *testing.T, sidecar Sidecar) {
				t.Run("Has expected env values",
					assert.ContainerHasEnvValues(sidecar.Container, map[string]string{
						"XDS_CLUSTER": "mock",
						"XDS_HOST":    "control.myns.svc.cluster.local",
						"XDS_ZONE":    "myzone",
					}),
				)

				t.Run("Has no StaticConfig", func(t *testing.T) {
					if len(sidecar.StaticConfig) != 0 {
						t.Error("found unexpected StaticConfig")
					}
				})
			},
		},
		{
			name:       "With xdsCluster edge",
			xdsCluster: "edge",
			options:    baseOptions,
			checkSidecar: func(t *testing.T, sidecar Sidecar) {
				if len(sidecar.StaticConfig) == 0 {
					t.Fatal("no StaticConfig was set")
				}
				// y, _ := yaml.Marshal(sidecar.StaticConfig)
				// fmt.Println(string(y))
				t.Run("StaticConfig discovers from control.<namespace>.svc.cluster.local",
					assert.JSONHasSubstrings(sidecar.StaticConfig,
						`"address":"control.myns.svc.cluster.local","port_value":50000`,
					),
				)
			},
		},
		{
			name:       "With xdsCluster control",
			xdsCluster: "control",
			options:    baseOptions,
			checkSidecar: func(t *testing.T, sidecar Sidecar) {
				if len(sidecar.StaticConfig) == 0 {
					t.Fatal("no StaticConfig was set")
				}
				// y, _ := yaml.Marshal(sidecar.StaticConfig)
				// fmt.Println(string(y))
				t.Run("StaticConfig discovers from localhost",
					assert.JSONHasSubstrings(sidecar.StaticConfig,
						`"address":"127.0.0.1","port_value":50000`,
					),
				)
			},
		},
		{
			name:       "With xdsCluster catalog",
			xdsCluster: "catalog",
			options:    baseOptions,
			checkSidecar: func(t *testing.T, sidecar Sidecar) {
				if len(sidecar.StaticConfig) == 0 {
					t.Fatal("no StaticConfig was set")
				}
				// y, _ := yaml.Marshal(sidecar.StaticConfig)
				// fmt.Println(string(y))
				t.Run("StaticConfig discovers from control.<namespace>.svc.cluster.local",
					assert.JSONHasSubstrings(sidecar.StaticConfig,
						`"address":"control.myns.svc.cluster.local","port_value":50000`,
					),
				)
			},
		},
		{
			name:       "With xdsCluster gm-redis",
			xdsCluster: "gm-redis",
			options:    baseOptions,
			checkSidecar: func(t *testing.T, sidecar Sidecar) {
				if len(sidecar.StaticConfig) == 0 {
					t.Fatal("no StaticConfig was set")
				}
				// y, _ := yaml.Marshal(sidecar.StaticConfig)
				// fmt.Println(string(y))
				t.Run("StaticConfig's gm-redis goes to localhost",
					assert.JSONHasSubstrings(sidecar.StaticConfig,
						`"address":"127.0.0.1","port_value":6379`,
					),
				)
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
			var sidecar Sidecar
			if tc.xdsCluster != "" {
				sidecar = vc.SidecarTemplate()(tc.xdsCluster)
			} else {
				sidecar = vc.SidecarTemplate()("mock")
			}
			if tc.printYAML {
				y, _ := yaml.Marshal(sidecar)
				fmt.Println(string(y))
			}
			tc.checkSidecar(t, sidecar)
		})
	}
}
