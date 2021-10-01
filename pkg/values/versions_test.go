package values

import (
	"testing"
)

var expectedVersions = []string{"1.6"}

func TestLoadYAMLVersions(t *testing.T) {
	versions, err := LoadYAMLVersions()
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range expectedVersions {
		if _, ok := versions[v]; !ok {
			t.Errorf("did not load version file for %s", v)
		}
	}
}

// func TestVersions(t *testing.T) {
// 	baseSchema := buildSchemaOrDie(t, false)

// 	versions, err := LoadYAMLVersions()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	for name, data := range versions {
// 		t.Run(name, func(t *testing.T) {
// 			validateYAML(t, baseSchema, data)
// 		})
// 	}
// }
