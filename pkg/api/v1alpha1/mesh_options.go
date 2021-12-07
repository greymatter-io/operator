package v1alpha1

import (
	"encoding/json"
	"fmt"

	"github.com/greymatter-io/operator/pkg/cuedata"
	"github.com/greymatter-io/operator/pkg/version"

	"cuelang.org/go/cue"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("v1alpha1")
)

// Options returns a slice of cue.Value derived from configured Mesh values.
func (m Mesh) Options(clusterIngressDomain string) []cue.Value {

	// ClusterIngressDomain is defined in OpenShift clusters, but empty otherwise.
	// If it is defined, we specify the mesh name as a subdomain for OpenShift's router.
	environment := "kubernetes"
	var ingressSubDomain string
	if clusterIngressDomain != "" {
		environment = "openshift"
		ingressSubDomain = fmt.Sprintf("%s.%s", m.Name, clusterIngressDomain)
	}

	opts := []cue.Value{
		cuedata.Strings(map[string]string{
			"Environment":      environment,
			"MeshName":         m.Name,
			"ReleaseVersion":   m.Spec.ReleaseVersion,
			"InstallNamespace": m.Spec.InstallNamespace,
			"Zone":             m.Spec.Zone,
			"IngressSubDomain": ingressSubDomain,
		}),
		cuedata.StringSlices(map[string][]string{
			"WatchNamespaces": append(m.Spec.WatchNamespaces, m.Spec.InstallNamespace),
		}),
		version.JWTSecrets(),
	}

	if m.Spec.ExternalRedis != nil {
		opts = append(opts, version.Redis(m.Spec.ExternalRedis.URL))
	} else {
		opts = append(opts, version.Redis(""))
	}

	if len(m.Spec.UserTokens) > 0 {
		users, err := json.Marshal(m.Spec.UserTokens)
		if err != nil {
			logger.Error(err, "Failed to unmarshal UserTokens", "Namesapce", m.Spec.InstallNamespace, "Mesh", m.Name)
		} else {
			opts = append(opts, version.UserTokens(string(users)))
		}
	}

	return opts
}

// ExternalRedisConfig instructs core services to use an external Redis server for caching.
// TODO: Instead of `url`, require host, port, password, dbs. No username option.
type ExternalRedisConfig struct {
	// +kubebuilder:validation:Required
	URL string `json:"url"`
	// +optional
	CertSecretName string `json:"cert_secret_name"`
}

type UserToken struct {
	Label  string              `json:"label"`
	Values map[string][]string `json:"values"`
}
