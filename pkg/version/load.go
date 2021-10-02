package version

import (
	"embed"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.installvalues")
	//go:embed */*.cue
	filesystem embed.FS
)

func Load() (map[string]Version, error) {
	base, err := loadBase()
	if err != nil {
		return nil, err
	}
	if err := base.Err(); err != nil {
		return nil, err
	}

	versions, err := loadVersions()
	if err != nil {
		return nil, err
	}

	for name, version := range versions {
		version.cv = base.Unify(version.cv)
		if err := version.cv.Err(); err != nil {
			logger.Error(err, "found invalid install version file", "filename", fmt.Sprintf("%s.cue", name))
			delete(versions, name)
		} else {
			versions[name] = version
		}
	}

	return versions, nil
}

func loadBase() (cue.Value, error) {
	files, err := filesystem.ReadDir("base")
	if err != nil {
		logger.Error(err, "failed to load install definition files")
		return cue.Value{}, err
	}

	var baseSchemas []string
	for _, file := range files {
		data, err := filesystem.ReadFile(fmt.Sprintf("base/%s", file.Name()))
		if err != nil {
			logger.Error(err, "failed to load install definition file", "filename", file.Name())
			return cue.Value{}, err
		}
		baseSchemas = append(baseSchemas, string(data))
	}

	value, err := CueFromStrings(baseSchemas...)
	if err != nil {
		logger.Error(err, "failed to parse install definition files")
		return cue.Value{}, err
	}

	return value, nil
}

func loadVersions() (map[string]Version, error) {
	files, err := filesystem.ReadDir("versions")
	if err != nil {
		logger.Error(err, "failed to load install version files")
		return nil, err
	}

	cueVersions := make(map[string]Version)
	for _, file := range files {
		data, err := filesystem.ReadFile(fmt.Sprintf("versions/%s", file.Name()))
		if err != nil {
			logger.Error(err, "failed to load install version file", "filename", file.Name())
			return nil, err
		}

		value, err := CueFromStrings(string(data))
		if err != nil {
			logger.Error(err, "failed to parse install version file", "filename", file.Name())
			return nil, err
		}

		name := strings.Replace(file.Name(), ".cue", "", 1)

		cueVersions[name] = Version{value}
	}

	if len(cueVersions) == 0 {
		logger.Error(err, "no install version files loaded")
		return nil, err
	}

	return cueVersions, err
}
