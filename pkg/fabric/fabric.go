// Package fabric defines functions for generating templates for each service in a mesh.
package fabric

import (
	"encoding/json"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cueutils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
)

var (
	logger = ctrl.Log.WithName("fabric")
)

// Fabric contains a cue.Value that holds fabric object templates for a single mesh,
// defined by options passed into its factory function.
type Fabric struct {
	cue cue.Value
}

// New returns a new *Fabric instance.
// It receives the meshconfigs template cue.Value
// and a Mesh custom resource to unify it with.
func New(tmpl cue.Value, mesh *v1alpha1.Mesh) (*Fabric, error) {
	m, err := cueutils.FromStruct("mesh", mesh)
	if err != nil {
		return nil, err
	}

	return &Fabric{cue: tmpl.Unify(m)}, nil
}

// Objects contains all fabric objects to apply for adding a workload to a mesh.
// It is a recursive type, enabling references from nested fabric objects to parent Objects.
// Objects should not be defined using Go structs, but parsed from Cue using *Fabric.Service.
type Objects struct {
	Proxy    json.RawMessage   `json:"proxy"`
	Domain   json.RawMessage   `json:"domain"`
	Listener json.RawMessage   `json:"listener"`
	Clusters []json.RawMessage `json:"clusters"`
	Routes   []json.RawMessage `json:"routes"`
	// Ingresses are in the same pod as a sidecar, reached via 10808.
	// The key takes the form of '{sidecar-cluster-name}-{port}'.
	Ingresses *Objects `json:"ingresses"`
	// HTTP egresses are reached via the same listener on port 10909.
	// They can be local (in the same mesh) or external.
	HTTPEgresses *Objects `json:"httpEgresses"`
	// TCP egresses are served at 10910 and up (one listener each).
	// Note that 10910 and 10911 are reserved for internal use by Redis and NATS,
	// so any configured TCP egresses via annotations will start at 10912.
	TCPEgresses []Objects `json:"tcpEgresses"`
	// A list of keys for all local egress clusters routed to from this service.
	// Each egress's listener must be modified to reference this service's SVID in its subjects.
	LocalEgresses  []string        `json:"localEgresses"`
	CatalogService json.RawMessage `json:"catalogservice"`
}

// EdgeDomain extracts the edge domain fabric object from a *Fabric's cue.Value.
// The edge domain is parsed individually since it is created as the root mesh domain.
func (f *Fabric) EdgeDomain() json.RawMessage {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var e struct {
		EdgeDomain json.RawMessage `json:"edgeDomain"`
	}
	codec.Encode(f.cue, &e)
	return e.EdgeDomain
}

// Service creates meshconfigs for adding a workload to a mesh.
func (f *Fabric) Service(name string, workload runtime.Object) (Objects, error) {
	w, err := cueutils.FromStruct("workload", workload)
	if err != nil {
		return Objects{}, err
	}

	value := f.cue.Unify(w)
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
