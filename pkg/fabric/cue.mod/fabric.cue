package fabric

import (
  // "strconv"
  "greymatter.io/operator/fabric/api"
)

edge: #Tmpl & { _name: "edge" }

service: #Tmpl & {
  _name: ServiceName
  _ports: ServicePorts
}

#Tmpl: {
  _name: string
  _ports: [...int32]
  proxy: api.#Proxy & {
    name: _name
    zone_key: Zone
    domain_keys: [_name]
    listener_keys: [_name]
  }
  domain: api.#Domain & {
    domain_key: _name
    zone_key: Zone
    name: "*"
    port: MeshPort
  }
  listener: api.#Listener & {
    name: _name
    zone_key: Zone
    port: domain.port
    domain_keys: [_name]
  }
  cluster: api.#Cluster & {
    name: _name
    zone_key: Zone
  }
  if _name != "edge" && _name != "dashboard" {
    route: api.#Route & {
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
    route: api.#Route & {
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
  locals: [
    for _, Port in _ports if len(_ports) > 0 {
      {
        cluster: api.#Cluster & {
          name: "\(_name)-\(Port)"
          zone_key: Zone
          instances: [
            {
              host: "127.0.0.1"
              port: Port
            }
          ]
        }
        route: api.#Route & {
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
    }
  ]
}
