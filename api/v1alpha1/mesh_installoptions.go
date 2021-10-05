package v1alpha1

import "github.com/greymatter-io/operator/pkg/version"

func (m Mesh) InstallOptions() []version.InstallOption {
	opts := []version.InstallOption{
		version.ProxyPort(m.Spec.ProxyPort),
		version.Namespace(m.ObjectMeta.Namespace),
		version.Redis(m.Spec.ExternalRedis.URL),
		// version.WatchNamespaces(m.Spec.WatchNamespaces...)
	}

	// opts = append(opts, ...)

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
