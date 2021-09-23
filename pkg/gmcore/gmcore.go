// Package gmcore exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package gmcore

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Stores a map of Grey Matter InstallValuesConfig and a reference from each mesh to a version
type Installer struct {
	// A map of Grey Matter version (v*.*) -> InstallValuesConfig read from the filesystem.
	values map[string]*v1alpha1.InstallValuesConfig
	// A map of meshes referencing a Grey Matter version.
	meshes map[string]string
}

//go:embed values/*.yaml
var filesystem embed.FS

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New() (*Installer, error) {

	// TODO: Allow the user to specify a directory for mounting new values files.
	// Later on, let the user define each InstallValuesConfig custom resource via apiserver.
	files, err := filesystem.ReadDir("values")
	if err != nil {
		return nil, fmt.Errorf("failed to embed files into program: %w", err)
	}

	values, err := loadValues(files)
	if err != nil {
		return nil, fmt.Errorf("failed to load install values: %w", err)
	}

	return &Installer{
		values: values,
		meshes: make(map[string]string),
	}, nil
}

func loadValues(files []fs.DirEntry) (map[string]*v1alpha1.InstallValuesConfig, error) {
	templates := make(map[string]*v1alpha1.InstallValuesConfig)

	for _, file := range files {
		fileName := file.Name()
		data, _ := filesystem.ReadFile(fmt.Sprintf("values/%s", fileName))
		cfg := &v1alpha1.InstallValuesConfig{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML from file %s: %w", fileName, err)
		} else {
			templates[cfg.Name] = cfg
		}
	}

	return templates, nil
}
