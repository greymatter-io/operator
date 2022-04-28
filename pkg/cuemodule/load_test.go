package cuemodule

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

func TestLoadStatus(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	_, mesh := LoadAll("")
	logger.Info("LoadAll mesh status", "mesh.Status.SidecarList", mesh.Status.SidecarList)
}
