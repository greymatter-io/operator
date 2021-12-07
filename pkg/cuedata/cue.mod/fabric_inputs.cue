package fabric

// Values pre-defined when loading manifests
MeshName: string
ReleaseVersion: string
Zone: string
Spire: bool
Redis: {...}

// Values injected in fabric.Service
ServiceName: string
HttpFilters: #EnabledHttpFilters
NetworkFilters: #EnabledNetworkFilters
Ingresses: [string]: int32
HTTPEgresses: [...#EgressArgs]
TCPEgresses: [...#EgressArgs]

#EnabledHttpFilters: {
  "gm.metrics": true
}

#EnabledNetworkFilters: {
  "envoy.tcp_proxy": *false | bool
}

#EgressArgs: {
  isExternal: bool
  cluster: string
  host: string
  port: int
  tcpPort: int
}
