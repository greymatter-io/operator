package installer

import (
	_ "embed"
	"strings"

	"github.com/cloudflare/cfssl/csr"
	"github.com/ghodss/yaml"
	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	"github.com/greymatter-io/operator/pkg/k8sapi"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	//go:embed spire.yaml
	spireYAML string
)

type SpireManifests struct {
	Namespace           *corev1.Namespace            `yaml:"namespace"`
	Secret              *corev1.Secret               `yaml:"secret"`
	Service             *corev1.Service              `yaml:"service"`
	StatefulSet         *appsv1.StatefulSet          `yaml:"statefulset"`
	DaemonSet           *appsv1.DaemonSet            `yaml:"daemonset"`
	Role                *rbacv1.Role                 `yaml:"role"`
	RoleBinding         *rbacv1.RoleBinding          `yaml:"rolebinding"`
	ServiceAccounts     []*corev1.ServiceAccount     `yaml:"serviceaccounts"`
	ClusterRoles        []*rbacv1.ClusterRole        `yaml:"clusterroles"`
	ClusterRoleBindings []*rbacv1.ClusterRoleBinding `yaml:"clusterrolebindings"`
	ConfigMaps          []*corev1.ConfigMap          `yaml:"configmaps"`
}

func applySpire(c client.Client, crd *extv1.CustomResourceDefinition, cs *cfsslsrv.CFSSLServer) error {
	spire, err := getSpireManifests()
	if err != nil {
		return err
	}

	spire.Secret, err = injectPKI(spire.Secret, cs)
	if err != nil {
		return err
	}

	for _, manifest := range []client.Object{
		spire.Namespace,
		spire.Secret,
		spire.Service,
		spire.StatefulSet,
		spire.DaemonSet,
		spire.Role,
		spire.RoleBinding,
	} {
		k8sapi.Apply(c, manifest, crd, k8sapi.GetOrCreate)
	}

	for _, sa := range spire.ServiceAccounts {
		k8sapi.Apply(c, sa, crd, k8sapi.GetOrCreate)
	}
	for _, cr := range spire.ClusterRoles {
		k8sapi.Apply(c, cr, crd, k8sapi.GetOrCreate)
	}
	for _, crb := range spire.ClusterRoleBindings {
		k8sapi.Apply(c, crb, crd, k8sapi.GetOrCreate)
	}
	for _, cm := range spire.ConfigMaps {
		k8sapi.Apply(c, cm, crd, k8sapi.GetOrCreate)
	}

	return nil
}

func injectPKI(secret *corev1.Secret, cs *cfsslsrv.CFSSLServer) (*corev1.Secret, error) {
	root := cs.GetRootCA()
	ca, caKey, err := cs.RequestIntermediateCA(csr.CertificateRequest{
		CN:         "Grey Matter SPIFFE Intermediate CA",
		KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
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

func getSpireManifests() (SpireManifests, error) {
	var m SpireManifests
	if err := yaml.Unmarshal([]byte(spireYAML), &m); err != nil {
		return m, err
	}

	return m, nil
}
