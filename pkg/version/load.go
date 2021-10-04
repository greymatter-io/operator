package version

import (
	"embed"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
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

	var errs errors.Error
	for name, version := range versions {
		version.cue = base.Unify(version.cue)
		if err := version.cue.Err(); err != nil {
			logger.Error(err, "found invalid install version file", "filename", fmt.Sprintf("%s.cue", name))
			if errs == nil {
				errs = errors.Errors(err)[0]
			} else {
				errs = errors.Wrap(errs, err)
			}
			delete(versions, name)
		} else {
			versions[name] = version
		}
	}

	return versions, errs
}

func loadBase() (cue.Value, error) {
	instances := load.Instances(
		[]string{"./cue.mod/"},
		&load.Config{Package: "base"},
	)
	base := cuecontext.New().BuildInstance(instances[0])
	if err := base.Err(); err != nil {
		return base, base.Err()
	}
	return base, nil
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

		value := Cue(string(data))
		if err := value.Err(); err != nil {
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
