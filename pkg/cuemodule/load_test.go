package cuemodule

import (
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestLoadBase(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if _, err := LoadPackageForTest("base"); err != nil {
		LogError(logger, err)
		t.FailNow()
	}
}

func TestLoadMeshConfigs(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if _, err := LoadPackageForTest("meshconfigs"); err != nil {
		LogError(logger, err)
		t.FailNow()
	}
}
