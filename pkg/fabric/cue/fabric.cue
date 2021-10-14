// Inputs

Zone: *"default-zone" | string

MeshPort: *10808 | int32

ServiceName: string
ServiceIngresses: [string]: int32

// Outputs

edge: #Tmpl & { _name: "edge" }

service: #Tmpl & {
  _name: ServiceName
  _ports: ServiceIngresses
}

#Tmpl: {
  _name: string
  _ports: [string]: int32
  proxy: #Proxy & {
    name: _name
    zone_key: Zone
    domain_keys: [_name]
    listener_keys: [_name]
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
  if _name != "edge" && _name != "dashboard" {
    route: #Route & {
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
    route: #Route & {
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
  ingresses: {
    for k, v in _ports if len(_ports) > 0 {
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
        if len(_ports) == 1 {
          route: #Route & {
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
        if len(_ports) > 1 {
          route: #Route & {
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
      }
    }
  }
}
