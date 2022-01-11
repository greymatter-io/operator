package fabric

// This file is for manually testing various inputs with the cue CLI.
// i.e. "cue eval -e service.routes"

MeshName: "m"
ReleaseVersion: "1.7"
Zone: "z"
ServiceName: "mock"
Ingresses: {}
IngressTCPPortName: ""
HTTPEgresses: [
  {
    isExternal: false
    cluster: "mock2"
    host: ""
    port: 0
    tcpPort: 0
  },
  {
    isExternal: true
    cluster: "lambda"
    host: "something.com"
    port: 80
    tcpPort: 0
  },
]
TCPEgresses: [{
  isExternal: false
  cluster: "gm-redis"
  host: ""
  port: 0
  tcpPort: 10910
}]
