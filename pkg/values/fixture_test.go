package values

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
)

func TestFixture(t *testing.T) {
	buildInstances := load.Instances([]string{"../installer/base.cue"}, nil)
	if len(buildInstances) != 1 {
		t.Fatal("expected one Cue build instance from single entrypoint base.cue")
	}
	bi := buildInstances[0]
	if bi.Err != nil {
		t.Fatal(bi.Err)
	}
	base := cuecontext.New().BuildInstance(bi)
	if err := yaml.Validate([]byte(fixture), base); err != nil {
		for _, e := range errors.Errors(err) {
			t.Error(e)
		}
	}
}
