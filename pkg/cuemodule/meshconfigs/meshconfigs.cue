package meshconfigs

edgeDomain: {
	domain_key: "edge"
	port:       10808
}

service: {
	if workload.metadata.name != "gm-redis" {
		catalogservice: #CatalogService & {
			if workload.metadata.name == "edge" {
				name:            "Grey Matter Edge"
				version:         _ServiceVersions.edge
				description:     "Handles north/south traffic flowing through the mesh."
				api_endpoint:    "/"
				business_impact: "critical"
			}
			if workload.metadata.name == "control" {
				name:              "Grey Matter Control"
				version:           _ServiceVersions.control
				description:       "Manages the configuration of the Grey Matter data plane."
				api_endpoint:      "/services/control/api/v1.0/"
				api_spec_endpoint: "/services/control/api/"
				business_impact:   "critical"
			}
			if workload.metadata.name == "catalog" {
				name:              "Grey Matter Catalog"
				version:           _ServiceVersions.catalog
				description:       "Interfaces with the control plane to expose the current state of the mesh."
				api_endpoint:      "/services/catalog/"
				api_spec_endpoint: "/services/catalog/"
				business_impact:   "high"
			}
			if workload.metadata.name == "dashboard" {
				name:            "Grey Matter Dashboard"
				version:         _ServiceVersions.dashboard
				description:     "A user dashboard that paints a high-level picture of the mesh."
				business_impact: "high"
			}
			if workload.metadata.name == "jwt-security" {
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
			workload.metadata.name,
			if len(HTTPEgresses) > 0 {
				"\(workload.metadata.name)-egress-http"
			},
			for k, _ in TCPEgresses {
				"\(workload.metadata.name)-egress-tcp-to-\(k)"
			},
		]
		listener_keys: [
			workload.metadata.name,
			if len(HTTPEgresses) > 0 {
				"\(workload.metadata.name)-egress-http"
			},
			for k, _ in TCPEgresses {
				"\(workload.metadata.name)-egress-tcp-to-\(k)"
			},
		]
	}

	domain: {
		domain_key: workload.metadata.name
		port:       10808
	}

	listener: {
		listener_key: workload.metadata.name
		port:         domain.port
		domain_keys: [workload.metadata.name]

		if workload.metadata.name != "gm-redis" {
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
					if workload.metadata.name == "edge" {
						metrics_key_depth: "1"
					}
					if workload.metadata.name != "edge" {
						metrics_key_depth: "3"
					}
					if mesh.spec.release_version != "1.6" {
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
					for v in Ingresses {
						_key:        "\(workload.metadata.name):\(v)"
						cluster:     _key
						stat_prefix: _key
					}
				}
			}
		}
	}

	clusters: [
		{
			cluster_key: workload.metadata.name
		},
	]

	routes: [
		if workload.metadata.name != "dashboard" && workload.metadata.name != "edge" {
			{
				route_key:  workload.metadata.name
				domain_key: "edge"
				route_match: {
					path:       "/services/\(workload.metadata.name)/"
					match_type: "prefix"
				}
				redirects: [
					{
						from:          "^/services/\(workload.metadata.name)$"
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
									cluster_key: workload.metadata.name
									weight:      1
								},
							]
						}
					},
				]
			}
		},
		if workload.metadata.name == "dashboard" {
			{
				route_key:  workload.metadata.name
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
									cluster_key: workload.metadata.name
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
			for v in Ingresses if len(Ingresses) > 0 && workload.metadata.name != "edge" {
				{
					cluster_key: "\(workload.metadata.name):\(v)"
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
			for k, v in Ingresses if len(Ingresses) > 0 && workload.metadata.name != "edge" {
				if len(Ingresses) == 1 {
					{
						_rk:        "\(workload.metadata.name):\(v)"
						route_key:  _rk
						domain_key: workload.metadata.name
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
						_rk:        "\(workload.metadata.name):\(v)"
						route_key:  _rk
						domain_key: workload.metadata.name
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
			_dk: "\(workload.metadata.name)-egress-http"
			domain: {
				domain_key: _dk
				port:       10909
			}
			listener: {
				listener_key: _dk
				port:         10909
				domain_keys: [_dk]
			}
			clusters: [
				for k, v in HTTPEgresses {
					if v.isExternal {
						{
							cluster_key: "\(workload.metadata.name)-to-\(k)"
							instances: [
								{
									host: v.host
									port: v.port
								},
							]
						}
					}
					if !v.isExternal {
						{
							cluster_key: "\(workload.metadata.name)-to-\(k)"
							name:        k
						}
					}
				},
			]
			routes: [
				for k, v in HTTPEgresses {
					{
						_rk:        "\(workload.metadata.name)-to-\(k)"
						route_key:  _rk
						domain_key: _dk
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
				},
			]
		}
	}

	tcpEgresses: [
		for k, v in TCPEgresses {
			_dk: "\(workload.metadata.name)-egress-tcp-to-\(k)"
			{
				domain: {
					domain_key: _dk
					port:       v.tcpPort
				}
				listener: {
					listener_key: _dk
					domain_keys: [_dk]
					port: v.tcpPort
					active_network_filters: [
						"envoy.tcp_proxy",
					]
					network_filters: {
						envoy_tcp_proxy: {
							cluster:     k
							stat_prefix: k
						}
					}
				}
				clusters: [
					{
						cluster_key: "\(workload.metadata.name)-to-\(k)"
						name:        k
						if v.isExternal {
							instances: [
								{
									host: v.host
									port: v.port
								},
							]
						}
					},
				]
				routes: [
					{
						_rk:        "\(workload.metadata.name)-to-\(k)"
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
