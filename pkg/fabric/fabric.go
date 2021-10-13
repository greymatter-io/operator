// Package fabric defines functions for generating templates for each service in a mesh.
package fabric

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/greymatter-io/operator/pkg/version"
)

type Fabric struct {
	cue cue.Value
}

type Objects struct {
	Proxy    json.RawMessage `json:"proxy,omitempty"`
	Domain   json.RawMessage `json:"domain,omitempty"`
	Listener json.RawMessage `json:"listener,omitempty"`
	Cluster  json.RawMessage `json:"cluster,omitempty"`
	Route    json.RawMessage `json:"route,omitempty"`
	Locals   []Objects       `json:"locals,omitempty"`
}

func New(zone string, meshPort int32) (*Fabric, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to determine working directory")
	}
	instances := load.Instances([]string{"greymatter.io/operator/fabric/cue.mod:fabric"}, &load.Config{
		Package:    "fabric",
		ModuleRoot: wd,
		Dir:        fmt.Sprintf("%s/cue.mod", wd),
	})
	value := cuecontext.New().BuildInstance(instances[0])
	if err := value.Err(); err != nil {
		return nil, err
	}
	return &Fabric{cue: value.Unify(
		version.Cue(fmt.Sprintf(`
			Zone: "%s"
			MeshPort: %d
		`, zone, meshPort)),
	)}, nil
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
func (f *Fabric) Service(name string, ports []int32) Objects {
	value := f.cue.Unify(
		version.Cue(fmt.Sprintf(`
			ServiceName: "%s",
			ServicePorts: %s
		`, name, strings.Join(strings.Fields(fmt.Sprint(ports)), ", "))),
	)

	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var s struct {
		Service Objects `json:"service"`
	}
	codec.Encode(value, &s)
	return s.Service
}
