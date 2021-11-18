package config

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

//go:embed *
var configFS embed.FS

func LoadManifests() (string, error) {
	kfs, err := mkKyamlFileSys(configFS)
	if err != nil {
		return "", fmt.Errorf("failed to populate in-memory file system: %w", err)
	}

	opts := krusty.MakeDefaultOptions()
	opts.DoLegacyResourceSort = true

	k := krusty.MakeKustomizer(opts)
	res, err := k.Run(kfs, "context/kubernetes")
	if err != nil {
		return "", fmt.Errorf("failed to perform kustomization: %w", err)
	}

	yml, err := res.AsYaml()
	if err != nil {
		return "", fmt.Errorf("failed to parse as yaml")
	}

	return string(yml), nil
}

func mkKyamlFileSys(efs embed.FS) (filesys.FileSystem, error) {
	kfs := filesys.MakeFsInMemory()
	loadFunc := mkFileLoader(efs, kfs)
	if err := fs.WalkDir(efs, ".", loadFunc); err != nil {
		return kfs, err
	}
	return kfs, nil
}

func mkFileLoader(efs embed.FS, kfs filesys.FileSystem) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		return kfs.WriteFile(path, data)
	}
}
