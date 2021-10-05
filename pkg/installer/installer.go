// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"github.com/greymatter-io/operator/pkg/version"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("pkg.installer")
)

// Stores a map of version.Version and a distinct version.Sidecar for each mesh.
type Installer struct {
	client client.Client
	// A map of Grey Matter version (v*.*) -> Version read from the filesystem.
	versions map[string]version.Version
	// A map of meshes -> Sidecar, used for sidecar injection
	sidecars map[string]version.Sidecar
}

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New(c client.Client) (*Installer, error) {
	versions, err := version.Load()
	if err != nil {
		logger.Error(err, "Failed to initialize installer")
		return nil, err
	}

	return &Installer{
		client:   c,
		versions: versions,
		sidecars: make(map[string]version.Sidecar),
	}, nil
}
