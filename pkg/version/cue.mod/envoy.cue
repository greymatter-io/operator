package base

import "strconv"

envoyMeshConfigs: envoy & {
	static_resources: {
		clusters: [
			envoyCluster & {
				_name: "xds_cluster"
				_host: sidecar.controlHost
				_port: 50000
			},
			envoyCluster & {
				_name: "gm-redis"
				_host: Redis.host
				_port: strconv.Atoi(Redis.port)
			},
		]
		listeners: [
			envoyTCPListener & {
				_name: sidecar.xdsCluster
				_port: 10910
				_cluster: "gm-redis"
			}
		]
	}
}

envoyRedis: envoy & {
	static_resources: {
		clusters: [
			envoyCluster & {
				_name: "xds_cluster"
				_host: sidecar.controlHost
				_port: 50000
			},
			envoyCluster & {
				_name: "gm-redis:6379"
				_host: "127.0.0.1"
				_port: 6379
			}
		]
		listeners: [
			envoyTCPListener & {
				_name: "gm-redis"
				_port: 10910
				_cluster: "gm-redis:6379"
			}
		]
	}
}

envoyCluster: {
	_name: string
	_host: string
	_port: int

	name: _name
	http2_protocol_options: {}
	type: "STRICT_DNS"
	connect_timeout: "10s"
	lb_policy: "LEAST_REQUEST"
	load_assignment: {
		cluster_name: _name
		endpoints: [
			{
				lb_endpoints: [
					{
						endpoint: address: socket_address: {
							address:    _host
							port_value: _port
						}
					}
				]
			}
		]
	}
}

envoyTCPListener: {
	_name: string
	_port: int
	_cluster: string

	name: "\(_name):\(_port)"
	address: socket_address: {
		address:    "0.0.0.0"
		port_value: _port
	}
	filter_chains: [
		{
			filters: [
				{
					name: "envoy.filters.network.tcp_proxy"
					typed_config: {
						"@type":     "type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy"
						cluster:     _cluster
						stat_prefix: _cluster
					}
				}
			]
		}
	]
}

envoy: {
	node: {
		cluster: sidecar.xdsCluster
		id:      sidecar.node
		locality: {
			region: "default-region"
			zone:   Zone
		}
	}
	dynamic_resources: {
		ads_config: {
			api_type: "GRPC"
			grpc_services: envoy_grpc: cluster_name: "xds_cluster"
			transport_api_version: "V3"
		}
		cds_config: {
			ads: {}
			resource_api_version: "V3"
		}
		lds_config: {
			ads: {}
			resource_api_version: "V3"
		}
	}
	admin: {
		access_log_path: "/dev/stdout"
		address: socket_address: {
			address:    "127.0.0.1"
			port_value: 8001
		}
	}
}