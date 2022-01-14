package fabric

// This file is for manually testing various inputs with the cue CLI.
// i.e. "cue eval -e service.routes"

MeshName: "m"
ReleaseVersion: "1.7"
Zone: "z"
ServiceName: "mock"
Ingresses: "": 3000
IngressTCPPortName: ""
HTTPEgresses: []
TCPEgresses: [{
  isExternal: false
  cluster: "gm-redis"
  host: ""
  port: 0
  tcpPort: 10910
}]
