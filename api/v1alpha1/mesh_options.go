package v1alpha1

import (
	"encoding/json"
	"fmt"

	"github.com/greymatter-io/operator/pkg/cueutils"
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
		cueutils.Strings(map[string]string{
			"Environment":      environment,
			"MeshName":         m.Name,
			"ReleaseVersion":   m.Spec.ReleaseVersion,
			"InstallNamespace": m.Spec.InstallNamespace,
			"Zone":             m.Spec.Zone,
			"IngressSubDomain": ingressSubDomain,
		}),
		cueutils.StringSlices(map[string][]string{
			"WatchNamespaces": append(m.Spec.WatchNamespaces, m.Spec.InstallNamespace),
		}),
		version.JWTSecrets(),
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

type UserToken struct {
	Label  string              `json:"label"`
	Values map[string][]string `json:"values"`
}

// ImageSecret can be defined on a per-image basis. The secret name,
// as well as the secret host namespace are required due to the operator
// being segmented in its own namespace.
type ImageSecret struct {
	Name      string `json:"secret_name"`
	Namespace string `json:"secret_namespace"`
}
