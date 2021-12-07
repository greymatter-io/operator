package fabric

import (
	"fmt"

	"cuelang.org/go/cue"
	"github.com/greymatter-io/operator/pkg/cuedata"
)

var (
	value cue.Value
)

// Load reads the fabric Cue package in order to load fabric object templates.
// It should be called on startup of the operator.
func Load() error {
	var err error

	value, err = cuedata.LoadPackages("fabric")
	if err != nil {
		cuedata.LogError(logger, err)
		return fmt.Errorf("failed to load fabric templates")
	}

	// var data []string

	// for _, file := range []string{"api", "common", "fabric"} {
	// 	contents, err := filesystem.ReadFile(fmt.Sprintf("cue/%s.cue", file))
	// 	if err != nil {
	// 		return fmt.Errorf("failed to load fabric template file %s: %w", file, err)
	// 	}
	// 	data = append(data, string(contents))
	// }

	// v := cuedata.FromStrings(data...)
	// if err := v.Err(); err != nil {
	// 	cuedata.LogError(logger, err)
	// 	return fmt.Errorf("failed to load fabric templates")
	// }

	// value = &v

	return nil
}
