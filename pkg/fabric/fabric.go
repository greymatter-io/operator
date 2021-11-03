// Package fabric defines functions for generating templates for each service in a mesh.
package fabric

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/greymatter-io/operator/pkg/cueutils"
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

// New returns a new *Fabric instance. It receives mesh options
// which are unified with base fabric object templates defined in cue/fabric.cue.
func New(options []cue.Value) *Fabric {
	v := *value
	for _, o := range options {
		v = v.Unify(o)
	}
	return &Fabric{cue: v}
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
	TCPEgresses    []Objects       `json:"tcpEgresses"`
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

// EgressArgs contains fabric object values for egress cluster routes.
// It is a struct for passing values to be parsed in cue/fabric.cue.
type EgressArgs struct {
	IsExternal bool `json:"isExternal"`
	// A reference to either another cluster in a mesh,
	// or a cluster that should be created with the provided host:port.
	Cluster string `json:"cluster"`
	Host    string `json:"host"`
	Port    int32  `json:"port"`
	TCPPort int32  `json:"tcpPort"`
}

// Service creates fabric objects for adding a workload to a mesh,
// given the workload's annotations and a map of its parsed container ports.
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
	httpFilters := parseFilters(annotations["greymatter.io/http-filters"])
	networkFilters := parseFilters(annotations["greymatter.io/network-filters"])

	// All HTTP egress routes use 10909. They can be local (in-mesh) or external.
	// Array literals are used here so that they are parsed as non-nil arrays in Cue.
	httpEgresses, _ := parseLocalEgressArgs([]EgressArgs{}, annotations["greymatter.io/egress-http-local"], 0)
	httpEgresses, _ = parseExternalEgressArgs(httpEgresses, annotations["greymatter.io/egress-http-external"], 0)

	// TCP egresses are served at 10910 and up (one listener each).
	// Redis and (eventually) NATS TCP egresses are prepended by default for all services.
	// TODO: Pass options here for specifying an external Redis egress (if configured).
	tcpEgresses := []EgressArgs{}
	if name != "gm-redis" {
		tcpEgresses = append(tcpEgresses, EgressArgs{Cluster: "gm-redis", TCPPort: 10910})
	}
	// if name != "gm-nats" {
	// 	tcpEgresses = append(tcpEgresses, EgressArgs{Cluster: "gm-nats", TCPPort: 10911})
	// }

	tcpPort := int32(10912)
	tcpEgresses, tcpPort = parseLocalEgressArgs(tcpEgresses, annotations["greymatter.io/egress-tcp-local"], tcpPort)
	tcpEgresses, _ = parseExternalEgressArgs(tcpEgresses, annotations["greymatter.io/egress-tcp-external"], tcpPort)

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

func parseFilters(s string) map[string]bool {
	filters := make(map[string]bool)
	if len(s) == 0 {
		return filters
	}
	var slice []string
	if err := json.Unmarshal([]byte(s), &slice); err != nil {
		logger.Error(err, "failed to unmarshal", "filters", s)
		return filters
	}
	for _, filter := range slice {
		trimmed := strings.TrimSpace(filter)
		if trimmed != "" {
			filters[trimmed] = true
		}
	}
	return filters
}

func parseLocalEgressArgs(args []EgressArgs, s string, tcpPort int32) ([]EgressArgs, int32) {
	if len(s) == 0 {
		if args == nil {
			return []EgressArgs{}, tcpPort
		}
		return args, tcpPort
	}

	var slice []string
	if err := json.Unmarshal([]byte(s), &slice); err != nil {
		logger.Error(err, "failed to unmarshal", "egress-local", s)
		return args, tcpPort
	}

	for _, cluster := range slice {
		trimmed := strings.TrimSpace(cluster)
		if trimmed == "" {
			continue
		}
		args = append(args, EgressArgs{
			Cluster: strings.TrimSpace(cluster),
			TCPPort: tcpPort,
		})
		if tcpPort >= 10912 {
			tcpPort++
		}
	}

	return args, tcpPort
}

type ExtEgressAnnotation struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int32  `json:"port"`
}

func parseExternalEgressArgs(args []EgressArgs, s string, tcpPort int32) ([]EgressArgs, int32) {
	if len(s) == 0 {
		if args == nil {
			return []EgressArgs{}, tcpPort
		}
		return args, tcpPort
	}

	var slice []struct {
		Name string `json:"name"`
		Host string `json:"host"`
		Port int32  `json:"port"`
	}
	if err := json.Unmarshal([]byte(s), &slice); err != nil {
		logger.Error(err, "failed to unmarshal", "egress-local", s)
		return args, tcpPort
	}

	for _, e := range slice {
		if e.Name == "" || e.Host == "" || e.Port == 0 {
			logger.Error(fmt.Errorf("unable to parse"), "required: name, host, port", "value", e)
		}

		args = append(args, EgressArgs{
			IsExternal: true,
			Cluster:    strings.TrimSpace(e.Name),
			Host:       strings.TrimSpace(e.Host),
			Port:       e.Port,
			TCPPort:    tcpPort,
		})
		if tcpPort >= 10912 {
			tcpPort++
		}
	}

	return args, tcpPort
}

func mustMarshal(i interface{}, fallback string) string {
	result, err := json.Marshal(i)
	if err != nil {
		logger.Error(err, "failed to marshal, using default", "data", i)
		result = json.RawMessage(fallback)
	}
	return string(result)
}
