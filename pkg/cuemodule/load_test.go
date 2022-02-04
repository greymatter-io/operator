package cuemodule

import (
	"testing"

	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger = ctrl.Log.WithName("cuemodule")
)

func TestLoadBase(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if _, err := LoadPackageForTest("base"); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}
}

func TestLoadMeshConfigs(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if _, err := LoadPackageForTest("meshconfigs"); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}
}
