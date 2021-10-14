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
	logger = ctrl.Log.WithName("pkg.fabric")
)

type Fabric struct {
	cue cue.Value
}

func New(mesh *v1alpha1.Mesh) (*Fabric, error) {
	return &Fabric{cue: value.Unify(
		cueutils.FromStrings(fmt.Sprintf(`
			Zone: "%s"
			MeshPort: %d
		`, mesh.Spec.Zone, mesh.Spec.MeshPort)),
	)}, nil
}

type Objects struct {
	Proxy     json.RawMessage    `json:"proxy,omitempty"`
	Domain    json.RawMessage    `json:"domain,omitempty"`
	Listener  json.RawMessage    `json:"listener,omitempty"`
	Cluster   json.RawMessage    `json:"cluster,omitempty"`
	Route     json.RawMessage    `json:"route,omitempty"`
	Ingresses map[string]Objects `json:"ingresses,omitempty"`
}

// Extracts edge configs from a Fabric's cue.Value.
func (f *Fabric) Edge() Objects {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var e struct {
		Edge Objects `json:"edge"`
	}
	codec.Encode(f.cue, &e)
	return e.Edge
}

// Extracts service configs from a Fabric's cue.Value.
func (f *Fabric) Service(name string, ingresses map[string]int32) (Objects, error) {
	if len(ingresses) == 0 {
		return Objects{}, fmt.Errorf("no ingresses specified")
	}

	j, err := json.Marshal(ingresses)
	if err != nil {
		return Objects{}, fmt.Errorf("failed to marshal ingresses")
	}

	value := f.cue.Unify(
		cueutils.FromStrings(fmt.Sprintf(`
			ServiceName: "%s",
			ServiceIngresses: %s
		`, name, string(j))),
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
