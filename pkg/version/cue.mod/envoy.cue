package base

import "strconv"

envoyEdge: envoy & {
	static_resources: {
		clusters: [
			envoyCluster & {
				_name: "xds_cluster"
				_host: sidecar.controlHost
				_port: 50000
			},
			envoyCluster & {
				_name: "control"
				_host: sidecar.controlHost
				_port: 5555
			},
			envoyCluster & {
				_name: "catalog"
				_host: "catalog.\(InstallNamespace).svc.cluster.local"
				_port: 8080
			}
		]
		listeners: [
			envoyHTTPListener & {
				_name: "edge"
				_port: 10808
				_routes: [
					{
						match: {
							prefix: "/services/control/api/"
							case_sensitive: false
						}
						route: {
							cluster: "control"
							prefix_rewrite: "/"
							timeout: "60s"
						}
					},
					{
						match: {
							prefix: "/services/catalog/"
							case_sensitive: false
						}
						route: {
							cluster: "catalog"
							prefix_rewrite: "/"
							timeout: "60s"
						}
					}
				]
			}
		]
	}
}

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
	type: "STRICT_DNS"
	connect_timeout: "5s"
	if _name == "xds_cluster" {
		http2_protocol_options: {}
	}
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

envoyHTTPListener: listener & {
	_name: string
	_port: int
	_routes: [...{...}]
	_key: "\(_name)-\(_port)"

	filter_chains: [
		{
			filters: [
				{
					name: "envoy.filters.network.http_connection_manager"
					typed_config: {
						"@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
						stat_prefix: _key
						codec_type: "AUTO"
						route_config: {
							name: "\(_name):\(_port)"
							virtual_hosts: [
								{
									name: "*-\(_port)"
									domains: ["*", "*:\(_port)"]
									routes: _routes
								}
							]
						}
						http_filters: [
              {
                name: "envoy.filters.http.cors",
                typed_config: "@type": "type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors"
              },
							{
								name: "envoy.filters.http.router",
								typed_config: {
									"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router"
									upstream_log: [
										{
											name: "envoy.access_loggers.file"
											typed_config: {
												"@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog",
                        path: "/dev/stdout"
                      }
										}
									]
								}
							}
						]
					}
				}
			]
		}
	]
}

envoyTCPListener: listener & {
	_cluster: string

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

listener: {
	_name: string
	_port: int

	name: "\(_name):\(_port)"
	address: socket_address: {
		address:    "0.0.0.0"
		port_value: _port
	}
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
		lds_config: {
			ads: {}
			resource_api_version: "V3"
		}
		cds_config: {
			ads: {}
			resource_api_version: "V3"
		}
		ads_config: {
			api_type: "GRPC"
			grpc_services: [
				{
					envoy_grpc: cluster_name: "xds_cluster"
				}
			]
			transport_api_version: "V3"
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
