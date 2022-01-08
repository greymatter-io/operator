// Inputs

// Values pre-defined from version.Version
MeshName: string
ReleaseVersion: string
Zone: string
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

// Identify core service versions for each Grey Matter release
ServiceVersions: {
  if ReleaseVersion == "1.7" {
    edge: "1.7.0"
    control: "1.7.0"
    catalog: "3.0.0"
    dashboard: "6.0.0"
    jwtsecurity: "1.3.0"
  }
  if ReleaseVersion == "1.6" {
    edge: "1.6.3"
    control: "1.6.5"
    catalog: "2.0.1"
    dashboard: "5.1.1"
    jwtsecurity: "1.3.0"
  }
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
        version: ServiceVersions.edge
        description: "Handles north/south traffic flowing through the mesh."
        api_endpoint: "/"
        business_impact: "critical"
      }
      if ServiceName == "control" {
        name: "Grey Matter Control"
        version: ServiceVersions.control
        description: "Manages the configuration of the Grey Matter data plane."
        api_endpoint: "/services/control/api/v1.0/"
        api_spec_endpoint: "/services/control/api/"
        business_impact: "critical"
      }
      if ServiceName == "catalog" {
        name: "Grey Matter Catalog"
        version: ServiceVersions.catalog
        description: "Interfaces with the control plane to expose the current state of the mesh."
        api_endpoint: "/services/catalog/"
        api_spec_endpoint: "/services/catalog/"
        business_impact: "high"
      }
      if ServiceName == "dashboard" {
        name: "Grey Matter Dashboard"
        version: ServiceVersions.dashboard
        description: "A user dashboard that paints a high-level picture of the mesh."
        business_impact: "high"
      }
      if ServiceName == "jwt-security" {
        name: "Grey Matter JWT Security"
        version: ServiceVersions.jwtsecurity
        description: "A JWT token generation and retrieval service."
        api_endpoint: "/services/jwt-security/"
        api_spec_endpoint: "/services/jwt-security/"
        business_impact: "high"
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

    if ServiceName != "gm-redis" {
      active_http_filters: [
        "gm.metrics"
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
      }
    }

    if NetworkFilters["envoy.tcp_proxy"] && len(Ingresses) == 1 {
      active_network_filters: [
        "envoy.tcp_proxy"
      ]
      network_filters: {
        envoy_tcp_proxy: {
          for _, v in Ingresses {
            _key: "\(ServiceName):\(v)"
            cluster: _key
            stat_prefix: _key
          }
        }
      }
    }
  }

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

  ingresses: {
    clusters: [...#Cluster] & [
      for _, v in Ingresses if len(Ingresses) > 0 && ServiceName != "edge" {
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
    ]
    routes: [...#Route] & [
      for k, v in Ingresses if len(Ingresses) > 0 && ServiceName != "edge" {
        if len(Ingresses) == 1 {
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
        if len(Ingresses) > 1 {
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
