package only

let Name = defaults.redis_cluster_name
let RedisIngressName = "\(Name)_local"
// TODO spire mTLS config on redis listener with dynamically filled subjects


redis_config: [

  // Redis TCP ingress
  #domain & { domain_key: RedisIngressName, port: defaults.ports.redis_ingress },
  #cluster & { cluster_key: RedisIngressName, _upstream_port: 6379},
  #route & { route_key: RedisIngressName }, // unused route must exist for the cluster to be registered
  #listener & {
    listener_key: RedisIngressName 
    port: defaults.ports.redis_ingress
    _tcp_upstream: RedisIngressName // this _actually_ connects the cluster to the listener
  },
  #proxy & {
    proxy_key: Name
    domain_keys: [RedisIngressName]
    listener_keys: [RedisIngressName] 
  },

  // Reusable cluster for egress to redis - create a route to this cluster from each pod
  // TODO For mTLS, we will need to programmatically fill the allowable subjects
  #cluster & { cluster_key: Name },
]