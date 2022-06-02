package cuemodule

import (
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestLoadStatus(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	operatorCUE, _ := LoadAll("")
	_, defaults := operatorCUE.ExtractConfig()
	defaults.SidecarList = nil
	if defaults.SidecarList == nil {
		defaults.SidecarList = []string{}
	}
	tempOperatorCUE, _ := operatorCUE.TempGMValueUnifiedWithDefaults(defaults)
	redisListener, _ := tempOperatorCUE.ExtractRedisListener()
	logger.Info("blurp", "listener", redisListener)
	//logger.Info("LoadAll sidecarList", "SidecarList", defaults.SidecarList)
}
