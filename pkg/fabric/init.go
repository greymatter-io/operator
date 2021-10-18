package fabric

import (
	"embed"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/greymatter-io/operator/pkg/cueutils"
)

var (
	//go:embed cue/*.cue
	filesystem embed.FS
	value      *cue.Value
)

func Init() error {
	var data []string

	for _, file := range []string{"api", "common", "fabric"} {
		contents, err := filesystem.ReadFile(fmt.Sprintf("cue/%s.cue", file))
		if err != nil {
			return fmt.Errorf("failed to load fabric template file %s: %w", file, err)
		}
		data = append(data, string(contents))
	}

	v := cueutils.FromStrings(data...)
	if err := v.Err(); err != nil {
		cueutils.LogError(logger, err)
		return fmt.Errorf("failed to load fabric templates")
	}

	value = &v

	return nil
}
