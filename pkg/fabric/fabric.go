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
	Proxy    json.RawMessage   `json:"proxy"`
	Domain   json.RawMessage   `json:"domain"`
	Listener json.RawMessage   `json:"listener"`
	Clusters []json.RawMessage `json:"clusters"`
	Routes   []json.RawMessage `json:"routes"`
	// Ingresses are in the same pod as a sidecar, reached via 10808.
	// The key takes the form of '{sidecar-cluster}-{containerPort.name}'.
	Ingresses *Objects `json:"ingresses"`
	// HTTP egresses are reached via the same listener on port 10909.
	// They can be local (in the same mesh) or external.
	HTTPEgresses *Objects `json:"httpEgresses"`
	// TCP egresses are served at 10910 and up (one listener each).
	TCPEgresses    []Objects       `json:"tcpEgresses"`
	CatalogService json.RawMessage `json:"catalogservice"`
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
	IsExternal bool   `json:"isExternal"`
	Cluster    string `json:"cluster"` // name of a cluster; if external, cluster is created
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	TCPPort    int32  `json:"tcpPort"`
}

// Extracts service configs from a Fabric's cue.Value.
func (f *Fabric) Service(name string, annotations map[string]string, ingresses map[string]int32) (Objects, error) {

	// All ingress routes use 10808. They are defined from named container ports.
	// There may only be one ingress if TCP (via the envoy.tcp_proxy filter).
	if ingresses == nil {
		ingresses = make(map[string]int32)
	}

	// Annotations are used for setting filters and creating egress routes.
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Filter names are passed as comma-delimited strings, but transformed into lookup tables for Cue.
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

	// All HTTP egress routes use 10909. They can be local (in-mesh) or external.
	httpEgresses := []EgressArgs{}
	if locals, ok := annotations["greymatter.io/http-local-egress"]; ok {
		for _, cluster := range strings.Split(locals, ",") {
			httpEgresses = append(httpEgresses, EgressArgs{Cluster: strings.TrimSpace(cluster)})
		}
	}
	if externals, ok := annotations["greymatter.io/http-external-egress"]; ok {
		for _, e := range strings.Split(externals, ",") {
			split := strings.Split(strings.TrimSpace(e), ";")
			if len(split) != 2 {
				logger.Error(fmt.Errorf("unable to parse"), "HTTP external egress", "value", e)
			} else {
				addr := strings.Split(split[1], ":")
				if len(addr) != 2 {
					logger.Error(fmt.Errorf("unable to parse"), "HTTP external egress", "value", e)
				} else {
					port, err := strconv.ParseInt(addr[1], 0, 32)
					if err != nil {
						logger.Error(fmt.Errorf("unable to parse"), "HTTP external egress", "value", e)
					} else {
						httpEgresses = append(httpEgresses, EgressArgs{
							IsExternal: true,
							Cluster:    split[0],
							Host:       addr[0],
							Port:       int32(port),
						})
					}
				}
			}
		}
	}

	// TCP egresses are served at 10910 and up (one listener each).
	tcpEgresses := []EgressArgs{}
	tcpPort := int32(10910)
	if locals, ok := annotations["greymatter.io/tcp-local-egress"]; ok {
		for _, cluster := range strings.Split(locals, ",") {
			tcpEgresses = append(tcpEgresses, EgressArgs{
				Cluster: strings.TrimSpace(cluster),
				TCPPort: tcpPort,
			})
			tcpPort++
		}
	}
	if externals, ok := annotations["greymatter.io/tcp-external-egress"]; ok {
		for _, e := range strings.Split(externals, ",") {
			split := strings.Split(strings.TrimSpace(e), ";")
			if len(split) != 2 {
				logger.Error(fmt.Errorf("unable to parse"), "TCP external egress", "value", e)
			} else {
				addr := strings.Split(split[1], ":")
				if len(addr) != 2 {
					logger.Error(fmt.Errorf("unable to parse"), "TCP external egress", "value", e)
				} else {
					port, err := strconv.ParseInt(addr[1], 0, 32)
					if err != nil {
						logger.Error(fmt.Errorf("unable to parse"), "TCP external egress", "value", e)
					} else {
						tcpEgresses = append(tcpEgresses, EgressArgs{
							IsExternal: true,
							Cluster:    split[0],
							Host:       addr[0],
							Port:       int32(port),
							TCPPort:    tcpPort,
						})
						tcpPort++
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
			HTTPEgresses: %s
			TCPEgresses: %s
		`,
			name,
			mustMarshal(httpFilters, `{}`),
			mustMarshal(networkFilters, `{}`),
			mustMarshal(ingresses, `[]`),
			mustMarshal(httpEgresses, `[]`),
			mustMarshal(tcpEgresses, `[]`),
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

func mustMarshal(i interface{}, fallback string) string {
	result, err := json.Marshal(i)
	if err != nil {
		logger.Error(err, "failed to marshal, using default", "data", i)
		result = json.RawMessage(fallback)
	}
	return string(result)
}
