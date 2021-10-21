// Package fabric defines functions for generating templates for each service in a mesh.
package fabric

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	"cuelang.org/go/pkg/strconv"
)

var (
	logger = ctrl.Log.WithName("fabric")
)

type Fabric struct {
	cue cue.Value
}

func New(mesh *v1alpha1.Mesh) *Fabric {
	return &Fabric{cue: value.Unify(
		cueutils.FromStrings(fmt.Sprintf(`
			MeshName: "%s"
			Zone: "%s"
			MeshPort: %d
		`, mesh.Name, mesh.Spec.Zone, mesh.Spec.MeshPort)),
	)}
}

type Objects struct {
	Proxy    json.RawMessage   `json:"proxy,omitempty"`
	Domain   json.RawMessage   `json:"domain,omitempty"`
	Listener json.RawMessage   `json:"listener,omitempty"`
	Clusters []json.RawMessage `json:"clusters,omitempty"`
	Routes   []json.RawMessage `json:"routes,omitempty"`
	// Ingresses are in the same pod as a sidecar, reached via 10808.
	Ingresses map[string]Objects `json:"ingresses,omitempty"`
	// HTTP local egresses are in the same mesh, reached via 10818.
	HTTPLocalEgresses *Objects `json:"httpLocalEgresses,omitempty"`
	// HTTP external egresses are outside of the mesh, reached via 10919.
	HTTPExternalEgresses *Objects `json:"httpExternalEgresses,omitempty"`
	// TCPEgresses
	CatalogService json.RawMessage `json:"catalogservice,omitempty"`
}

type Egress struct {
}

// Extracts the edge domain from a Fabric's cue.Value.
// The edge domain is needed separately since it is referenced by sidecar routes.
func (f *Fabric) EdgeDomain() json.RawMessage {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var e struct {
		EdgeDomain json.RawMessage `json:"edgeDomain"`
	}
	codec.Encode(f.cue, &e)
	return e.EdgeDomain
}

type EgressArgs struct {
	IsTCP        bool   `json:"isTCP"`
	Cluster      string `json:"cluster"` // name of a cluster; if external, cluster is generated
	ExternalHost string `json:"externalHost"`
	ExternalPort int32  `json:"externalPort"`
}

// Extracts service configs from a Fabric's cue.Value.
func (f *Fabric) Service(name string, annotations map[string]string, ingresses map[string]int32) (Objects, error) {
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if ingresses == nil {
		ingresses = make(map[string]int32)
	}

	httpFilters := make(map[string]bool)
	if hf, ok := annotations["greymatter.io/http-filters"]; ok {
		for _, f := range strings.Split(hf, ",") {
			httpFilters[f] = true
		}
	}
	networkFilters := make(map[string]bool)
	if nf, ok := annotations["greymatter.io/network-filters"]; ok {
		for _, f := range strings.Split(nf, ",") {
			networkFilters[f] = true
		}
	}

	// All ingress routes, whether HTTP or TCP, use 10808.
	// HTTP egress routes that are local (in-mesh) go to 10818; external go to 10909.
	// TCP egress routes that are local or external go to 10920 and up, and are assigned.

	// TODO: If TCP, the user should use tcp:cluster
	httpLocalEgresses := []EgressArgs{}
	if le, ok := annotations["greymatter.io/local-egress"]; ok {
		for _, cluster := range strings.Split(le, ",") {
			httpLocalEgresses = append(httpLocalEgresses, EgressArgs{Cluster: strings.TrimSpace(cluster)})
		}
	}

	// TODO: Consider parsing JSON here, as this is getting too complex
	httpExternalEgresses := []EgressArgs{}
	if ex, ok := annotations["greymatter.io/external-egress"]; ok {
		for _, e := range strings.Split(ex, ",") {
			split := strings.Split(e, ";")
			if len(split) == 2 {
				addr := strings.Split(split[1], ":")
				if len(addr) == 2 {
					port, err := strconv.ParseInt(addr[1], 0, 32)
					if err != nil {
						logger.Error(err, "invalid port specified in external egress", "value", e)
					} else {
						httpExternalEgresses = append(httpExternalEgresses, EgressArgs{
							Cluster:      split[0],
							ExternalHost: addr[0],
							ExternalPort: int32(port),
						})
					}
				}
			}
		}
	}

	value := f.cue.Unify(
		cueutils.FromStrings(fmt.Sprintf(`
			ServiceName: "%s"
			HttpFilters: %s
			NetworkFilters: %s
			Ingresses: %s
			HTTPLocalEgresses: %s
			HTTPExternalEgresses: %s
		`,
			name,
			mustMarshal(httpFilters),
			mustMarshal(networkFilters),
			mustMarshal(ingresses),
			mustMarshal(httpLocalEgresses),
			mustMarshal(httpExternalEgresses),
		)),
	)
	if err := value.Err(); err != nil {
		cueutils.LogError(logger, err)
		return Objects{}, err
	}

	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var s struct {
		Service Objects `json:"service"`
	}
	codec.Encode(value, &s)
	return s.Service, nil
}

func mustMarshal(i interface{}) string {
	result, err := json.Marshal(i)
	if err != nil {
		logger.Error(err, "failed to marshal", "data", i)
	}
	return string(result)
}
