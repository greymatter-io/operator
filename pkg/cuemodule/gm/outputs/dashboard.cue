// Grey Matter configuration for Dashboard's sidecar

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
    _tcp_upstream: defaults.redis_cluster_name // NB this points at a cluster name, not key
  },

  // shared proxy object
  #proxy & {
    proxy_key: Name,
    domain_keys: [LocalName, EgressToRedisName]
    listener_keys: [LocalName, EgressToRedisName]
  },


  // edge->sidecar
  #cluster & {
    cluster_key: Name
    _spire_other: Name
  },
  #route & { domain_key: "edge", route_key: Name },

]
