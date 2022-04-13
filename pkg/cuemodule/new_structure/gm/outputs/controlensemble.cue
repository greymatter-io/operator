package only

let Name = "controlensemble"
let CatalogIngressName = "\(Name)_catalog"
let RedisIngressName = "\(Name)_redis"

controlensemble_config: [

  // Catalog HTTP ingress
  #domain & { domain_key: CatalogIngressName },
  #listener & {
    listener_key: CatalogIngressName
    _spire_self: Name
  },
  #cluster & { cluster_key: CatalogIngressName, _upstream_port: 8080 },
  #route & { route_key: CatalogIngressName }, // TODO needs special routematch

  // Redis TCP ingress
  #domain & { domain_key: RedisIngressName, port: defaults.ports.redis_ingress },
  #cluster & { cluster_key: RedisIngressName, _upstream_port: 6379},
  #route & { route_key: RedisIngressName }, // unused route must exist for the cluster to be registered
  #listener & {
    listener_key: RedisIngressName 
    port: defaults.ports.redis_ingress
    _tcp_upstream: RedisIngressName // this _actually_ connects the cluster to the listener
  },

  // Proxy object shared between Catalog and Redis ingresses
  #proxy & {
    proxy_key: Name
    domain_keys: [RedisIngressName, CatalogIngressName]
    listener_keys: [RedisIngressName, CatalogIngressName] 
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