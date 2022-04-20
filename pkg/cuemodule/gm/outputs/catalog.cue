// Grey Matter configuration for Catalog's sidecar

package only

let Name = "catalog"
let CatalogIngressName = "\(Name)_local"
let EgressToRedisName = "\(Name)_egress_to_redis"

catalog_config: [

  // Catalog HTTP ingress
  #domain & { domain_key: CatalogIngressName },
  #listener & {
    listener_key: CatalogIngressName
    _spire_self: Name
  },
  #cluster & { cluster_key: CatalogIngressName, _upstream_port: 8080 },
  #route & { route_key: CatalogIngressName },


  // egress->redis
  #domain & { domain_key: EgressToRedisName, port: defaults.ports.redis_ingress },
  #cluster & {
    cluster_key: EgressToRedisName
    name: defaults.redis_cluster_name
    _spire_self: Name
    _spire_other: defaults.redis_cluster_name
  },
  #route & { route_key: EgressToRedisName }, // unused route must exist for the cluster to be registered with sidecar
  #listener & {
    listener_key: EgressToRedisName
    ip: "127.0.0.1" // egress listeners are local-only
    port: defaults.ports.redis_ingress
    _tcp_upstream: defaults.redis_cluster_name
  },

  // shared proxy object
  #proxy & {
    proxy_key: Name
    domain_keys: [CatalogIngressName, EgressToRedisName]
    listener_keys: [CatalogIngressName, EgressToRedisName]
  },

  // Edge config for catalog ingress
  #cluster & {
    cluster_key: Name
    _spire_other: Name
    },
  #route & {
    domain_key: "edge",
    route_key: Name
    route_match: {
      path: "/services/catalog/"
    }
    redirects: [
      {
        from: "^/services/catalog$"
        to: route_match.path
        redirect_type: "permanent"
      }
    ]
    prefix_rewrite: "/"
  }
]
