package v1alpha1

import (
	"encoding/json"

	"github.com/greymatter-io/operator/pkg/version"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("v1alpha1")
)

func (m Mesh) InstallOptions() []version.InstallOption {
	opts := []version.InstallOption{
		version.MeshName(m.Name),
		version.MeshPort(m.Spec.MeshPort),
		version.EdgeTls(m.Spec.EdgeTlsIngress),
		version.InstallNamespace(m.ObjectMeta.Namespace),
		// TODO: figure out how to get domain from the clusterUrl without making this a mesh config
		version.IngressSubDomain(m.Spec.ClusterUrl),
		version.WatchNamespaces(append(m.Spec.WatchNamespaces, m.ObjectMeta.Namespace)...),
	}

	if m.Spec.Zone != "" {
		opts = append(opts, version.Zone(m.Spec.Zone))
	}

	if m.Spec.ExternalRedis != nil {
		opts = append(opts, version.Redis(m.Spec.ExternalRedis.URL))
	} else {
		opts = append(opts, version.Redis(""))
	}

	if len(m.Spec.UserTokens) > 0 {
		users, err := json.Marshal(m.Spec.UserTokens)
		if err != nil {
			logger.Error(err, "Failed to unmarshal UserTokens", "Namesapce", m.Namespace, "Mesh", m.Name)
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
