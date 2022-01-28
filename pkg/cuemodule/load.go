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

// Loader loads a package from our CUE module.
type Loader func(string) (cue.Value, error)

// LoadPackage loads a package from our Cue module.
// Packages are added to subdirectories and declared with the same name as the subdirectory.
func LoadPackage(pkgName string) (cue.Value, error) {
	dirPath, err := os.Getwd()
	if err != nil {
		return cue.Value{}, err
	}

	return loadPackage(pkgName, dirPath)
}

// LoadPackageForTest loads a package from our Cue module within a test context.
func LoadPackageForTest(pkgName string) (cue.Value, error) {
	_, filename, _, _ := runtime.Caller(0)
	dirPath := path.Dir(filename)

	return loadPackage(pkgName, dirPath)
}

func loadPackage(pkgName, dirPath string) (cue.Value, error) {
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
