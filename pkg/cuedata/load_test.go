package cuedata

import (
	"testing"

	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestLoadBase(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if _, err := LoadPackages("base", "onesix"); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}
}
