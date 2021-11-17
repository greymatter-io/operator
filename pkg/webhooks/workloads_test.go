package webhooks

import (
	"testing"
)

func TestNetworkMode(t *testing.T) {
	const networkModeLabel = "greymatter.io/network-mode"
	labels := make(map[string]string)

	t.Run("label not present => add sidecar", func(t *testing.T) {
		b := networkModeIncludeSidecar(labels)
		if b != true {
			t.Fatalf("Sidecar should be added")
		}
	})

	// asume default as route
	t.Run("label present and unusual => add sidecar", func(t *testing.T) {
		labels[networkModeLabel] = "yeahboi"

		b := networkModeIncludeSidecar(labels)
		if b != true {
			t.Fatalf("Sidecar should be added")
		}

	})

	//is internal
	t.Run("label present and internal => add sidecar", func(t *testing.T) {
		labels[networkModeLabel] = "internal"
		//
		b := networkModeIncludeSidecar(labels)
		if b != true {
			t.Fatalf("Sidecar should be added")
		}
	})

	// is exclude
	t.Run("label present and exclude => false", func(t *testing.T) {
		labels[networkModeLabel] = "exclude"
		//
		b := networkModeIncludeSidecar(labels)
		if b != false {
			t.Fatalf("Sidecar should not be added")
		}
	})

}
