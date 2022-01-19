package cuemodule

import (
	"fmt"
	"path"
	"runtime"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

var dirPath string

// Initialize the path to our Cue module directory (i.e. this directory).
func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Errorf("failed to retrieve path to Cue module"))
	} else {
		dirPath = path.Dir(filename)
	}
}

// Loads a package from our Cue module.
// Packages are added to subdirectories and declared with the same name as the subdirectory.
func LoadPackage(pkgName string) (cue.Value, error) {
	instances := load.Instances([]string{"greymatter.io/operator/" + pkgName}, &load.Config{
		ModuleRoot: dirPath,
	})

	if len(instances) != 1 {
		return cue.Value{}, fmt.Errorf("did not load expected package %s", pkgName)
	}

	value := cuecontext.New().BuildInstance(instances[0])
	if err := value.Err(); err != nil {
		return cue.Value{}, err
	}

	return value, nil
}
