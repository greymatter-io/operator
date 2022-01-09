// Inputs

// Values pre-defined from version.Version
MeshName: string
ReleaseVersion: string
Zone: string
Redis: {...}

// Values injected in fabric.Service
ServiceName: string
Ingresses: [string]: int32
IngressTCPPortName: string
HTTPEgresses: [...#EgressArgs]
TCPEgresses: [...#EgressArgs]

#EgressArgs: {
  isExternal: bool
  cluster: string
  host: string
  port: int
  tcpPort: int
}

// Outputs

edgeDomain: #Domain & {
  domain_key: "edge"
  name: "*"
  port: 10808
}

service: {
  if ServiceName != "gm-redis" {
    catalogservice: #CatalogService & {
      if ServiceName == "edge" {
        name: "Grey Matter Edge"
        description: "Handles north/south traffic flowing through the mesh."
        api_endpoint: "/"
      }
      if ServiceName == "jwt-security" {
        name: "Grey Matter JWT Security"
        description: "A JWT token generation and retrieval service."
        api_endpoint: "/services/jwt-security/"
        api_spec_endpoint: "/services/jwt-security/"
      }
      if ServiceName == "control" {
        name: "Grey Matter Control"
        description: "Manages the configuration of the Grey Matter data plane."
        api_endpoint: "/services/control/api/"
        api_spec_endpoint: "/services/control/api/"
      }
      if ServiceName == "catalog" {
        name: "Grey Matter Catalog"
        description: "Interfaces with the control plane to expose the current state of the mesh."
        api_endpoint: "/services/catalog/"
        api_spec_endpoint: "/services/catalog/"
      }
      if ServiceName == "dashboard" {
        name: "Grey Matter Dashboard"
        description: "A user dashboard that paints a high-level picture of the mesh."
      }
    }
  }

  proxy: #Proxy & {
    name: ServiceName
    domain_keys: [
      ServiceName,
      if len(HTTPEgresses) > 0 {
        "\(ServiceName)-egress-http",
      }
      for _, e in TCPEgresses {
        "\(ServiceName)-egress-tcp-to-\(e.cluster)"
      }
    ]
    listener_keys: [
      ServiceName,
      if len(HTTPEgresses) > 0 {
        "\(ServiceName)-egress-http",
      }
      for _, e in TCPEgresses {
        "\(ServiceName)-egress-tcp-to-\(e.cluster)"
      }
    ]
  }

  domain: #Domain & {
    domain_key: ServiceName
    name: "*"
    port: 10808
  }
  listener: #Listener & {
    name: ServiceName
    port: domain.port
    domain_keys: [ServiceName]

    // Only configure HTTP filters for HTTP services.
    if IngressTCPPortName == "" {
      active_http_filters: [
        "gm.metrics",
        if ServiceName != "edge" {
          "gm.observables",
        }
      ]
      http_filters: {
        gm_metrics: {
          metrics_host: "0.0.0.0"
          metrics_port: 8081
          metrics_dashboard_uri_path: "/metrics"
          metrics_prometheus_uri_path: "prometheus"
          metrics_ring_buffer_size: 4096
          prometheus_system_metrics_interval_seconds: 15
          metrics_key_function: "depth"
          if ServiceName == "edge" {
            metrics_key_depth: "1"
          }
          if ServiceName != "edge" {
            metrics_key_depth: "3"
          }
          if ReleaseVersion != "1.6" {
            metrics_receiver: {
              redis_connection_string: "redis://127.0.0.1:10910"
              push_interval_seconds: 10
            }
          }
        }
        if ServiceName != "edge" {
          gm_observables: {
            topic: ServiceName
          }
        }
      }
    }

    // If a TCP port name is specified in annotations, configure the TCP proxy filter.
    if IngressTCPPortName != "" {
      active_network_filters: [
        "envoy.tcp_proxy"
      ]
      network_filters: {
        envoy_tcp_proxy: {
          _key: "\(ServiceName):\(Ingresses[IngressTCPPortName])"
          cluster: _key
          stat_prefix: _key
        }
      }
    }
  }

  // TCP listeners are not edge routable
  if IngressTCPPortName == "" {
    clusters: [...#Cluster] & [
      {
        name: ServiceName
      }
    ]

    routes: [...#Route] & [
      if ServiceName != "dashboard" && ServiceName != "edge" {
        {
          route_key: ServiceName
          domain_key: "edge"
          route_match: {
            path: "/services/\(ServiceName)/"
            match_type: "prefix"
          }
          redirects: [
            {
              from: "^/services/\(ServiceName)$"
              to: route_match.path
              redirect_type: "permanent"
            }
          ]
          prefix_rewrite: "/"
          rules: [
            {
              constraints: {
                light: [
                  {
                    cluster_key: ServiceName
                    weight: 1
                  }
                ]
              }
            }
          ]
        }
      }
      if ServiceName == "dashboard" {
        {
          route_key: ServiceName
          domain_key: "edge"
          route_match: {
            path: "/"
            match_type: "prefix"
          }
          rules: [
            {
              constraints: {
                light: [
                  {
                    cluster_key: ServiceName
                    weight: 1
                  }
                ]
              }
            }
          ]
        }
      }
    ]
  }

  ingresses: {
    clusters: [...#Cluster] & [
      for k, v in Ingresses if len(Ingresses) > 0 && ServiceName != "edge" {
        // If this is an HTTP service, configure all container ports as ingresses.
        // If this is a TCP service, configure the named container port as a single ingress.
        if IngressTCPPortName == "" || IngressTCPPortName == k {
          {
            name: "\(ServiceName):\(v)"
            instances: [
              {
                host: "127.0.0.1"
                port: v
              }
            ]
          }
        }
      }
    ]
    routes: [...#Route] & [
      for k, v in Ingresses if len(Ingresses) > 0 && ServiceName != "edge" {
        // If this is an HTTP service with a single container port
        // or a TCP service, configure a single route to it at the root of the listener.
        if len(Ingresses) == 1 || (IngressTCPPortName != "" && IngressTCPPortName == k) {
          {
            _rk: "\(ServiceName):\(v)"
            route_key: _rk
            domain_key: ServiceName
            route_match: {
              path: "/"
              match_type: "prefix"
            }
            rules: [
              {
                constraints: {
                  light: [
                    {
                      cluster_key: _rk
                      weight: 1
                    }
                  ]
                }
              }
            ]
          }
        }
        // If this is an HTTP service with multiple container ports, make edge-routable.
        if len(Ingresses) > 1 && IngressTCPPortName == "" {
          {
            _rk: "\(ServiceName):\(v)"
            route_key: _rk
            domain_key: ServiceName
            route_match: {
              path: "/\(k)/"
              match_type: "prefix"
            }
            redirects: [
              {
                from: "^/\(k)$"
                to: route_match.path
                redirect_type: "permanent"
              }
            ]
            prefix_rewrite: "/"
            rules: [
              {
                constraints: {
                  light: [
                    {
                      cluster_key: _rk
                      weight: 1
                    }
                  ]
                }
              }
            ]
          }
        }
      }
    ]
  }

  httpEgresses: {
    _dk: "\(ServiceName)-egress-http"
    if len(HTTPEgresses) > 0 {
      domain: #Domain & {
        domain_key: _dk
        name: "*"
        port: 10909
      }
      listener: #Listener & {
        name: _dk
        port: 10909
        domain_keys: [_dk]
      }
      clusters: [...#Cluster] & [
        for _, e in HTTPEgresses {
          if e.isExternal {
            {
              name: "\(ServiceName)-to-\(e.cluster)"
              instances: [
                {
                  host: e.host
                  port: e.port
                }
              ]
            }
          }
          if !e.isExternal {
            {
              name: e.cluster
              cluster_key: "\(ServiceName)-to-\(e.cluster)"
            }
          }
        }
      ]
      routes: [...#Route] & [
        for _, e in HTTPEgresses {
          {
            _rk: "\(ServiceName)-to-\(e.cluster)"
            route_key: _rk
            domain_key: _dk
            route_match: {
              path: "/\(e.cluster)/"
              match_type: "prefix"
            }
            redirects: [
              {
                from: "^/\(e.cluster)$"
                to: route_match.path
                redirect_type: "permanent"
              }
            ]
            prefix_rewrite: "/"
            rules: [
              {
                constraints: {
                  light: [
                    {
                      cluster_key: _rk
                      weight: 1
                    }
                  ]
                }
              }
            ]
          }
        }
      ]
    }
  }

  tcpEgresses: [
    for _, e in TCPEgresses {
      _dk: "\(ServiceName)-egress-tcp-to-\(e.cluster)"
      domain: #Domain & {
        domain_key: _dk
        port: e.tcpPort
      }
      listener: #Listener & {
        name: _dk
        domain_keys: [_dk]
        port: e.tcpPort
        active_network_filters: [
          "envoy.tcp_proxy"
        ]
        network_filters: {
          envoy_tcp_proxy: {
            cluster: e.cluster
            stat_prefix: e.cluster
          }
        }
      }
      clusters: [...#Cluster] & [
        {
          name: e.cluster
          cluster_key: "\(ServiceName)-to-\(e.cluster)"
          if e.isExternal {
            instances: [
              {
                host: e.host
                port: e.port
              }
            ]
          }
        }
      ]
      routes: [...#Route] & [
        {
          _rk: "\(ServiceName)-to-\(e.cluster)"
          route_key: _rk
          domain_key: _dk
          route_match: {
            path: "/"
            match_type: "prefix"
          }
          rules: [
            {
              constraints: {
                light: [
                  {
                    cluster_key: _rk
                    weight: 1
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
