package cuemodule

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

var dirPath string

// Initialize the path to our Cue module directory.
// We expect the working directory in the container to be `/app`.
// If not running from `/app`, this is in a unit test which needs the runtime file path.
func init() {
	dirPath, _ = os.Getwd()
	if dirPath != "/app" {
		_, filename, _, _ := runtime.Caller(0)
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
