// Package gmcore exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package gmcore

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Stores a map of Grey Matter InstallValues and a reference from each mesh to a version
type Installer struct {
	// The scheme for this program, used for setting controller references on resources.
	scheme *runtime.Scheme
	// A map of Grey Matter version (v*.*) -> *InstallValues read from the filesystem.
	baseValues map[string]*v1alpha1.InstallValues
	// A map of meshes -> *Values for proxy templates, used for sidecar injection
	proxyValues map[string]*v1alpha1.Values
}

//go:embed values/*.yaml
var filesystem embed.FS

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New(scheme *runtime.Scheme) (*Installer, error) {

	// TODO: Allow the user to specify a directory for mounting new values files.
	// Later on, let the user define each InstallationConfig custom resource via apiserver.
	files, err := filesystem.ReadDir("values")
	if err != nil {
		return nil, fmt.Errorf("failed to embed files into program: %w", err)
	}

	baseValues, err := loadBaseValues(files)
	if err != nil {
		return nil, fmt.Errorf("failed to load install values: %w", err)
	}

	return &Installer{
		scheme:      scheme,
		baseValues:  baseValues,
		proxyValues: make(map[string]*v1alpha1.Values),
	}, nil
}

func loadBaseValues(files []fs.DirEntry) (map[string]*v1alpha1.InstallValues, error) {
	templates := make(map[string]*v1alpha1.InstallValues)

	for _, file := range files {
		fileName := file.Name()
		data, _ := filesystem.ReadFile(fmt.Sprintf("values/%s", fileName))
		cfg := &v1alpha1.InstallationConfig{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML from file %s: %w", fileName, err)
		} else {
			templates[cfg.Name] = &cfg.InstallValues
		}
	}

	return templates, nil
}
