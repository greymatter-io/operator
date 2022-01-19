package meshconfigs

edgeDomain: {
	domain_key: "edge"
	port:       10808
}

service: {
	if ServiceName != "gm-redis" {
		catalogservice: #CatalogService & {
			if ServiceName == "edge" {
				name:            "Grey Matter Edge"
				version:         _ServiceVersions.edge
				description:     "Handles north/south traffic flowing through the mesh."
				api_endpoint:    "/"
				business_impact: "critical"
			}
			if ServiceName == "control" {
				name:              "Grey Matter Control"
				version:           _ServiceVersions.control
				description:       "Manages the configuration of the Grey Matter data plane."
				api_endpoint:      "/services/control/api/v1.0/"
				api_spec_endpoint: "/services/control/api/"
				business_impact:   "critical"
			}
			if ServiceName == "catalog" {
				name:              "Grey Matter Catalog"
				version:           _ServiceVersions.catalog
				description:       "Interfaces with the control plane to expose the current state of the mesh."
				api_endpoint:      "/services/catalog/"
				api_spec_endpoint: "/services/catalog/"
				business_impact:   "high"
			}
			if ServiceName == "dashboard" {
				name:            "Grey Matter Dashboard"
				version:         _ServiceVersions.dashboard
				description:     "A user dashboard that paints a high-level picture of the mesh."
				business_impact: "high"
			}
			if ServiceName == "jwt-security" {
				name:              "Grey Matter JWT Security"
				version:           _ServiceVersions.jwtsecurity
				description:       "A JWT token generation and retrieval service."
				api_endpoint:      "/services/jwt-security/"
				api_spec_endpoint: "/services/jwt-security/"
				business_impact:   "high"
			}
		}
	}

	proxy: {
		domain_keys: [
			ServiceName,
			if len(HTTPEgresses) > 0 {
				"\(ServiceName)-egress-http"
			},
			for _, e in TCPEgresses {
				"\(ServiceName)-egress-tcp-to-\(e.cluster)"
			},
		]
		listener_keys: [
			ServiceName,
			if len(HTTPEgresses) > 0 {
				"\(ServiceName)-egress-http"
			},
			for _, e in TCPEgresses {
				"\(ServiceName)-egress-tcp-to-\(e.cluster)"
			},
		]
	}

	domain: {
		domain_key: ServiceName
		port:       10808
	}

	listener: {
		listener_key: ServiceName
		port: domain.port
		domain_keys: [ServiceName]

		if ServiceName != "gm-redis" {
			active_http_filters: [
				"gm.metrics",
			]
			http_filters: {
				gm_metrics: {
					metrics_host:                               "0.0.0.0"
					metrics_port:                               8081
					metrics_dashboard_uri_path:                 "/metrics"
					metrics_prometheus_uri_path:                "prometheus"
					metrics_ring_buffer_size:                   4096
					prometheus_system_metrics_interval_seconds: 15
					metrics_key_function:                       "depth"
					if ServiceName == "edge" {
						metrics_key_depth: "1"
					}
					if ServiceName != "edge" {
						metrics_key_depth: "3"
					}
					if ReleaseVersion != "1.6" {
						metrics_receiver: {
							redis_connection_string: "redis://127.0.0.1:10910"
							push_interval_seconds:   10
						}
					}
				}
			}
		}

		if NetworkFilters["envoy.tcp_proxy"] && len(Ingresses) == 1 {
			active_network_filters: [
				"envoy.tcp_proxy",
			]
			network_filters: {
				envoy_tcp_proxy: {
					for _, v in Ingresses {
						_key:        "\(ServiceName):\(v)"
						cluster:     _key
						stat_prefix: _key
					}
				}
			}
		}
	}

	clusters: [
		{
			cluster_key: ServiceName
		},
	]

	routes: [
		if ServiceName != "dashboard" && ServiceName != "edge" {
			{
				route_key:  ServiceName
				domain_key: "edge"
				route_match: {
					path:       "/services/\(ServiceName)/"
					match_type: "prefix"
				}
				redirects: [
					{
						from:          "^/services/\(ServiceName)$"
						to:            route_match.path
						redirect_type: "permanent"
					},
				]
				prefix_rewrite: "/"
				rules: [
					{
						constraints: {
							light: [
								{
									cluster_key: ServiceName
									weight:      1
								},
							]
						}
					},
				]
			}
		},
		if ServiceName == "dashboard" {
			{
				route_key:  ServiceName
				domain_key: "edge"
				route_match: {
					path:       "/"
					match_type: "prefix"
				}
				rules: [
					{
						constraints: {
							light: [
								{
									cluster_key: ServiceName
									weight:      1
								},
							]
						}
					},
				]
			}
		},
	]

	ingresses: {
		clusters: [
			for _, v in Ingresses if len(Ingresses) > 0 && ServiceName != "edge" {
				{
					cluster_key: "\(ServiceName):\(v)"
					instances: [
						{
							host: "127.0.0.1"
							port: v
						},
					]
				}
			},
		]
		routes: [
			for k, v in Ingresses if len(Ingresses) > 0 && ServiceName != "edge" {
				if len(Ingresses) == 1 {
					{
						_rk:        "\(ServiceName):\(v)"
						route_key:  _rk
						domain_key: ServiceName
						route_match: {
							path:       "/"
							match_type: "prefix"
						}
						rules: [
							{
								constraints: {
									light: [
										{
											cluster_key: _rk
											weight:      1
										},
									]
								}
							},
						]
					}
				}
				if len(Ingresses) > 1 {
					{
						_rk:        "\(ServiceName):\(v)"
						route_key:  _rk
						domain_key: ServiceName
						route_match: {
							path:       "/\(k)/"
							match_type: "prefix"
						}
						redirects: [
							{
								from:          "^/\(k)$"
								to:            route_match.path
								redirect_type: "permanent"
							},
						]
						prefix_rewrite: "/"
						rules: [
							{
								constraints: {
									light: [
										{
											cluster_key: _rk
											weight:      1
										},
									]
								}
							},
						]
					}
				}
			},
		]
	}

	if len(HTTPEgresses) > 0 {
		httpEgresses: {
			_dk: "\(ServiceName)-egress-http"
				domain: {
					domain_key: _dk
					port:       10909
				}
				listener: {
					listener_key: _dk
					port: 10909
					domain_keys: [_dk]
				}
				clusters: [
					for _, e in HTTPEgresses {
						if e.isExternal {
							{
								cluster_key: "\(ServiceName)-to-\(e.cluster)"
								instances: [
									{
										host: e.host
										port: e.port
									},
								]
							}
						}
						if !e.isExternal {
							{
								cluster_key: "\(ServiceName)-to-\(e.cluster)"
								name:        e.cluster
							}
						}
					},
				]
				routes: [
					for _, e in HTTPEgresses {
						{
							_rk:        "\(ServiceName)-to-\(e.cluster)"
							route_key:  _rk
							domain_key: _dk
							route_match: {
								path:       "/\(e.cluster)/"
								match_type: "prefix"
							}
							redirects: [
								{
									from:          "^/\(e.cluster)$"
									to:            route_match.path
									redirect_type: "permanent"
								},
							]
							prefix_rewrite: "/"
							rules: [
								{
									constraints: {
										light: [
											{
												cluster_key: _rk
												weight:      1
											},
										]
									}
								},
							]
						}
					},
				]
			}
	}

	tcpEgresses: [
		for _, e in TCPEgresses {
			_dk: "\(ServiceName)-egress-tcp-to-\(e.cluster)"
			{
				domain: {
					domain_key: _dk
					port:       e.tcpPort
				}
				listener: {
					listener_key: _dk
					domain_keys: [_dk]
					port: e.tcpPort
					active_network_filters: [
						"envoy.tcp_proxy",
					]
					network_filters: {
						envoy_tcp_proxy: {
							cluster:     e.cluster
							stat_prefix: e.cluster
						}
					}
				}
				clusters: [
					{
						cluster_key: "\(ServiceName)-to-\(e.cluster)"
						name:        e.cluster
						if e.isExternal {
							instances: [
								{
									host: e.host
									port: e.port
								},
							]
						}
					},
				]
				routes: [
					{
						_rk:        "\(ServiceName)-to-\(e.cluster)"
						route_key:  _rk
						domain_key: _dk
						route_match: {
							path:       "/"
							match_type: "prefix"
						}
						rules: [
							{
								constraints: {
									light: [
										{
											cluster_key: _rk
											weight:      1
										},
									]
								}
							},
						]
					},
				]
			}
		},
	]
}
