// All necessary Grey Matter configuration for an injected sidecar
// created during deployment assist.

package only

#sidecar_config: {
  Name: string | *"edge" // workaround for CUE's behavior with conflicting defaults
  Port: int | *8080
  LocalName: "\(Name)_local"
  EgressToRedisName: "\(Name)_egress_to_redis"

  objects: [
    // sidecar -> local service
    #domain & { domain_key: LocalName },
    #listener & {
      listener_key: LocalName,
      _spire_self: Name
    },
    #cluster & { cluster_key: LocalName, _upstream_port: Port },
    #route & { route_key: LocalName},

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

  // proxy shared between local ingress and redis egress
    #proxy & {
      proxy_key: Name
      domain_keys: [LocalName, EgressToRedisName]
      listener_keys: [LocalName, EgressToRedisName] 
    },

    // edge->sidecar
    #cluster & {
      cluster_key: Name,
      _spire_other: Name
    },
    #route & {
      domain_key: "edge",
      route_key: Name // destination cluster name is the same as route_key by default
      route_match: {
        path: "/services/\(Name)/"
      }
      redirects: [
        {
          from: "^/services/\(Name)$"
          to: route_match.path
          redirect_type: "permanent"
        }
      ]
      prefix_rewrite: "/"
    },

    #CatalogService & { // HACK this is my own version from catalogentries.cue - fix greymatter.cue version
      name:             Name  // TODO nice human readable name - let them pass it?
      mesh_id: mesh.metadata.name
      service_id: Name // Catalog lives behind the controlensemble sidecar, for the moment
      //version:         strings.Split(mesh.spec.images.catalog, ":")[1] // TODO let them pass a version?
      //description:       "Interfaces with the control plane to expose the current state of the mesh."
      api_endpoint:      "/services/\(Name)/"
      api_spec_endpoint: "/services/\(Name)/"
      business_impact:   "low"
      enable_instance_metrics: true
      enable_historical_metrics: false
    },
  ]
}