package installer

import (
	"fmt"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
)

var expectedVersions = []string{"1.6"}

func TestLoadVersions(t *testing.T) {
	versions, err := loadVersions()
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range expectedVersions {
		if _, ok := versions[v]; !ok {
			t.Errorf("did not load version file for %s", v)
		}
	}
}

func TestVersions(t *testing.T) {
	base, err := loadBaseCueValue()
	if err != nil {
		t.Fatal(err)
	}

	files, err := loadFiles()
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		fileName := file.Name()
		t.Run(fileName, func(t *testing.T) {
			data, err := filesystem.ReadFile(fmt.Sprintf("versions/%s", fileName))
			if err != nil {
				t.Fatal("failed to read file", err)
			}
			if err := yaml.Validate([]byte(data), base); err != nil {
				for _, e := range errors.Errors(err) {
					t.Error(e)
				}
			}
		})
	}
}

func TestBaseCueValue(t *testing.T) {
	base, err := loadBaseCueValue()
	if err != nil {
		t.Fatal(err)
	}

	if err := base.Validate(); err != nil {
		for _, e := range errors.Errors(err) {
			t.Error(e)
		}
	}
}

func loadBaseCueValue() (cue.Value, error) {
	buildInstances := load.Instances([]string{"base.cue"}, nil)
	if len(buildInstances) != 1 {
		return cue.Value{}, fmt.Errorf("expected one Cue build instance from single entrypoint base.cue")
	}
	bi := buildInstances[0]
	if bi.Err != nil {
		return cue.Value{}, bi.Err
	}
	return cuecontext.New().BuildInstance(bi), nil
}
