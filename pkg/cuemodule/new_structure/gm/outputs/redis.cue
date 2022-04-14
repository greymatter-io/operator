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
    // custom secret instead of listener helpers because we need to accept multiple subjects in this listener
    if flags.spire {
      secret: #spire_secret & {
        _name: Name
        _subjects: ["dashboard", "catalog", "controlensemble", "edge"]
      }
    }
  },
  #proxy & {
    proxy_key: Name
    domain_keys: [RedisIngressName]
    listener_keys: [RedisIngressName] 
  },
]