// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/version"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("installer")
)

// Installer stores a map of version.Version and a distinct version.Sidecar for each mesh.
type Installer struct {
	*sync.RWMutex
	*cli.CLI

	client client.Client
	// The Docker image pull secret to create in namespaces where core services are installed.
	imagePullSecret *corev1.Secret
	// A map of Grey Matter version (v*.*) -> Version read from the filesystem.
	versions map[string]version.Version
	// A map of namespaces -> mesh name.
	namespaces map[string]string
	// A map of mesh -> function that returns a sidecar (given an xdsCluster name), used for sidecar injection
	sidecars map[string]func(string) version.Sidecar
	// The cluster ingress domain
	clusterIngressDomain string
}

// New returns a new *Installer instance for installing Grey Matter components and dependencies.
func New(c client.Client, gmcli *cli.CLI, ingressName string) (*Installer, error) {
	versions, err := version.Load()
	if err != nil {
		logger.Error(err, "Failed to initialize installer")
		return nil, err
	}

	i := &Installer{
		RWMutex:    &sync.RWMutex{},
		CLI:        gmcli,
		client:     c,
		versions:   versions,
		namespaces: make(map[string]string),
		sidecars:   make(map[string]func(string) version.Sidecar),
	}

	// Copy the image pull secret from the apiserver (block until it's retrieved).
	// This secret will be re-created in each install namespace where our core services are pulled.
	i.imagePullSecret = getImagePullSecret(c)

	// Try to get the OpenShift cluster ingress domain if it exists.
	// TODO: When not in OpenShift, check for other supported ingress class types such as Nginx or Voyager.
	// If no supported ingress types are found, just assume the user will configure ingress on their own.
	clusterIngressDomain, err := getOpenshiftClusterIngressDomain(c, ingressName)
	if err != nil {
		return i, nil
	}

	i.clusterIngressDomain = clusterIngressDomain
	return i, nil
}

// Retrieves the image pull secret in the gm-operator namespace.
// This retries indefinitely at 30s intervals and will block by design.
func getImagePullSecret(c client.Client) *corev1.Secret {
	key := client.ObjectKey{Name: "gm-docker-secret", Namespace: "gm-operator"}
	operatorSecret := &corev1.Secret{}
	for operatorSecret.CreationTimestamp.IsZero() {
		if err := c.Get(context.TODO(), key, operatorSecret); err != nil {
			logger.Error(err, "No 'gm-docker-secret' image pull secret found in gm-operator namespace. Will retry in 30s.")
			time.Sleep(time.Second * 30)
		}
	}

	// Return new secret with just the dockercfgjson (without additional metadata).
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-docker-secret"},
		Type:       operatorSecret.Type,
		Data:       operatorSecret.Data,
	}
}

func getOpenshiftClusterIngressDomain(c client.Client, ingressName string) (string, error) {
	clusterIngressList := &configv1.IngressList{}
	if err := c.List(context.TODO(), clusterIngressList); err != nil {
		return "", err
	} else {
		for _, i := range clusterIngressList.Items {
			if i.Name == ingressName {
				logger.Info("Identified OpenShift cluster domain", "Host", i.Spec.Domain)
				return i.Spec.Domain, nil
			}
		}
	}
	return "", fmt.Errorf("found cluster list however specified cluster ingress name [%s] not found", ingressName)
}

// Check that a suported ingress controller class exists in a kubernetes cluster.
// This will be expanded later on as we support additional ingress implementations.
//lint:ignore U1000 save for reference
func isSupportedKubernetesIngressClassPresent(c client.Client) bool {
	ingressClassList := &networkingv1.IngressClassList{}
	if err := c.List(context.TODO(), ingressClassList); err != nil {
		return false
	}
	for _, i := range ingressClassList.Items {
		switch i.Spec.Controller {
		case "nginx", "voyager":
			return true
		}
	}
	return false
}
