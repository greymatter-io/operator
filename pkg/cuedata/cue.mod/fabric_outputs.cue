package fabric

edgeDomain: #Domain & {
  domain_key: "edge"
  zone_key: Zone
  name: "*"
  port: 10808
}

service: {
  catalogservice: #CatalogService & {
    mesh_id: MeshName
    service_id: ServiceName

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
    if ServiceName == "gm-redis" {
      name: "Redis"
      description: "A data store for caching Grey Matter core service configurations."
    }
  }

  proxy: #Proxy & {
    name: ServiceName
    zone_key: Zone
    domain_keys: [
      ServiceName,
      if len(HTTPEgresses) > 0 {
        "\(ServiceName)-egress-http",
      }
      for _, e in TCPEgresses {
        if e.isExternal {
          "\(ServiceName)-egress-tcp-to-external-\(e.cluster)"
        }
        if !e.isExternal {
          "\(ServiceName)-egress-tcp-to-\(e.cluster)"
        }
      }
    ]
    listener_keys: [
      ServiceName,
      if len(HTTPEgresses) > 0 {
        "\(ServiceName)-egress-http",
      }
      for _, e in TCPEgresses {
        if e.isExternal {
          "\(ServiceName)-egress-tcp-to-external-\(e.cluster)"
        }
        if !e.isExternal {
          "\(ServiceName)-egress-tcp-to-\(e.cluster)"
        }
      }
    ]
  }

  domain: #Domain & {
    domain_key: ServiceName
    zone_key: Zone
    name: "*"
    port: 10808
  }
  listener: #Listener & {
    name: ServiceName
    zone_key: Zone
    port: domain.port
    domain_keys: [ServiceName]
    active_http_filters: [
      if ServiceName != "gm-redis" {
        "gm.metrics"
      }
    ]
    http_filters: {
      if ServiceName != "gm-redis" {
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
              // TODO: Use NATS for the metrics_receiver universally instead of Redis.
              // No external NATS option is required since it's an event bus, not a DB.
              redis_connection_string: "redis://:\(Redis.password)@127.0.0.1:10910"
              push_interval_seconds: 10
            }
          }
        }
      }
    }
    active_network_filters: [
      if NetworkFilters["envoy.tcp_proxy"] && len(Ingresses) == 1 {
        "envoy.tcp_proxy"
      }
    ]
    network_filters: {
      if NetworkFilters["envoy.tcp_proxy"] && len(Ingresses) == 1 {
        envoy_tcp_proxy: {
          for k, v in Ingresses {
            let key = "\(ServiceName):\(v)"
            cluster: key
            stat_prefix: key
          }
        }
      }
    }
  }

  clusters: [...#Cluster] & [
    {
      name: ServiceName
      zone_key: Zone
    }
  ]

  routes: [...#Route] & [
    if ServiceName != "dashboard" && ServiceName != "edge" {
      {
        route_key: ServiceName
        domain_key: "edge"
        zone_key: Zone
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
        zone_key: Zone
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
        let key = "\(ServiceName):\(v)"
        {
          name: key
          zone_key: Zone
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
        let key = "\(ServiceName):\(v)"
        if len(Ingresses) == 1 {
          {
            route_key: key
            domain_key: ServiceName
            zone_key: Zone
            route_match: {
              path: "/"
              match_type: "prefix"
            }
            rules: [
              {
                constraints: {
                  light: [
                    {
                      cluster_key: key
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
            route_key: key
            domain_key: ServiceName
            zone_key: Zone
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
                      cluster_key: key
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
    if len(HTTPEgresses) > 0 {
      let key = "\(ServiceName)-egress-http"
      domain: #Domain & {
        zone_key: Zone
        domain_key: key
        name: "*"
        port: 10909
      }
      listener: #Listener & {
        name: key
        zone_key: Zone
        port: 10909
        domain_keys: [key]

        // Temp kludge: Enable the metrics filter and receiver for gm-metrics.
        // This is required here since we are mocking an http listener for the metrics_receiver.
        // If TCP is configured on a listener, no HTTP metrics filter is set :/
        if ServiceName == "gm-redis" && ReleaseVersion != "1.6" {
          active_http_filters: ["gm.metrics"]
          http_filters: {
            gm_metrics: {
              metrics_host: "0.0.0.0"
              metrics_port: 8081
              metrics_dashboard_uri_path: "/metrics"
              metrics_prometheus_uri_path: "prometheus"
              metrics_ring_buffer_size: 4096
              prometheus_system_metrics_interval_seconds: 15
              metrics_key_function: "depth"
              metrics_key_depth: "3"
              metrics_receiver: {
                redis_connection_string: "redis://:\(Redis.password)@127.0.0.1:10808"
                push_interval_seconds: 10
              }
            }
          }
        }
      }
      clusters: [...#Cluster] & [
        for _, e in HTTPEgresses if e.isExternal {
          {
            name: "\(ServiceName)-to-external-\(e.cluster)"
            zone_key: Zone
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
        for _, e in HTTPEgresses {
          {
            if e.isExternal {
              route_key: "\(ServiceName)-to-external-\(e.cluster)"
            }
            if !e.isExternal {
              route_key: "\(ServiceName)-to-\(e.cluster)"
            }
            domain_key: key
            zone_key: Zone
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
                      if e.isExternal {
                        cluster_key: "\(ServiceName)-to-external-\(e.cluster)"
                      }
                      if !e.isExternal {
                        cluster_key: e.cluster
                      }
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
      _key: string
      if e.isExternal {
        _key: "\(ServiceName)-egress-tcp-to-external-\(e.cluster)"
      }
      if !e.isExternal {
        _key: "\(ServiceName)-egress-tcp-to-\(e.cluster)"
      }
      domain: #Domain & {
        zone_key: Zone
        domain_key: _key
        port: e.tcpPort
      }
      listener: #Listener & {
        zone_key: Zone
        name: _key
        listener_key: _key
        domain_keys: [_key]
        port: e.tcpPort
        active_network_filters: [
          "envoy.tcp_proxy"
        ]
        network_filters: {
          envoy_tcp_proxy: {
            if e.isExternal {
              cluster: "\(ServiceName)-to-external-\(e.cluster)"
              stat_prefix: "\(ServiceName)-to-external-\(e.cluster)"
            }
            if !e.isExternal {
              cluster: e.cluster
              stat_prefix: e.cluster
            }
          }
        }
      }
      clusters: [...#Cluster] & [
        if e.isExternal {
          {
            name: "\(ServiceName)-to-external-\(e.cluster)"
            zone_key: Zone
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
          if e.isExternal {
            route_key: "\(ServiceName)-to-external-\(e.cluster)"
          }
          if !e.isExternal {
            route_key: "\(ServiceName)-to-\(e.cluster)"
          }
          domain_key: _key
          zone_key: Zone
          route_match: {
            path: "/"
            match_type: "prefix"
          }
          rules: [
            {
              constraints: {
                light: [
                  {
                    if e.isExternal {
                      cluster_key: "\(ServiceName)-to-external-\(e.cluster)"
                    }
                    if !e.isExternal {
                      cluster_key: e.cluster
                    }
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
