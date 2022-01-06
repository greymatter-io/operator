package base

envoyEdge: #envoy & {
	static_resources: {
		clusters: [
			#envoyCluster & {
				_name: "xds_cluster"
				_host: "control.\(InstallNamespace).svc.cluster.local"
				_port: 50000
			},
			#envoyCluster & {
				_name: "_control"
				_alt: "control"
				_host: "control.\(InstallNamespace).svc.cluster.local"
				_port: 10707
				_tlsContext: "spire"
				_spireSecret: "edge"
				_spireSubject: "control"
			},
			#envoyCluster & {
				_name: "_catalog"
				_alt: "catalog"
				_host: "catalog.\(InstallNamespace).svc.cluster.local"
				_port: 10707
				_tlsContext: "spire"
				_spireSecret: "edge"
				_spireSubject: "catalog"
			},
			#envoyCluster & {
				_name: "spire_agent"
			}
		]
		listeners: [
			#envoyHTTPListener & {
				_name: "bootstrap"
				_port: 10707
				_routes: [
					{
						match: {
							prefix: "/control/"
							case_sensitive: false
						}
						route: {
							cluster: "_control"
							prefix_rewrite: "/"
							timeout: "5s"
						}
					},
					{
						match: {
							prefix: "/catalog/"
							case_sensitive: false
						}
						route: {
							cluster: "_catalog"
							prefix_rewrite: "/"
							timeout: "5s"
						}
					}
				]
			}
		]
	}
}

envoyMeshConfig: #envoy & {
	static_resources: {
		clusters: [
			#envoyCluster & {
				_name: "xds_cluster"
				_host: sidecar.controlHost
				_port: 50000
			},
			#envoyCluster & {
				_name: "gm-redis"
				_host: "gm-redis.\(InstallNamespace).svc.cluster.local"
				_port: 10707
				_tlsContext: "spire"
				_spireSecret: sidecar.xdsCluster
				_spireSubject: "gm-redis"
			},
			#envoyCluster & {
				_name: "bootstrap"
				_alt: "\(sidecar.xdsCluster):\(sidecar.localPort)"
				_host: "127.0.0.1"
				_port: sidecar.localPort
			},
			#envoyCluster & {
				_name: "spire_agent"
			}
		]
		listeners: [
			#envoyTCPListener & {
				_name: sidecar.xdsCluster
				_port: 10910
				_cluster: "gm-redis"
			},
			if sidecar.xdsCluster == "control" || sidecar.xdsCluster == "catalog" {
				#envoyHTTPListener & {
					_name: "bootstrap"
					_port: 10707
					_routes: [
						{
							match: {
								prefix: "/"
								case_sensitive: false
							}
							route: {
								cluster: "bootstrap"
								timeout: "5s"
							}
						}
					]
					_tlsContext: "spire"
					_spireSecret: sidecar.xdsCluster
					_spireSubjects: ["edge"]
				}
			}
		]
	}
}

envoyRedis: #envoy & {
	static_resources: {
		clusters: [
			#envoyCluster & {
				_name: "xds_cluster"
				_host: sidecar.controlHost
				_port: 50000
			},
			#envoyCluster & {
				_name: "bootstrap"
				_alt: "gm-redis:6379"
				_host: "127.0.0.1"
				_port: 6379
			},
			#envoyCluster & {
				_name: "spire_agent"
			}
		]
		listeners: [
			#envoyTCPListener & {
				_name: "bootstrap"
				_port: 10707
				_cluster: "bootstrap"
				_tlsContext: "spire"
				_spireSecret: "gm-redis"
				_spireSubjects: ["control", "catalog", "jwt-security"]
			}
		]
	}
}

#envoyCluster: {
	_name: string
	_alt: *"" | string
	_host: string
	_port: int
	_tlsContext: *"" | string
	_spireSecret: string
	_spireSubject: string

	name: _name
	if _alt != "" {
		alt_stat_name: _alt
	}
	if _name == "spire_agent" {
		type: "STATIC"
	}
	if _name != "spire_agent" {
		type: "STRICT_DNS"
	}
	connect_timeout: "5s"
	if _name == "xds_cluster" || _name == "spire_agent" {
		http2_protocol_options: {}
	}
	load_assignment: {
		cluster_name: _name
		endpoints: [
			{
				lb_endpoints: [
					{
						if _name == "spire_agent" {
							endpoint: address: pipe: path: "/run/spire/socket/agent.sock"
						}
						if _name != "spire_agent" {
							endpoint: address: socket_address: {
								address:    _host
								port_value: _port
							}
						}
					}
				]
			}
		]
	}
	if _tlsContext != "" {
		transport_socket: {
			name: "envoy.transport_sockets.tls"
			typed_config: {
				"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext"
				common_tls_context: {
					if _tlsContext == "spire" {
						tls_params: ecdh_curves: ["X25519:P-256:P-521:P-384"]
						tls_certificate_sds_secret_configs: [
							{
								name: "spiffe://greymatter.io/\(MeshName).\(_spireSecret)"
								sds_config: {
									resource_api_version: "V3"
									api_config_source: #adsConfig & {
										_name: "spire_agent"
									}
								}
							}
						]
						combined_validation_context: {
							default_validation_context: match_subject_alt_names: [
								{ exact: "spiffe://greymatter.io/\(MeshName).\(_spireSubject)" }
							]
							validation_context_sds_secret_config: {
								name: "spiffe://greymatter.io"
								sds_config: {
									resource_api_version: "V3"
									api_config_source: #adsConfig & {
										_name: "spire_agent"
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

#envoyHTTPListener: #envoyListener & {
	_name: string
	_port: int
	_routes: [...{...}]
	_key: "\(_name)-\(_port)"
	_tlsContext: string

	filter_chains: [
		{
			filters: [
				{
					name: "envoy.filters.network.http_connection_manager"
					typed_config: {
						"@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
						stat_prefix: _key
						codec_type: "AUTO"
						if _tlsContext != "" {
							set_current_client_cert_details: uri: true
							forward_client_cert_details: "APPEND_FORWARD"
						}
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

#envoyTCPListener: #envoyListener & {
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

#envoyListener: {
	_name: string
	_port: int
	_tlsContext: *"" | string
	_spireSecret: string
	_spireSubjects: [...string]

	name: "\(_name):\(_port)"
	address: socket_address: {
		address:    "0.0.0.0"
		port_value: _port
	}
	filter_chains: [...{...}]

	if _tlsContext != "" {
		filter_chains: [
			{
				transport_socket: {
					name: "envoy.transport_sockets.tls"
					typed_config: {
						"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext"
						require_client_certificate: true
						common_tls_context: {
							if _tlsContext == "spire" {
								tls_params: ecdh_curves: ["X25519:P-256:P-521:P-384"]
								tls_certificate_sds_secret_configs: [
									{
										name: "spiffe://greymatter.io/\(MeshName).\(_spireSecret)"
										sds_config: {
											resource_api_version: "V3"
											api_config_source: #adsConfig & {
												_name: "spire_agent"
											}
										}
									}
								]
								combined_validation_context: {
									default_validation_context: match_subject_alt_names: [
										for _, subject in _spireSubjects {
											{ exact: "spiffe://greymatter.io/\(MeshName).\(subject)" }
										}
									]
									validation_context_sds_secret_config: {
										name: "spiffe://greymatter.io"
										sds_config: {
											resource_api_version: "V3"
											api_config_source: #adsConfig & {
												_name: "spire_agent"
											}
										}
									}
								}
							}
						}
					}
				}
			}
		]
	}
}

#adsConfig: {
	_name: string
	api_type: "GRPC"
	transport_api_version: "V3"
	grpc_services: envoy_grpc: cluster_name: _name
}

#envoy: {
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
		ads_config: #adsConfig & {
			_name: "xds_cluster"
		}
	}
	static_resources: {
		clusters: [...{...}]
		listeners: [...{...}]
	}
	admin: {
		access_log_path: "/dev/stdout"
		address: socket_address: {
			address:    "127.0.0.1"
			port_value: 8001
		}
	}
}
