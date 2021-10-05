package version

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.version")
	//go:embed versions/*.cue
	filesystem embed.FS
)

func Load() (map[string]Version, error) {
	base, err := loadBase()
	if err != nil {
		return nil, err
	}

	versions, err := loadVersions(base)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

func loadBase() (cue.Value, error) {
	wd, err := os.Getwd()
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to determine working directory")
	}
	instances := load.Instances([]string{"greymatter.io/operator/cue.mod:base"}, &load.Config{
		Package:    "base",
		ModuleRoot: wd,
		Dir:        fmt.Sprintf("%s/cue.mod", wd),
	})
	base := cuecontext.New().BuildInstance(instances[0])
	if err := base.Err(); err != nil {
		return base, errors.Wrap(err.(errors.Error), fmt.Errorf("failed to load base install configuration module"))
	}

	logger.Info("Loaded base install configuration module")
	return base, nil
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
		version := Cue(string(data))
		if err := version.Err(); err != nil {
			logCueErrors(err)
			return nil, errors.Wrap(err.(errors.Error), fmt.Errorf("found invalid install configuration defined in %s", file.Name()))
		}

		// Unify version Cue value with base Cue value
		value := base.Unify(version)
		if err := value.Err(); err != nil {
			logCueErrors(err)
			return nil, errors.Wrap(err.(errors.Error), fmt.Errorf("found incompatible install configuration defined in %s", file.Name()))
		}

		name := strings.Replace(file.Name(), ".cue", "", 1)
		cueVersions[name] = Version{value}
		logger.Info("Loaded versioned install configuration", "name", name)
	}

	if len(cueVersions) == 0 {
		return nil, fmt.Errorf("no versioned install configurations were found")
	}

	return cueVersions, nil
}
