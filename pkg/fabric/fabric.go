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
	Proxy            json.RawMessage    `json:"proxy,omitempty"`
	Domain           json.RawMessage    `json:"domain,omitempty"`
	Listener         json.RawMessage    `json:"listener,omitempty"`
	Clusters         []json.RawMessage  `json:"clusters,omitempty"`
	Routes           []json.RawMessage  `json:"routes,omitempty"`
	Ingresses        map[string]Objects `json:"ingresses,omitempty"`
	LocalEgresses    *Objects           `json:"localEgresses,omitempty"`
	ExternalEgresses *Objects           `json:"externalEgresses,omitempty"`
	CatalogService   json.RawMessage    `json:"catalogservice,omitempty"`
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

type Egress struct {
	IsTCP        bool   `json:"isTCP"`
	Cluster      string `json:"cluster"` // name of a cluster; if external, cluster is generated
	ExternalHost string `json:"externalHost"`
	ExternalPort int32  `json:"externalPort"`
}

// Extracts service configs from a Fabric's cue.Value.
func (f *Fabric) Service(name string, annotations map[string]string, ingresses map[string]int32, egresses ...Egress) (Objects, error) {
	if ingresses == nil {
		ingresses = make(map[string]int32)
	}

	// All ingress routes, whether HTTP or TCP, use 10808.
	// HTTP egress routes that are local (in-mesh) go to 10818; external go to 10919.
	// TCP egress routes that are local or external go to 10920 and up, and are assigned.

	localEgresses := []Egress{} // TODO: tack on http vs tcp
	externalEgresses := []Egress{}
	for _, e := range egresses {
		if e.ExternalHost != "" && e.ExternalPort != 0 {
			externalEgresses = append(externalEgresses, e)
		} else {
			localEgresses = append(localEgresses, e)
		}
	}

	if annotations == nil {
		annotations = make(map[string]string)
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

	value := f.cue.Unify(
		cueutils.FromStrings(fmt.Sprintf(`
			ServiceName: "%s"
			HttpFilters: %s
			NetworkFilters: %s
			Ingresses: %s
			LocalEgresses: %s
			ExternalEgresses: %s
		`,
			name,
			mustMarshal(httpFilters),
			mustMarshal(networkFilters),
			mustMarshal(ingresses),
			mustMarshal(localEgresses),
			mustMarshal(externalEgresses),
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
