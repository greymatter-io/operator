package version

import (
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/assert"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/cueutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestBasicSidecar(t *testing.T) {
	testSidecarOutputs(t,
		func(t *testing.T, v *Version) {
			sidecar := v.SidecarTemplate()("mock")
			t.Run("env", assert.ContainerHasEnvValues(sidecar.Container, map[string]string{
				"XDS_CLUSTER": "mock",
				"XDS_HOST":    "control.default.svc.cluster.local",
				"XDS_ZONE":    "default-zone",
			}))
			if len(sidecar.StaticConfig) != 0 {
				t.Errorf("found unexpected StaticConfig")
			}
		},
	)
}

func TestControlSidecar(t *testing.T) {
	testSidecarOutputs(t,
		func(t *testing.T, v *Version) {
			sidecar := v.SidecarTemplate()("control")
			if len(sidecar.StaticConfig) == 0 {
				t.Fatal("no StaticConfig was set")
			}
			t.Run("StaticConfig discovers from localhost",
				assert.JSONHasSubstrings(sidecar.StaticConfig,
					`"address":"127.0.0.1","port_value":50000`,
				),
			)
		},
	)
}

func TestCatalogSidecar(t *testing.T) {
	testSidecarOutputs(t,
		func(t *testing.T, v *Version) {
			sidecar := v.SidecarTemplate()("catalog")
			if len(sidecar.StaticConfig) == 0 {
				t.Fatal("no StaticConfig was set")
			}
			t.Run("StaticConfig discovers from control.<namespace>.svc.cluster.local",
				assert.JSONHasSubstrings(sidecar.StaticConfig,
					`"address":"control.default.svc.cluster.local","port_value":50000`,
				),
			)
		},
	)
}

func TestRedisSidecar(t *testing.T) {
	testSidecarOutputs(t,
		func(t *testing.T, v *Version) {
			sidecar := v.SidecarTemplate()("gm-redis")
			if len(sidecar.StaticConfig) == 0 {
				t.Fatal("no StaticConfig was set")
			}
			t.Run("StaticConfig's gm-redis goes to localhost",
				assert.JSONHasSubstrings(sidecar.StaticConfig,
					`"address":"127.0.0.1","port_value":6379`,
				),
			)
		},
	)
}

func testSidecarOutputs(t *testing.T, run func(*testing.T, *Version)) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	base, err := cuemodule.LoadPackageForTest("base")
	if err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}
	version, err := New(base, &v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "default-mesh"},
		Spec: v1alpha1.MeshSpec{
			ReleaseVersion:   "latest",
			InstallNamespace: "default",
			Zone:             "default-zone",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	run(t, version)
}
