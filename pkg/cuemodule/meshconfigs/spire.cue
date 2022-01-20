package meshconfigs

#secret: {
	_name:    string
	_subject: string
	set_current_client_cert_details?: {...}
	forward_client_cert_details?: string

	secret_validation_name: "spiffe://greymatter.io"
	secret_name:            "spiffe://greymatter.io/\(mesh.metadata.name).\(_name)"
	subject_names: ["spiffe://greymatter.io/\(mesh.metadata.name).\(_subject)"]
	ecdh_curves: ["X25519:P-256:P-521:P-384"]
}

service: {
  clusters: [
    {
      require_tls: true
      secret:      #secret & {
        _name:    "edge"
        _subject: ServiceName
      }
    },
  ]

  listener: {
    if ServiceName != "edge" {
      secret: #secret & {
        _name:    ServiceName
        _subject: "edge"
        set_current_client_cert_details: uri: true
        forward_client_cert_details: "APPEND_FORWARD"
      }
    }
  }

  httpEgresses: {
    if len(HTTPEgresses) > 0 {
      clusters: [
        for i, e in HTTPEgresses {
          if e.isExternal {
            {}
          }
          if !e.isExternal {
            {
              require_tls: true
              secret:      #secret & {
                _name:    ServiceName
                _subject: e.cluster
              }
            }
          }
        },
      ]
    }
  }

  tcpEgresses: [
    for i, e in TCPEgresses {
      {
        clusters: [
          if e.isExternal {
            {}
          },
          if !e.isExternal {
            {
              require_tls: true
              secret:      #secret & {
                _name:    ServiceName
                _subject: e.cluster
              }
            }
          },
        ]
      }
    },
  ]

  localEgresses: [
    for _, e in HTTPEgresses if !e.isExternal {
      e.cluster
    },
    for _, e in TCPEgresses if !e.isExternal {
      e.cluster
    },
  ]
}
