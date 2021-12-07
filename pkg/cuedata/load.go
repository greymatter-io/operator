package cuedata

import (
	"fmt"
	"path"
	"runtime"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("cuedata")
)

func LoadPackages(names ...string) (cue.Value, error) {
	if len(names) == 0 {
		return cue.Value{}, fmt.Errorf("at least one package name argument required")
	}

	// Get this file's path at runtime in order to determine our Cue module directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return cue.Value{}, fmt.Errorf("failed to retrieve path to Cue modules")
	}
	dirPath := path.Dir(filename)

	var pkgs []string
	for _, name := range names {
		pkgs = append(pkgs, "greymatter.io/operator/cue.mod:"+name)
	}

	instances := load.Instances(pkgs, &load.Config{
		ModuleRoot: dirPath,
		Dir:        dirPath + "/cue.mod",
	})

	vs, err := cuecontext.New().BuildInstances(instances)
	if err != nil {
		return cue.Value{}, err
	}

	var loaded cue.Value
	for _, v := range vs {
		loaded = loaded.Unify(v)
	}

	if err := loaded.Err(); err != nil {
		return cue.Value{}, err
	}

	return loaded, nil
}
