package only

let Name= "dashboard"
let LocalName = "dashboard_local"
let EgressToRedisName = "\(Name)_egress_to_redis"

dashboard_config: [
  // sidecar->dashboard
  #domain & { domain_key: LocalName },
  #listener & { listener_key: LocalName },
  #cluster & { cluster_key: LocalName, _upstream_port: 1337 },
  #route & { route_key: LocalName },

  // edge->sidecar
  #cluster & { cluster_key: Name },
  #route & { domain_key: "edge", route_key: Name },

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
    proxy_key: Name,
    domain_keys: [LocalName, EgressToRedisName] // TODO seems like a mess now that defaults aren't for local. rework.
    listener_keys: [LocalName, EgressToRedisName]
  },
]

// TODO egress->redis