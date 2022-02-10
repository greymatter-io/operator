// Package installer exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a client for communicating with the cluster.
package installer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cuelang.org/go/cue"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	"github.com/greymatter-io/operator/pkg/version"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

	cs *cfsslsrv.CFSSLServer

	// The meshes.greymatter.io CRD, used as an owner when applying cluster-scoped resources.
	// If the operator is uninstalled on a cluster, owned cluster-scoped resources will be cleaned up.
	owner *extv1.CustomResourceDefinition
	// The Docker image pull secret to create in namespaces where core services are installed.
	imagePullSecret *corev1.Secret
	// The name of a configured cluster ingress name for OpenShift environments.
	clusterIngressName string
	// A map of namespaces -> mesh name.
	namespaces map[string]string
	// A map of mesh -> function that returns a sidecar (given an xdsCluster name), used for sidecar injection
	sidecars map[string]func(string) version.Sidecar
	// Base install configuration template for generating Grey Matter core service manifests
	baseTmpl cue.Value
	// The cluster ingress domain
	clusterIngressDomain string
}

// New returns a new *Installer instance for installing Grey Matter components and dependencies.
func New(c client.Client, load cuemodule.Loader, gmcli *cli.CLI, cs *cfsslsrv.CFSSLServer, clusterIngressName string) (*Installer, error) {
	baseTmpl, err := load("base")
	if err != nil {
		logger.Error(err, "Failed to load base install configuration templates")
		return nil, err
	}

	logger.Info("Loaded base install configuration templates")

	return &Installer{
		RWMutex:            &sync.RWMutex{},
		CLI:                gmcli,
		client:             c,
		cs:                 cs,
		clusterIngressName: clusterIngressName,
		namespaces:         make(map[string]string),
		sidecars:           make(map[string]func(string) version.Sidecar),
		baseTmpl:           baseTmpl,
	}, nil
}

// Start initializes resources and configurations after controller-manager has launched.
// It implements the controller-runtime Runnable interface.
func (i *Installer) Start(ctx context.Context) error {

	// Retrieve the operator image secret from the apiserver (block until it's retrieved).
	// This secret will be re-created in each install namespace and watch namespaces where core services are pulled.
	i.imagePullSecret = getImagePullSecret(i.client)

	// Get our Mesh CRD to set as an owner for cluster-scoped resources
	i.owner = &extv1.CustomResourceDefinition{}
	err := i.client.Get(ctx, client.ObjectKey{Name: "meshes.greymatter.io"}, i.owner)
	if err != nil {
		logger.Error(err, "Failed to get CustomResourceDefinition meshes.greymatter.io")
		return err
	}

	// Ensure our cluster-scoped RBAC permissions and SPIRE resources are created.
	applyClusterRBAC(i.client, i.owner)
	if i.cs != nil {
		applySpire(i.client, i.owner, i.cs)
	}

	// Try to get the OpenShift cluster ingress domain if it exists.
	clusterIngressDomain, ok := getOpenshiftClusterIngressDomain(i.client, i.clusterIngressName)
	if ok {
		// TODO: When not in OpenShift, check for other supported ingress class types such as Nginx or Voyager.
		// If no supported ingress types are found, just assume the user will configure ingress on their own.
		logger.Info("Identified OpenShift cluster domain name", "Domain", clusterIngressDomain)
		i.clusterIngressDomain = clusterIngressDomain
	}

	// Load existing meshes in the cluster
	if err := i.SyncMeshes(); err != nil {
		logger.Error(err, "Failed to sync existing meshes in cluster")
		return err
	}

	return nil
}

// SyncMeshes retrieves the list of existing meshes in the cluster,
// caches their sidecar templates and namespaces, and configures mesh clients.
// This essentially registers an existing mesh with the current (leader) pod.
func (i *Installer) SyncMeshes() error {
	meshList := &v1alpha1.MeshList{}
	if err := i.client.List(context.TODO(), meshList); err != nil {
		return err
	}

	for _, mesh := range meshList.Items {
		i.Lock()
		{
			// TODO (alec): this could get slow if we have a lot of meshes all unifying n number of times
			// with various options. CUE is fast so maybe this is fine but it's something to keep in mind.
			v, err := version.New(i.baseTmpl, &mesh, version.WithIngressSubDomain(i.clusterIngressDomain))
			if err != nil {
				return fmt.Errorf("failed to create version for mesh %s: %v", mesh.Name, err)
			}

			i.sidecars[mesh.Name] = v.SidecarTemplate()
			i.namespaces[mesh.Spec.InstallNamespace] = mesh.Name
			for _, namespace := range mesh.Spec.WatchNamespaces {
				i.namespaces[namespace] = mesh.Name
			}
		}
		i.Unlock()

		// We attempt an Apply here to statisfy the race condition of our server
		// missing CRD creation even though the webhooks are ready to accept config
		// If nothings changed this apply is ignored and we continue successfully.
		go i.ApplyMesh(nil, &mesh)
	}

	return nil
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

func applyClusterRBAC(c client.Client, crd *extv1.CustomResourceDefinition) {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-control"},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
	k8sapi.Apply(c, cr, crd, k8sapi.GetOrCreate)

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "gm-control"},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "gm-control",
		},
		Subjects: []rbacv1.Subject{},
	}
	k8sapi.Apply(c, crb, crd, k8sapi.GetOrCreate)
}

func getOpenshiftClusterIngressDomain(c client.Client, ingressName string) (string, bool) {
	clusterIngressList := &configv1.IngressList{}
	if err := c.List(context.TODO(), clusterIngressList); err != nil {
		return "", false
	} else {
		for _, i := range clusterIngressList.Items {
			if i.Name == ingressName {
				return i.Spec.Domain, true
			}
		}
	}
	return "", false
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
