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
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.installer")
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
	baseValues, err := loadVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to load install values: %w", err)
	}

	return &Installer{
		scheme:      scheme,
		baseValues:  baseValues,
		proxyValues: make(map[string]*values.ContainerValues),
	}, nil
}

func loadFiles() ([]fs.DirEntry, error) {
	files, err := filesystem.ReadDir("versions")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded version files: %w", err)
	}

	return files, nil
}

func loadVersions() (map[string]*values.Values, error) {
	files, err := loadFiles()
	if err != nil {
		return nil, fmt.Errorf("unable to load values: %w", err)
	}

	versions := make(map[string]*values.Values)

FILE_LOOP:
	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".yaml") {
			logger.Error(fmt.Errorf("detected version file with invalid extension (expected .yaml)"), "skipping", "filename", fileName)
			continue FILE_LOOP
		}
		name := strings.Replace(fileName, ".yaml", "", 1)
		data, _ := filesystem.ReadFile(fmt.Sprintf("versions/%s", fileName))
		iv := &values.Values{}
		if err := yaml.Unmarshal(data, iv); err != nil {
			logger.Error(err, "failed to unmarshal YAML from file", "filename", fileName)
			continue FILE_LOOP
		} else {
			versions[name] = iv
		}
	}

	return versions, nil
}
