package meshconfigs

import "github.com/greymatter-io/operator/api/v1alpha1"

mesh: v1alpha1.#Mesh

// MeshName: string
// ReleaseVersion: string
// Zone:           string

ServiceName:    string
Ingresses: [string]: int32
HTTPEgresses: [...#EgressArgs]
TCPEgresses: [...#EgressArgs]

#EgressArgs: {
	isExternal: bool
	cluster:    string
	host:       string
	port:       int
	tcpPort:    int
}

HttpFilters: {
  "gm.metrics": true
}

NetworkFilters: {
  "envoy.tcp_proxy": *false | bool
}
