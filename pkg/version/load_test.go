package version

import (
	"fmt"
	"testing"

	"cuelang.org/go/cue/errors"
)

var expectedVersions = []string{"1.6"}

func TestLoad(t *testing.T) {
	versions, err := Load()
	if err != nil {
		logCueErrors(err)
		t.Fatal("failed to load versions")
	}

	for _, name := range expectedVersions {
		t.Run(fmt.Sprintf("loads expected version %s", name), func(t *testing.T) {
			if _, ok := versions[name]; !ok {
				t.Fatal()
			}
		})
	}

	for name, version := range versions {
		t.Run(fmt.Sprintf("loads valid version %s", name), func(t *testing.T) {
			if err := version.cue.Err(); err != nil {
				logCueErrors(err)
				t.Errorf("found invalid version %s", name)
			}
		})
	}
}

func TestLoadBase(t *testing.T) {
	if _, err := loadBase(); err != nil {
		for _, e := range errors.Errors(err) {
			t.Error(e)
		}
	}
}
