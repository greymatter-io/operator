// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"context"
	"time"

	"github.com/greymatter-io/operator/pkg/version"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("pkg.installer")
)

// Stores a map of version.Version and a distinct version.Sidecar for each mesh.
type Installer struct {
	client client.Client
	// The Docker image pull secret to create in namespaces where core services are installed.
	imagePullSecret *corev1.Secret
	// The service account to create in namespaces to be used by Control for watching pods
	serviceAccount *corev1.ServiceAccount
	// A map of Grey Matter version (v*.*) -> Version read from the filesystem.
	versions map[string]version.Version
	// A map of meshes -> Sidecar, used for sidecar injection
	sidecars map[string]version.Sidecar
}

// Returns *Installer for tracking which Grey Matter version is installed for each mesh
func New(c client.Client, imagePullSecretName string) (*Installer, error) {
	versions, err := version.Load()
	if err != nil {
		logger.Error(err, "Failed to initialize installer")
		return nil, err
	}

	installer := &Installer{
		client:   c,
		versions: versions,
		sidecars: make(map[string]version.Sidecar),
	}

	// Copy the image pull secret from the apiserver (block until it's retrieved).
	// This secret will be re-created in each install namespace where our core services are pulled.
	installer.cacheImagePullSecret(c, imagePullSecretName)

	// Copy the pod watcher service account from the apiserver. This will be created by OLM.
	// This service account will be re-created in each install namespace to be used by Control.
	if err := installer.cacheServiceAccount(c); err != nil {
		logger.Error(err, "Failed to initialize installer")
		return nil, err
	}

	return installer, nil
}

// Retrieves the image pull secret in the gm-operator namespace (default name is gm-docker-secret).
// This retries indefinitely at 30s intervals and will block by design.
func (i *Installer) cacheImagePullSecret(c client.Client, imagePullSecretName string) {
	key := client.ObjectKey{Name: imagePullSecretName, Namespace: "gm-operator"}
	operatorSecret := &corev1.Secret{}
	for operatorSecret.CreationTimestamp.IsZero() {
		if err := c.Get(context.TODO(), key, operatorSecret); err != nil {
			logger.Error(err, "No image pull secret found in gm-operator namespace. Will retry in 30s.", "Name", imagePullSecretName)
			time.Sleep(time.Second * 30)
		}
	}

	// Store a new secret with just the dockercfgjson (without additional metadata).
	i.imagePullSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: imagePullSecretName},
		Type:       operatorSecret.Type,
		Data:       operatorSecret.Data,
	}
}

func (i *Installer) cacheServiceAccount(c client.Client) error {
	key := client.ObjectKey{Name: "gm-control", Namespace: "gm-operator"}
	sa := &corev1.ServiceAccount{}
	if err := c.Get(context.TODO(), key, sa); err != nil {
		return err
	}

	i.serviceAccount = &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: sa.Name},
		AutomountServiceAccountToken: func() *bool {
			b := true
			return &b
		}(),
	}

	return nil
}
