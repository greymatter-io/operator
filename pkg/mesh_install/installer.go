// Package mesh_install exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a k8sClient for communicating with the cluster.
package mesh_install

import (
	"context"
	"github.com/cloudflare/cfssl/csr"
	configv1 "github.com/openshift/api/config/v1"
	"strings"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/gmapi"
	"github.com/greymatter-io/operator/pkg/k8sapi"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("mesh_install")
)

// Installer stores a map of version.Version and a distinct version.Sidecar for each mesh.
type Installer struct {
	*gmapi.CLI // Grey Matter CLI
	k8sClient  *client.Client

	cfssl *cfsslsrv.CFSSLServer

	// The meshes.greymatter.io CRD, used as an owner when applying cluster-scoped resources.
	// If the operator is uninstalled on a cluster, owned cluster-scoped resources will be cleaned up.
	owner *extv1.CustomResourceDefinition
	// The Docker image pull secret to create in namespaces where core services are installed.
	imagePullSecret *corev1.Secret

	// Container for THE mesh (on the way to an experimental 1:1 operator:mesh paradigm)
	// Contains the default after load
	Mesh *v1alpha1.Mesh

	// Container for all K8s and GM CUE cue.Values
	OperatorCUE *cuemodule.OperatorCUE

	// Root on disk of the operator CUE. Used for reloading the default configs on teardown
	CueRoot string

	// Operator config loadable from CUE
	Config cuemodule.Config

	// Looked up on start
	clusterIngressDomain string
}

// New returns a new *Installer instance for installing Grey Matter components and dependencies.
func New(c *client.Client, operatorCUE *cuemodule.OperatorCUE, initialMesh *v1alpha1.Mesh, cueRoot string, gmcli *gmapi.CLI, cfssl *cfsslsrv.CFSSLServer) (*Installer, error) {
	config, defaults := operatorCUE.ExtractFlagsAndDefaults()
	gmcli.SidecarList = defaults.RedisSpireSubjects
	return &Installer{
		CLI:         gmcli,
		k8sClient:   c,
		cfssl:       cfssl,
		OperatorCUE: operatorCUE,
		Mesh:        initialMesh,
		CueRoot:     cueRoot,
		Config:      config,
	}, nil
}

// Start initializes resources and configurations after controller-manager has launched.
// It implements the controller-runtime Runnable interface.
func (i *Installer) Start(ctx context.Context) error {

	// Retrieve the operator image secret from the apiserver (block until it's retrieved).
	// This secret will be re-created in each install namespace and watch namespaces where core services are pulled.
	i.imagePullSecret = getImagePullSecret(i.k8sClient)

	// Get our Mesh CRD to set as an owner for cluster-scoped resources
	i.owner = &extv1.CustomResourceDefinition{}
	err := (*i.k8sClient).Get(ctx, client.ObjectKey{Name: "meshes.greymatter.io"}, i.owner)
	if err != nil {
		logger.Error(err, "Failed to get CustomResourceDefinition meshes.greymatter.io")
		return err
	}

	logger.Info("Attempting to apply spire server-ca secret")
	spireSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "server-ca",
			Namespace: "spire",
		},
	}

	// TODO the operator itself can launch _all_ the Spire manifests, as before, after a check that they aren't already there.
	if i.Config.Spire {
		spireSecret, err = injectPKI(spireSecret, i.cfssl)
		if err != nil {
			logger.Error(err, "Error while attempting to apply spire server-ca secret", "secret object", spireSecret)
			return err
		}
		k8sapi.Apply(i.k8sClient, spireSecret, i.owner, k8sapi.CreateOrUpdate)
	}

	// Try to get the OpenShift cluster ingress domain if it exists.
	clusterIngressDomain, ok := getOpenshiftClusterIngressDomain(i.k8sClient, i.Config.ClusterIngressName)
	if ok {
		// TODO: When not in OpenShift, check for other supported ingress class types such as Nginx or Voyager.
		// If no supported ingress types are found, just assume the user will configure ingress on their own.
		logger.Info("Identified OpenShift cluster domain name", "Domain", clusterIngressDomain)
		i.clusterIngressDomain = clusterIngressDomain
	}

	// Immediately apply the default mesh from the CUE if the flag is set
	go func() {
		if i.Config.AutoApplyMesh {
			time.Sleep(100 * time.Second) // DEBUG - does this work if we wait long enough for it to register the webhooks?
			k8sapi.Apply(i.k8sClient, i.Mesh, nil, k8sapi.CreateOrUpdate)
		}
	}()
	//// Look it back up to get its ID (for use as an owner of other resources)
	//meshList := &v1alpha1.MeshList{}
	//if err := i.k8sClient.List(context.TODO(), meshList); err != nil {
	//	return err
	//}
	//if len(meshList.Items) > 0 {
	//	for _, mesh := range meshList.Items {
	//		i.Mesh = &mesh
	//	}
	//	go i.ApplyMesh(nil, i.Mesh)
	//}

	return nil
}

// Retrieves the image pull secret in the gm-operator namespace.
// This retries indefinitely at 30s intervals and will block by design.
func getImagePullSecret(c *client.Client) *corev1.Secret {
	key := client.ObjectKey{Name: "gm-docker-secret", Namespace: "gm-operator"}
	operatorSecret := &corev1.Secret{}
	for operatorSecret.CreationTimestamp.IsZero() {
		if err := (*c).Get(context.TODO(), key, operatorSecret); err != nil {
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

func getOpenshiftClusterIngressDomain(c *client.Client, ingressName string) (string, bool) {
	clusterIngressList := &configv1.IngressList{}
	if err := (*c).List(context.TODO(), clusterIngressList); err != nil {
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

func injectPKI(secret *corev1.Secret, cs *cfsslsrv.CFSSLServer) (*corev1.Secret, error) {
	root := cs.GetRootCA()
	ca, caKey, err := cs.RequestIntermediateCA(csr.CertificateRequest{
		CN:         "Grey Matter SPIFFE Intermediate CA",
		KeyRequest: &csr.KeyRequest{A: "ecdsa", S: 256},
		Names: []csr.Name{
			{C: "US", ST: "VA", L: "Alexandria", O: "Grey Matter"},
		},
	})
	if err != nil {
		return nil, err
	}

	secret.StringData = map[string]string{
		"root.crt":         string(root),
		"intermediate.crt": strings.Join([]string{string(ca), string(root)}, "\n"),
		"intermediate.key": string(caKey),
	}

	return secret, nil
}
