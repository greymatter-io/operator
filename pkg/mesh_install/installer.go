// Package mesh_install exposes functions for applying resources to a Kubernetes cluster.
// Its exposed functions receive a K8sClient for communicating with the cluster.
package mesh_install

import (
	"context"
	"encoding/json"
	"github.com/cloudflare/cfssl/csr"
	"github.com/greymatter-io/operator/pkg/wellknown"
	configv1 "github.com/openshift/api/config/v1"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/gmapi"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	"github.com/greymatter-io/operator/pkg/sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	K8sClient  *client.Client

	cfssl *cfsslsrv.CFSSLServer

	// The meshes.greymatter.io CRD, used as an owner when applying cluster-scoped resources.
	// If the operator is uninstalled on a cluster, owned cluster-scoped resources will be cleaned up.
	owner *appsv1.StatefulSet
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

	// Select defaults that may be directly overridden from Go
	Defaults cuemodule.Defaults

	// Looked up on start
	clusterIngressDomain string

	// Sync configuration with access to a callback for updating on git repo changes
	Sync *sync.Sync
}

// New returns a new *Installer instance for installing Grey Matter components and dependencies.
func New(c *client.Client, operatorCUE *cuemodule.OperatorCUE, initialMesh *v1alpha1.Mesh, cueRoot string, gmcli *gmapi.CLI, cfssl *cfsslsrv.CFSSLServer, sync *sync.Sync) (*Installer, error) {
	config, defaults := operatorCUE.ExtractConfig()
	return &Installer{
		CLI:         gmcli,
		K8sClient:   c,
		cfssl:       cfssl,
		OperatorCUE: operatorCUE,
		Mesh:        initialMesh,
		CueRoot:     cueRoot,
		Config:      config,
		Defaults:    defaults,
		Sync:        sync,
	}, nil
}

// Start initializes resources and configurations after controller-manager has launched.
// It implements the controller-runtime Runnable interface.
func (i *Installer) Start(ctx context.Context) error {

	// Retrieve the operator image secret from the apiserver (block until it's retrieved).
	// This secret will be re-created in each install namespace and watch namespaces where core services are pulled.
	i.imagePullSecret = getImagePullSecret(i.K8sClient)

	// Get our Mesh CRD to set as an owner for cluster-scoped resources
	i.owner = &appsv1.StatefulSet{}
	// TODO don't hardcode operator namespace
	err := (*i.K8sClient).Get(ctx, client.ObjectKey{Name: "gm-operator", Namespace: "gm-operator"}, i.owner)
	if err != nil {
		logger.Error(err, "Failed to get the operator's own StatefulSet: 'gm-operator'")
		return err
	}

	if i.Config.Spire {
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
		spireSecret, err = injectGeneratedCertificates(spireSecret, i.cfssl)
		if err != nil {
			logger.Error(err, "Error while attempting to apply spire server-ca secret", "secret object", spireSecret)
			return err
		}
		k8sapi.Apply(i.K8sClient, spireSecret, i.owner, k8sapi.CreateOrUpdate)
	}

	// Try to get the OpenShift cluster ingress domain if it exists.
	clusterIngressDomain, ok := getOpenshiftClusterIngressDomain(i.K8sClient, i.Config.ClusterIngressName)
	if ok {
		// TODO: When not in OpenShift, check for other supported ingress class types such as Nginx or Voyager.
		// If no supported ingress types are found, just assume the user will configure ingress on their own.
		logger.Info("Identified OpenShift cluster domain name", "Domain", clusterIngressDomain)
		i.clusterIngressDomain = clusterIngressDomain
	}

	// called on completion of a sync cycle if there are new commits
	i.Sync.OnSyncCompleted = func() error {
		logger.Info("GitOps repo updated and synchronized. Reapplying configuration...")
		i.ApplyMesh()
		return nil
	}

	logger.Info("Apply loaded mesh to cluster.")
	i.ApplyMesh()
	i.Sync.Watch() // Executes its callback (defined above) whenever there are new commits

	// If Spire, set up to periodically reconcile the extant sidecars with the Redis listener's allowable subjects
	if i.Config.Spire {
		go i.reconcileSidecarListForRedisIngress(i.Mesh)
	}

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

func injectGeneratedCertificates(secret *corev1.Secret, cs *cfsslsrv.CFSSLServer) (*corev1.Secret, error) {
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
func (i *Installer) reconcileSidecarListForRedisIngress(mesh *v1alpha1.Mesh) {
	var redisListener json.RawMessage
	var tempOperatorCUE cuemodule.OperatorCUE
	var err error
ReconciliationLoop:
	for {
		time.Sleep(30 * time.Second)
		sidecarSet := make(map[string]struct{})
		// TODO it may be better to do Deployments and StatefulSets (but as a first pass, Pods are far simpler)
		i.RLock()
		// List all pods anywhere
		pods := &corev1.PodList{}
		(*i.K8sClient).List(context.TODO(), pods)
		for _, pod := range pods.Items {
			// Filter to only the relevant namespaces for this mesh
			watched := false
			for _, ns := range mesh.Spec.WatchNamespaces {
				if pod.Namespace == ns {
					watched = true
					break
				}
			}
			if watched || pod.Namespace == mesh.Spec.InstallNamespace {
				// Further filter to only the pods with a sidecar (assumed to have a container with a "proxy" port)
				for _, container := range pod.Spec.Containers {
					for _, p := range container.Ports {
						// TODO don't hard-code the port name, pull it from the CUE
						// TODO also, seriously? There's got to be a better way to identify sidecars than this
						if p.Name == "proxy" {
							if pod.Labels == nil {
								pod.Labels = make(map[string]string)
							}
							if clusterName, ok := pod.Labels[wellknown.LABEL_CLUSTER]; ok {
								sidecarSet[clusterName] = struct{}{}
							}
						}
					}
				}
			}
		}
		var sidecarList []string
		for name := range sidecarSet {
			sidecarList = append(sidecarList, name)
		}
		sort.Strings(sidecarList)
		sort.Strings(i.Defaults.SidecarList)
		if len(sidecarList) == 0 || reflect.DeepEqual(sidecarList, i.Defaults.SidecarList) {
			goto LoopEnd
		}
		logger.Info("The list of sidecars in the environment has changed. Updating Redis ingress for health checks.", "Updated List", sidecarList)
		i.Defaults.SidecarList = sidecarList
		tempOperatorCUE, err = i.OperatorCUE.TempGMValueUnifiedWithDefaults(i.Defaults)
		if err != nil {
			logger.Error(err,
				"error attempting to unify mesh after sidecarList update - this should never happen - check Mesh integrity",
				"Mesh", i.Mesh)
			goto LoopEnd
		}
		redisListener, err = tempOperatorCUE.ExtractRedisListener()
		if err != nil {
			logger.Error(err,
				"error extracting redis_listener from CUE - ignoring",
				"Mesh", i.Mesh)
			goto LoopEnd
		}
		if i.Client != nil {
			i.Client.ControlCmds <- gmapi.MkApply("listener", redisListener)
		}

	LoopEnd:
		if i.Client != nil {
			select {
			case <-i.Client.Ctx.Done():
				logger.Info("greymatter client context cancelled - stopping reconciliation loop")
				break ReconciliationLoop
			default:
			}
		}
		i.RUnlock()
	}
}
