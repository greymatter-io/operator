package only

let Name= "dashboard"
let LocalName = "dashboard_local"
let EgressToRedisName = "\(Name)_egress_to_redis"

dashboard_config: [
  // sidecar->dashboard
  #domain & { domain_key: LocalName },
  #listener & {
    listener_key: LocalName
    _spire_self: Name
  },
  #cluster & { cluster_key: LocalName, _upstream_port: 1337 },
  #route & { route_key: LocalName },

  // edge->sidecar
  #cluster & {
    cluster_key: Name
    _spire_other: Name
  },
  #route & { domain_key: "edge", route_key: Name },

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

  // shared proxy object
  #proxy & {
    proxy_key: Name,
    domain_keys: [LocalName, EgressToRedisName] // TODO seems like a mess now that defaults aren't for local. rework.
    listener_keys: [LocalName, EgressToRedisName]
  },
]
