package version

import (
	"embed"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/errors"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/cueutils"
)

var (
	//go:embed versions/*.cue
	filesystem embed.FS
)

func Load(pathElems ...string) (map[string]Version, error) {
	versions, err := loadBaseWithVersions(pathElems)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func loadBaseWithVersions(pathElems []string) (map[string]Version, error) {
	base, err := loadBase(pathElems)
	if err != nil {
		return nil, err
	}

	versions, err := loadVersions(base)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

func loadBase(pathElems []string) (cue.Value, error) {
	v, err := cuemodule.LoadPackage("base")
	if err != nil {
		return cue.Value{}, err
	}

	logger.Info("Loaded base install configuration module")
	return v, nil
}

func loadVersions(base cue.Value) (map[string]Version, error) {
	files, err := filesystem.ReadDir("versions")
	if err != nil {
		return nil, fmt.Errorf("failed to load versioned install configurations")
	}

	cueVersions := make(map[string]Version)
	for _, file := range files {
		data, err := filesystem.ReadFile(fmt.Sprintf("versions/%s", file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to load versioned install configuration from %s: %w", file.Name(), err)
		}

		// Build Cue value from version file
		version := cueutils.FromStrings(string(data))
		if err := version.Err(); err != nil {
			cueutils.LogError(logger, err)
			return nil, errors.Wrap(err.(errors.Error), fmt.Errorf("found invalid install configuration defined in %s", file.Name()))
		}

		// Unify version Cue value with base Cue value
		value := base.Unify(version)
		if err := value.Err(); err != nil {
			cueutils.LogError(logger, err)
			return nil, errors.Wrap(err.(errors.Error), fmt.Errorf("found incompatible install configuration defined in %s", file.Name()))
		}

		name := strings.Replace(file.Name(), ".cue", "", 1)
		cueVersions[name] = Version{name, value}
		logger.Info("Loaded versioned install configuration", "name", name)
	}

	if len(cueVersions) == 0 {
		return nil, fmt.Errorf("no versioned install configurations were found")
	}

	return cueVersions, nil
}
