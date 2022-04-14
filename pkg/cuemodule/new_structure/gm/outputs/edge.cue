package only


let Name= "edge"
let EgressToRedisName = "\(Name)_egress_to_redis"

edge_config: [
  #domain & { domain_key: Name },
  #listener & { listener_key: Name },
  // This cluster must exist (though it never receives traffic)
  // so that Catalog will be able to look-up edge instances
  #cluster & { cluster_key: Name }, 

  // egress->redis
  #domain & { domain_key: EgressToRedisName, port: defaults.ports.redis_ingress },
  #route & { // unused route must exist for the cluster to be registered with sidecar
    route_key: EgressToRedisName
    _upstream_cluster_key: defaults.redis_cluster_name
  },
  #listener & {
    listener_key: EgressToRedisName
    ip: "127.0.0.1" // egress listeners are local-only
    port: defaults.ports.redis_ingress
    _tcp_upstream: defaults.redis_cluster_name
  },
  // (the actual reusable redis cluster was defined once in the redis config)

  #proxy & {
    proxy_key: Name
    domain_keys: [Name, EgressToRedisName]
    listener_keys: [Name, EgressToRedisName]
  },
]