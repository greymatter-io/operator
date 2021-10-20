// Inputs

MeshName: string

Zone: *"default-zone" | string

MeshPort: *10808 | int32

ServiceName: string
ServiceIngresses: [string]: int32
ServiceLocalEgresses: [...string]
ServiceExternalEgresses: [...string]

// Outputs

edgeDomain: #Domain & {
  domain_key: "edge"
  zone_key: Zone
  name: "*"
  port: MeshPort
}

service: #Tmpl & {
  _name: ServiceName
  _ingresses: ServiceIngresses
  _localEgresses: ServiceLocalEgresses
  _externalEgresses: ServiceExternalEgresses
}

#Tmpl: {
  _name: string
  _ingresses: [string]: int32
  _localEgresses: [...string]
  _externalEgresses: [...string]

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
      if len(_localEgresses) > 0 {
        "\(_name)-http-local-egress",
      }
    ]
    listener_keys: [
      _name,
      if len(_localEgresses) > 0 {
        "\(_name)-http-local-egress",
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
  }
  cluster: #Cluster & {
    name: _name
    zone_key: Zone
  }
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
      let key = "\(_name)-\(v)"
      "\(key)": {
        cluster: #Cluster & {
          name: key
          zone_key: Zone
          instances: [
            {
              host: "127.0.0.1"
              port: v
            }
          ]
        }
        routes: [...#Route] & [
          if len(_ingresses) == 1 {
            {
              route_key: cluster.name
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
                        cluster_key: cluster.name
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
              route_key: cluster.name
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
                        cluster_key: cluster.name
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
  
  localEgresses: {
    if len(_localEgresses) > 0 {
      let key = "\(_name)-http-local-egress"
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
      routes: [...#Route] & [
        for _, cluster in _localEgresses {
          {
            route_key: "\(key)-to-\(cluster)"
            domain_key: key
            zone_key: Zone
            route_match: {
              path: "/\(cluster)/"
              match_type: "prefix"
            }
            redirects: [
              {
                from: "^/\(cluster)$"
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
                      cluster_key: cluster
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

  externalEgresses: {}
}
