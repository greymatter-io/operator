// Inputs

MeshName: string
Zone: *"default-zone" | string
MeshPort: *10808 | int32

#HttpFilters: {
  "gm.metrics": true
}

#NetworkFilters: {
  "envoy.tcp_proxy": *false | bool
}

#EgressArgs: {
  isExternal: bool
  cluster: string
  host: string
  port: int
  tcpPort: int
}

ServiceName: string
HttpFilters: #HttpFilters
NetworkFilters: #NetworkFilters
Ingresses: [string]: int32
HTTPEgresses: [...#EgressArgs]
TCPEgresses: [...#EgressArgs]

// Outputs

edgeDomain: #Domain & {
  domain_key: "edge"
  zone_key: Zone
  name: "*"
  port: MeshPort
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
    }
    if ServiceName == "control" {
      name: "Grey Matter Control"
      description: "Manages the configuration of the Grey Matter data plane."
      api_endpoint: "/services/control/api/"
    }
    if ServiceName == "catalog" {
      name: "Grey Matter Catalog"
      description: "Interfaces with the control plane to expose the current state of the mesh."
      api_endpoint: "/services/catalog/"
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
        "\(ServiceName)-http-egress",
      }
      for _, e in TCPEgresses if len(TCPEgresses) > 0 {
        if e.isExternal {
          "\(ServiceName)-tcp-egress-to-external-\(e.cluster)"
        }
        if !e.isExternal {
          "\(ServiceName)-tcp-egress-to-\(e.cluster)"
        }
      }
    ]
    listener_keys: [
      ServiceName,
      if len(HTTPEgresses) > 0 {
        "\(ServiceName)-http-egress",
      }
      for _, e in TCPEgresses if len(TCPEgresses) > 0 {
        if e.isExternal {
          "\(ServiceName)-tcp-egress-to-external-\(e.cluster)"
        }
        if !e.isExternal {
          "\(ServiceName)-tcp-egress-to-\(e.cluster)"
        }
      }
    ]
  }

  domain: #Domain & {
    domain_key: ServiceName
    zone_key: Zone
    name: "*"
    port: MeshPort
  }
  listener: #Listener & {
    name: ServiceName
    zone_key: Zone
    port: domain.port
    domain_keys: [ServiceName]
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
          metrics_key_depth: 1
        }
        if ServiceName != "edge" {
          metrics_key_depth: 3
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
            let key = "\(ServiceName)-\(k)"
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
    if ServiceName != "dashboard" {
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
    for k, v in Ingresses if len(Ingresses) > 0 {
      let key = "\(ServiceName)-\(k)"
      "\(key)": {
        clusters: [...#Cluster] & [
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
        ]
        routes: [...#Route] & [
          if len(Ingresses) == 1 {
            {
              route_key: clusters[0].name
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
                        cluster_key: clusters[0].name
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
              route_key: clusters[0].name
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
                        cluster_key: clusters[0].name
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
  }

  httpEgresses: {
    if len(HTTPEgresses) > 0 {
      let key = "\(ServiceName)-http-egress"
      domain: #Domain & {
        zone_key: Zone
        domain_key: key
        port: 10909
      }
      listener: #Listener & {
        zone_key: Zone
        name: key
        listener_key: key
        domain_keys: [key]
        port: 10909
      }
      clusters: [...#Cluster] & [
        for _, e in HTTPEgresses {
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
    for _, e in TCPEgresses if len(TCPEgresses) > 0 {
      _key: string
      if e.isExternal {
        _key: "\(ServiceName)-tcp-egress-to-external-\(e.cluster)"
      }
      if !e.isExternal {
        _key: "\(ServiceName)-tcp-egress-to-\(e.cluster)"
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
