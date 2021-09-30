// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/greymatter-io/operator/pkg/values"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Stores a map of versioned base values.Values and distinct proxy values.ContainerValues for each mesh.
type Installer struct {
	// The scheme used by the operator, for adding controller references to manifests.
	scheme *runtime.Scheme
	// A map of Grey Matter version (v*.*) -> *Values read from the filesystem.
	baseValues map[string]*values.Values
	// A map of meshes -> *ContainerValues for proxy templates, used for sidecar injection
	proxyValues map[string]*values.ContainerValues
}

//go:embed versions/*.yaml
var filesystem embed.FS

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New(scheme *runtime.Scheme) (*Installer, error) {

	// TODO: Allow the user to specify a directory for mounting new Values files.
	// This of course will mean we can't use embed.FS since those are compile-time embeds.
	// User-provided values files must be validated on startup (extract validation from versions_test.go).
	files, err := filesystem.ReadDir("versions")
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
		proxyValues: make(map[string]*values.ContainerValues),
	}, nil
}

func loadBaseValues(files []fs.DirEntry) (map[string]*values.Values, error) {
	templates := make(map[string]*values.Values)

	for _, file := range files {
		fileName := file.Name()
		name := strings.Replace(fileName, ".yaml", "", 1)
		data, _ := filesystem.ReadFile(fmt.Sprintf("values/%s", fileName))
		iv := &values.Values{}
		if err := yaml.Unmarshal(data, iv); err != nil {
			return nil, fmt.Errorf("failed to parse YAML from file %s: %w", fileName, err)
		} else {
			templates[name] = iv
		}
	}

	return templates, nil
}
