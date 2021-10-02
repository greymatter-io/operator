package version

import (
	"testing"
)

var expectedVersions = []string{"1.6"}

func TestLoad(t *testing.T) {
	versions, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range expectedVersions {
		if _, ok := versions[v]; !ok {
			t.Errorf("did not load valid install values for version %s", v)
		}
	}
}
