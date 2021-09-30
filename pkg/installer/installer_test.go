package installer

import (
	"testing"

	"github.com/greymatter-io/operator/pkg/values"
)

func TestLoadBaseValues(t *testing.T) {
	files, err := filesystem.ReadDir("versions")
	if err != nil {
		t.Fatal(err)
	}

	versions := make(map[string]*values.Values)
	t.Run("loads versions files from an embed.FS without error", func(t *testing.T) {
		vs, err := loadBaseValues(files)
		if err != nil {
			t.Error(err)
		} else {
			versions = vs
		}
	})

	t.Run("loads base values for Grey Matter v1.6", func(t *testing.T) {
		if _, ok := versions["v1.6"]; !ok {
			t.Error("did not find v1.6 in values")
		}
	})
}
