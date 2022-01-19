package fabric

import (
	"cuelang.org/go/cue"
	"github.com/greymatter-io/operator/pkg/cuemodule"
)

var value *cue.Value

// Init loads the meshconfigs Cue package in order to load our mesh config templates.
// It should be called on startup of the operator.
func Init() error {
	v, err := cuemodule.LoadPackage("meshconfigs")
	if err != nil {
		return err
	}

	value = &v

	return nil
}
