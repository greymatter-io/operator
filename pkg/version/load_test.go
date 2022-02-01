package version

import (
	"fmt"
	"testing"

	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var expectedVersions = []string{"1.6", "1.7"}

func TestLoad(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	versions, err := loadBaseWithVersions(cuemodule.LoadPackageForTest)
	if err != nil {
		cueutils.LogError(logger, err)
		t.Fatal("failed to load versions")
	}

	for _, name := range expectedVersions {
		t.Run(fmt.Sprintf("loads expected version %s", name), func(t *testing.T) {
			if _, ok := versions[name]; !ok {
				t.FailNow()
			}
		})
	}

	for name, version := range versions {
		t.Run(fmt.Sprintf("loads valid version %s", name), func(t *testing.T) {
			if err := version.cue.Err(); err != nil {
				cueutils.LogError(logger, err)
				t.Errorf("found invalid version %s", name)
			}
		})
	}
}
