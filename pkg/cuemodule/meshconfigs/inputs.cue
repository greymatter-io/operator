package meshconfigs

import (
	"encoding/json"
	"github.com/greymatter-io/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A Mesh CR being applied.
mesh: v1alpha1.#Mesh

// An abstraction around a Deployment/StatefulSet being applied.
workload: {
	metadata: metav1.#ObjectMeta
	// Specify expected optional annotations (which default to an empty string).
	metadata: annotations: {
		"greymatter.io/http-filters":         *"" | string
		"greymatter.io/network-filters":      *"" | string
		"greymatter.io/egress-http-local":    *"" | string
		"greymatter.io/egress-http-external": *"" | string
		"greymatter.io/egress-tcp-local":     *"" | string
		"greymatter.io/egress-tcp-external":  *"" | string
	}
	spec: template: spec: corev1.#PodSpec
	// spec.containers[n].ports[n].name is optional, so we default to an empty string.
	// This lets us evaluate the concrete value of port.name throughout our code.
	spec: template: spec: {
		containers: [...{
			ports: [...{
				name: *"" | string
			}]
		}]
	}
}

// A map derived from all container ports specified in the PodSpec.
Ingresses: {
	for container in workload.spec.template.spec.containers {
		for port in container.ports {
			if port.name != "" {
				"\(port.name)": port.containerPort
			}
			if port.name == "" {
				"\(port.containerPort)": port.containerPort
			}
		}
	}
}

egressHttpLocal: [...string]
if workload.metadata.annotations["greymatter.io/egress-http-local"] != "" {
	egressHttpLocal: json.Unmarshal(workload.metadata.annotations["greymatter.io/egress-http-local"])
}

egressHttpExternal: [...#EgressArgs]
if workload.metadata.annotations["greymatter.io/egress-http-external"] != "" {
	egressHttpExternal: json.Unmarshal(workload.metadata.annotations["greymatter.io/egress-http-external"])
}

HTTPEgresses: {
	for x in egressHttpLocal {
		"\(x)": #EgressArgs & {}
	}
	for x in egressHttpExternal {
		"\(x.name)": #EgressArgs & {
			name:       x.name
			isExternal: true
			host:       x.host
			port:       x.port
		}
	}
}

egressTcpLocal: [...string]
if workload.metadata.annotations["greymatter.io/egress-tcp-local"] != "" {
	egressTcpLocal: json.Unmarshal(workload.metadata.annotations["greymatter.io/egress-tcp-local"])
}

egressTcpExternal: [...#EgressArgs]
if workload.metadata.annotations["greymatter.io/egress-tcp-external"] != "" {
	egressTcpExternal: json.Unmarshal(workload.metadata.annotations["greymatter.io/egress-tcp-external"])
}

TCPEgresses: {
	if workload.metadata.name != "gm-redis" {
		"gm-redis": #EgressArgs & {tcpPort: 10910}
	}
	for i, x in egressTcpLocal {
		"\(x)": #EgressArgs & {tcpPort: 10912 + i}
	}
	for i, x in egressTcpExternal {
		"\(x.name)": #EgressArgs & {
			isExternal: true
			host:       x.host
			port:       x.port
			tcpPort:    10912 + len(egressTcpLocal) + i
		}
	}
}

#EgressArgs: {
	name:       *"" | string
	isExternal: *false | bool
	host:       *"" | string
	port:       *0 | int
	tcpPort:    *0 | int
}

HttpFilters: {
	"gm.metrics": true
	if workload.metadata.annotations["greymatter.io/http-filters"] != "" {
		for f in json.Unmarshal(workload.metadata.annotations["greymatter.io/http-filters"]) {
			"\(f)": true
		}
	}
}

NetworkFilters: {
	"envoy.tcp_proxy": *false | bool
	if workload.metadata.annotations["greymatter.io/network-filters"] != "" {
		for f in json.Unmarshal(workload.metadata.annotations["greymatter.io/network-filters"]) {
			"\(f)": true
		}
	}
}
