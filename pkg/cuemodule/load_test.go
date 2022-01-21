package cuemodule

import (
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger = ctrl.Log.WithName("cuemodule")
)

func TestLoadBase(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if _, err := LoadPackage("base"); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}
}

func TestLoadVersions(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	values, err := LoadPackage("base")
	if err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	tests := []struct {
		mesh *v1alpha1.Mesh
	}{
		{
			mesh: &v1alpha1.Mesh{
				Spec: v1alpha1.MeshSpec{
					ReleaseVersion: "latest",
				},
			},
		},
		{
			mesh: &v1alpha1.Mesh{
				Spec: v1alpha1.MeshSpec{
					ReleaseVersion: "1.7",
				},
			},
		},
		{
			mesh: &v1alpha1.Mesh{
				Spec: v1alpha1.MeshSpec{
					ReleaseVersion: "1.6",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.mesh.Spec.ReleaseVersion, func(t *testing.T) {
			// Unify this with a mesh CR
			mval, err := cueutils.FromStruct("mesh", test.mesh)
			if err != nil {
				cueutils.LogError(logger, err)
				t.FailNow()
			}

			// Unify the base values with the mesh CR
			v := values.Unify(mval)
			var versions struct {
				Versions map[string]string `json:"versions"`
			}
			//lint:ignore SA1019 will upgrade this later
			codec := gocodec.New(&cue.Runtime{}, nil)
			if err := codec.Encode(v, &versions); err != nil {
				cueutils.LogError(logger, err)
				t.FailNow()
			}

			for _, vers := range versions.Versions {
				if vers == "" {
					t.Fatal("missing version")
				}
			}
		})
	}
}
