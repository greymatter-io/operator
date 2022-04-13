
package only

import "greymatter.io/operator/greymatter-cue/greymatter"

/////////////////////////////////////////////////////////////
// "Functions" for Grey Matter config objects with defaults
// representing an ingress to a local service. Override-able.
/////////////////////////////////////////////////////////////

#domain: greymatter.#Domain & {
  domain_key: string
  name: string | *"*"
  port: int | *defaults.ports.default_ingress
  zone_key: mesh.spec.zone
}

#listener: greymatter.#Listener & {
  _tcp_upstream?: string // for TCP listeners, you can just specify the upstream cluster
  listener_key: string
  name: listener_key
  ip: string | *"0.0.0.0"
  port: int | *defaults.ports.default_ingress
  domain_keys: [...string] | *[listener_key]

  // if there's a tcp cluster, 
  if _tcp_upstream != _|_ {
    active_network_filters: ["envoy.tcp_proxy"]
    network_filters: envoy_tcp_proxy: {
      cluster: _tcp_upstream
      stat_prefix: _tcp_upstream
    }
  }
  // if there isn't a tcp cluster, then assume http filters, and provide the usual defaults
  if _tcp_upstream == _|_ {
    active_http_filters: [...string] | *[ "gm.metrics" ]
    http_filters: {
      gm_metrics: {
        metrics_host: "0.0.0.0" // TODO are we still scraping externally? If not, set this to 127.0.0.1
        metrics_port: 8081
        metrics_dashboard_uri_path: "/metrics"
        metrics_prometheus_uri_path: "prometheus" // TODO slash or no slash?
        metrics_ring_buffer_size: 4096
        prometheus_system_metrics_interval_seconds: 15
        metrics_key_function: "depth"
        metrics_key_depth: string | *"1"
        metrics_receiver: {
          redis_connection_string: string | *"redis://127.0.0.1:\(defaults.ports.redis_ingress)"
          push_interval_seconds: 10
        }
      }
    }
  }
  zone_key: mesh.spec.zone
  protocol: "http_auto" // vestigial
}

#cluster: greymatter.#Cluster & {
  // You can either specify the upstream with these, or leave it to service discovery
  _upstream_host: string | *"127.0.0.1"
  _upstream_port: int
  cluster_key: string
  name: cluster_key
  instances: [...greymatter.#Instance] | *[]
  if _upstream_port != _|_ {
    instances: [{ host: _upstream_host, port: _upstream_port }]
  } 
  zone_key: mesh.spec.zone
}

#route: greymatter.#Route & {
  route_key: string
  domain_key: string | *route_key
  route_match: {
    path: string | *"/"
    match_type: string | *"prefix"
  }
  rules: [{
    constraints: light: [{
      cluster_key: route_key
      weight: 1
    }]
  }]
  zone_key: mesh.spec.zone
  prefix_rewrite: string | *"/"
}

#proxy: greymatter.#Proxy & {
  proxy_key: string
  name: proxy_key
  domain_keys: [...string] | *[proxy_key] // TODO how to get more in here for, e.g., the extra egresses?
  listener_keys: [...string] | *[proxy_key]
  zone_key: mesh.spec.zone
}



#secret: {
	_name:    string
	_subject: string
	set_current_client_cert_details?: {...}
	forward_client_cert_details?: string

	secret_validation_name: "spiffe://greymatter.io"
	secret_name:            "spiffe://greymatter.io/\(mesh.metadata.name).\(_name)"
	subject_names: ["spiffe://greymatter.io/\(mesh.metadata.name).\(_subject)"]
	ecdh_curves: ["X25519:P-256:P-521:P-384"]
}