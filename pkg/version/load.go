package version

import (
	"fmt"
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// Load reads in the base CUE module for the component version package.
// If not successful, it returns an empty cue.Value{}.
// Otherwise the loaded cue.Value is returned with no error.
func Load(pathElems ...string) (cue.Value, error) {
	base, err := loadBase(pathElems)
	if err != nil {
		return cue.Value{}, err
	}
	return base, nil
}

func loadBase(pathElems []string) (cue.Value, error) {
	var dirPath string
	if len(pathElems) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return cue.Value{}, fmt.Errorf("failed to determine working directory")
		}
		dirPath = wd
	} else {
		dirPath = path.Join(pathElems...)
	}
	instances := load.Instances([]string{"greymatter.io/operator/version/cue.mod:base"}, &load.Config{
		Package:    "base",
		ModuleRoot: dirPath,
		Dir:        fmt.Sprintf("%s/cue.mod", dirPath),
	})
	base := cuecontext.New().BuildInstance(instances[0])
	if err := base.Err(); err != nil {
		return base, err
	}

	logger.Info("Loaded base install configuration module")
	return base, nil
}
