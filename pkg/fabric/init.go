package fabric

import (
	"embed"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"github.com/greymatter-io/operator/pkg/cueutils"
)

var (
	//go:embed cue/*.cue
	filesystem embed.FS
	value      *cue.Value
)

// Init reads embedded Cue files in order to load fabric object templates.
// It should be called on startup of the operator.
func Init() error {
	var data []string

	for _, file := range []string{"api", "fabric", "spire"} {
		contents, err := filesystem.ReadFile(fmt.Sprintf("cue/%s.cue", file))
		if err != nil {
			return fmt.Errorf("failed to load fabric template file %s: %w", file, err)
		}
		// Remove the package declaration from each file (the first line).
		// Otherwise, CUE's API won't be able to unify the files into a single value.
		contentsStr := string(contents)
		contentsStr = contentsStr[strings.Index(contentsStr, "\n"):]
		data = append(data, contentsStr)
	}

	v := cueutils.FromStrings(data...)
	if err := v.Err(); err != nil {
		cueutils.LogError(logger, err)
		return fmt.Errorf("failed to load fabric templates")
	}

	value = &v

	return nil
}
