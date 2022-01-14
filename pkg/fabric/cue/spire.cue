package fabric

MeshName: string
ServiceName: string
IngressTCPPortName: string
HTTPEgresses: [...#args]
TCPEgresses: [...#args]

#args: {
  isExternal: bool
  cluster: string
  ...
}

#secret: {
  _name: string
  _subject: string
  set_current_client_cert_details?: {...}
  forward_client_cert_details?: string

  secret_validation_name: "spiffe://greymatter.io"
  secret_name: "spiffe://greymatter.io/\(MeshName).\(_name)"
  subject_names: ["spiffe://greymatter.io/\(MeshName).\(_subject)"]
  ecdh_curves: ["X25519:P-256:P-521:P-384"]
}

service: clusters: [
  // If not empty, this is a non-HTTP service, not routable from edge.
  if IngressTCPPortName == "" {
    {
      require_tls: true
      secret: #secret & {
        _name: "edge"
        _subject: ServiceName
      }
    }
  }
]

service: listener: {
  if ServiceName != "edge" {
    secret: #secret & {
      _name: ServiceName
      _subject: "edge"
      set_current_client_cert_details: uri: true
      forward_client_cert_details: "APPEND_FORWARD"
    }
  }
}

service: httpEgresses: clusters: [
  for i, e in HTTPEgresses {
    if e.isExternal {
      {}
    }
    if !e.isExternal {
      {
        require_tls: true
        secret: #secret & {
          _name: ServiceName
          _subject: e.cluster
        }
      }
    }
  }
]

service: tcpEgresses: [
  for i, e in TCPEgresses {
    clusters: [
      if e.isExternal {
        {}
      }
      if !e.isExternal {
        {
          require_tls: true
          secret: #secret & {
            _name: ServiceName
            _subject: e.cluster
          }
        }
      }
    ]
  }
]

service: localEgresses: [
  for _, e in HTTPEgresses if !e.isExternal {
    e.cluster
  }
  for _, e in TCPEgresses if !e.isExternal {
    e.cluster
  }
]
