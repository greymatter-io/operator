package version

import (
	"fmt"

	"github.com/greymatter-io/operator/pkg/cuedata"
)

var versionPackages = map[string]string{
	"onesix":   "1.6",
	"oneseven": "1.7",
}

func Load() (map[string]Version, error) {
	versions, err := loadBaseWithVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to load versioned install configurations: %w", err)
	}
	return versions, nil
}

func loadBaseWithVersions() (map[string]Version, error) {
	versions := make(map[string]Version)
	for k, v := range versionPackages {
		cv, err := cuedata.LoadPackages("base", k)
		if err != nil {
			return nil, err
		}
		versions[v] = Version{v, cv}
		logger.Info("Loaded versioned install configuration", "name", v)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versioned install configurations were found")
	}

	return versions, nil
}
