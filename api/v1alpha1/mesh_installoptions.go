package v1alpha1

import "github.com/greymatter-io/operator/pkg/version"

func (m Mesh) InstallOptions() []version.InstallOption {
	opts := []version.InstallOption{
		version.Namespace(m.ObjectMeta.Namespace),
		// version.IngressPort(...)
		version.Redis(m.Spec.ExternalRedis),
		// version.WatchNamespaces(m.Spec.WatchNamespaces...)
	}
	opts = append(opts, version.Namespace(m.ObjectMeta.Namespace))

	return opts
}
