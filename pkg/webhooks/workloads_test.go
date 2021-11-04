package webhooks

import (
	"testing"
)

func TestExcludeFromMesh(t *testing.T) {
	const exclusionLabel = "greymatter.io/exclude-from-mesh"
	labels := make(map[string]string)

	t.Run("label not present => false", func(t *testing.T) {
		if excludeFromMesh(labels) {
			t.Fatalf("no label present but resource marked for exclusion")
		}
	})

	t.Run("label present and unreadable => false", func(t *testing.T) {
		labels[exclusionLabel] = "yeahboi"
		if excludeFromMesh(labels) {
			t.Fatalf("label found to exclude but value is not parsable.  This will add it to the mesh.")
		}
	})

	t.Run("label present and true => true", func(t *testing.T) {
		labels[exclusionLabel] = "true"
		if !excludeFromMesh(labels) {
			t.Fatalf("label found to exclude mesh but did not exlude it")
		}
	})

	t.Run("label present and false => false", func(t *testing.T) {
		labels[exclusionLabel] = "false"
		if excludeFromMesh(labels) {
			t.Fatalf("label found to exclude mesh but did not exlude it")
		}
	})

}
