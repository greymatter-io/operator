package config

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

func TestLoadManifests(t *testing.T) {
	manifests, err := LoadManifests()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(manifests)
}

func TestMkKyamlFileSys(t *testing.T) {
	kfs, err := mkKyamlFileSys(configFS)
	if err != nil {
		t.Fatal(err)
	}

	// Walk the embed.FS and ensure all YAML files exist in the kyaml file system
	err = fs.WalkDir(configFS, ".", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".yaml") {
			if !kfs.Exists(path) {
				return fmt.Errorf("kfs does not have file %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	// Walk the kyaml file system and ensure all files exist in embed.FS
	err = kfs.Walk("/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if _, err := configFS.Open(strings.TrimPrefix(path, "/")); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
