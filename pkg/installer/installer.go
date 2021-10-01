// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/greymatter-io/operator/pkg/values"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.installer")
	scheme *runtime.Scheme
)

// Stores a map of versioned base values.Values and distinct proxy values.ContainerValues for each mesh.
type Installer struct {
	// A map of Grey Matter version (v*.*) -> *Values read from the filesystem.
	baseValues map[string]*values.Values
	// A map of meshes -> *ContainerValues for proxy templates, used for sidecar injection
	proxyValues map[string]*values.ContainerValues
}

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New(runtimeScheme *runtime.Scheme) (*Installer, error) {
	scheme = runtimeScheme

	baseValues, err := loadVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to start Installer: %w", err)
	}

	return &Installer{
		baseValues:  baseValues,
		proxyValues: make(map[string]*values.ContainerValues),
	}, nil
}

func loadVersions() (map[string]*values.Values, error) {
	yamls, err := values.LoadYAMLVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML for installation templates: %w", err)
	}

	versions := make(map[string]*values.Values)
	var logged []string

YAML_LOOP:
	for name, y := range yamls {
		iv := &values.Values{}
		if err := yaml.Unmarshal(y, iv); err != nil {
			logger.Error(err, "failed to unmarshal YAML; unable to load installation templates", "version", name)
			continue YAML_LOOP
		} else {
			versions[name] = iv
			logged = append(logged, name)
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no valid installation templates were loaded")
	}

	logger.Info("loaded Grey Matter installation templates", "versions", logged)

	return versions, nil
}
