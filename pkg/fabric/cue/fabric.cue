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
  isTCP: bool
  cluster: string
  externalHost: string
  externalPort: int
}

ServiceName: string
HttpFilters: #HttpFilters
NetworkFilters: #NetworkFilters
Ingresses: [string]: int32
HTTPLocalEgresses: [...#EgressArgs]
HTTPExternalEgresses: [...#EgressArgs]

// Outputs

edgeDomain: #Domain & {
  domain_key: "edge"
  zone_key: Zone
  name: "*"
  port: MeshPort
}

service: #Tmpl & {
  _name: ServiceName
  _ingresses: Ingresses
  _httpLocalEgresses: HTTPLocalEgresses
  _httpExternalEgresses: HTTPExternalEgresses
}

#Tmpl: {
  _name: string
  _ingresses: [string]: int32
  _httpLocalEgresses: [...#EgressArgs]
  _httpExternalEgresses: [...#EgressArgs]

  catalogservice: #CatalogService & {
    mesh_id: MeshName
    service_id: _name

    if _name == "edge" {
      name: "Grey Matter Edge"
      description: "Handles north/south traffic flowing through the mesh."
      api_endpoint: "/"
    }
    if _name == "jwt-security" {
      name: "Grey Matter JWT Security"
      description: "A JWT token generation and retrieval service."
      api_endpoint: "/services/jwt-security/"
    }
    if _name == "control" {
      name: "Grey Matter Control"
      description: "Manages the configuration of the Grey Matter data plane."
      api_endpoint: "/services/control/api/"
    }
    if _name == "catalog" {
      name: "Grey Matter Catalog"
      description: "Interfaces with the control plane to expose the current state of the mesh."
      api_endpoint: "/services/catalog/"
    }
    if _name == "dashboard" {
      name: "Grey Matter Dashboard"
      description: "A user dashboard that paints a high-level picture of the mesh."
    }
    if _name == "gm-redis" {
      name: "Redis"
      description: "A data store for caching Grey Matter core service configurations."
    }
  }

  proxy: #Proxy & {
    name: _name
    zone_key: Zone
    domain_keys: [
      _name,
      if len(_httpLocalEgresses) > 0 {
        "\(_name)-http-local-egress",
      }
      if len(_httpExternalEgresses) > 0 {
        "\(_name)-http-external-egress",
      }
    ]
    listener_keys: [
      _name,
      if len(_httpLocalEgresses) > 0 {
        "\(_name)-http-local-egress",
      }
      if len(_httpExternalEgresses) > 0 {
        "\(_name)-http-external-egress",
      }
    ]
  }
  domain: #Domain & {
    domain_key: _name
    zone_key: Zone
    name: "*"
    port: MeshPort
  }
  listener: #Listener & {
    name: _name
    zone_key: Zone
    port: domain.port
    domain_keys: [_name]
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
        if _name == "edge" {
          metrics_key_depth: 1
        }
        if _name != "edge" {
          metrics_key_depth: 3
        }
      }
    }
    active_network_filters: [
      if NetworkFilters["envoy.tcp_proxy"] && len(_ingresses) == 1 {
        "envoy.tcp_proxy"
      }
    ]
    network_filters: {
      if NetworkFilters["envoy.tcp_proxy"] && len(_ingresses) == 1 {
        envoy_tcp_proxy: {
          for k, v in _ingresses {
            let key = "\(_name)-\(k)"
            cluster: key
            stat_prefix: key
          }
        }
      }
    }
  }
  clusters: [...#Cluster] & [
    {
      name: _name
      zone_key: Zone
    }
  ]
  routes: [...#Route] & [
    if _name != "dashboard" {
      {
        route_key: _name
        domain_key: "edge"
        zone_key: Zone
        route_match: {
          path: "/services/\(_name)/"
          match_type: "prefix"
        }
        redirects: [
          {
            from: "^/services/\(_name)$"
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
                  cluster_key: _name
                  weight: 1
                }
              ]
            }
          }
        ]
      }
    }
    if _name == "dashboard" {
      {
        route_key: _name
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
                  cluster_key: _name
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
    for k, v in _ingresses if len(_ingresses) > 0 {
      let key = "\(_name)-\(k)"
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
          if len(_ingresses) == 1 {
            {
              route_key: clusters[0].name
              domain_key: _name
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
          if len(_ingresses) > 1 {
            {
              route_key: clusters[0].name
              domain_key: _name
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
  
  httpLocalEgresses: {
    if len(_httpLocalEgresses) > 0 {
      let key = "\(_name)-http-local-egress"
      domain: #Domain & {
        zone_key: Zone
        domain_key: key
        port: 10818
      }
      listener: #Listener & {
        zone_key: Zone
        name: key
        listener_key: key
        domain_keys: [key]
        port: 10818
      }
      routes: [...#Route] & [
        for _, e in _httpLocalEgresses {
          {
            route_key: "\(key)-to-\(e.cluster)"
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
                      cluster_key: e.cluster
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

  httpExternalEgresses: {
    if len(_httpExternalEgresses) > 0 {
      let key = "\(_name)-http-external-egress"
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
        for _, e in _httpExternalEgresses {
          {
            name: "\(_name)-to-\(e.cluster)"
            zone_key: Zone
            instances: [
              {
                host: e.externalHost
                port: e.externalPort
              }
            ]
          }
        }
      ]
      routes: [...#Route] & [
        for _, e in _httpExternalEgresses {
          {
            route_key: "\(key)-to-\(e.cluster)"
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
                      cluster_key: "\(_name)-to-\(e.cluster)"
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
