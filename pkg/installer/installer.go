// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"github.com/greymatter-io/operator/pkg/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.installer")
	scheme *runtime.Scheme
)

// Stores a map of version.Version and a distinct proxy corev1.Container for each mesh.
type Installer struct {
	// A map of Grey Matter version (v*.*) -> Version read from the filesystem.
	versions map[string]version.Version
	// A map of meshes -> corev1.Container, used for sidecar injection
	sidecars map[string]*corev1.Container
}

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New(runtimeScheme *runtime.Scheme) (*Installer, error) {
	scheme = runtimeScheme

	versions, err := version.Load()
	if err != nil {
		logger.Error(err, "failed to start Installer")
		return nil, err
	}

	return &Installer{
		versions: versions,
		sidecars: make(map[string]*corev1.Container),
	}, nil
}
