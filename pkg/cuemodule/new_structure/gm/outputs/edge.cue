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
  #route & { route_key: EgressToRedisName }, // unused route must exist for the cluster to be registered
  #cluster & {
    cluster_key: EgressToRedisName,
    _upstream_host: "controlensemble.\(mesh.spec.install_namespace).svc.cluster.local"
    _upstream_port: defaults.ports.redis_ingress
  },
  #listener & {
    listener_key: EgressToRedisName
    ip: "127.0.0.1" // egress listeners are local-only
    port: defaults.ports.redis_ingress
    _tcp_upstream: EgressToRedisName
  },

  #proxy & {
    proxy_key: Name
    domain_keys: [Name, EgressToRedisName]
    listener_keys: [Name, EgressToRedisName]
  },
]