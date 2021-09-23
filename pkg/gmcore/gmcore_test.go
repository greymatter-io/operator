package gmcore

import (
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
)

func TestLoadValues(t *testing.T) {
	files, err := filesystem.ReadDir("values")
	if err != nil {
		t.Fatal(err)
	}

	values := make(map[string]*v1alpha1.SystemValuesConfig)
	t.Run("loads values files from an embed.FS without error", func(t *testing.T) {
		vs, err := loadValues(files)
		if err != nil {
			t.Error(err)
		} else {
			values = vs
		}
	})

	t.Run("loads an expected embedded values file and stores it in a map", func(t *testing.T) {
		if _, ok := values["v1.6"]; !ok {
			t.Error("expected to load values from values/v1.6.yaml")
		}
	})
}
