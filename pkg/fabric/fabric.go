// Package fabric defines functions for generating templates for each service in a mesh.
package fabric

import (
	"encoding/json"
	"fmt"

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
	Cluster          json.RawMessage    `json:"cluster,omitempty"`
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
	// Protocol string // http or tcp; TODO
	Cluster  string // the name of a cluster; if external, this is a FQDN
	External bool
}

// Extracts service configs from a Fabric's cue.Value.
func (f *Fabric) Service(name string, ingresses map[string]int32, egresses ...Egress) (Objects, error) {
	if ingresses == nil {
		ingresses = make(map[string]int32)
	}

	i, err := json.Marshal(ingresses)
	if err != nil {
		return Objects{}, fmt.Errorf("failed to marshal ingresses")
	}

	localEgresses := []string{}
	externalEgresses := []string{}
	for _, e := range egresses {
		if e.External {
			externalEgresses = append(externalEgresses, e.Cluster)
		} else {
			localEgresses = append(localEgresses, e.Cluster)
		}
	}

	locals, err := json.Marshal(localEgresses)
	if err != nil {
		return Objects{}, fmt.Errorf("failed to marshal local egresses")
	}

	externals, err := json.Marshal(externalEgresses)
	if err != nil {
		return Objects{}, fmt.Errorf("failed to marshal external egresses")
	}

	value := f.cue.Unify(
		cueutils.FromStrings(fmt.Sprintf(`
			ServiceName: "%s"
			ServiceIngresses: %s
			ServiceLocalEgresses: %s
			ServiceExternalEgresses: %s
		`, name, string(i), string(locals), string(externals))),
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
