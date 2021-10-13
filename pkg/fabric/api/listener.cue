package api

#Listener: {
	name:         string
	listener_key: name
	zone_key:     string
	ip:           "0.0.0.0"
	port:         int32
	protocol:			"http_auto"
	domain_keys:	[...string]
	active_http_filters?: [...string]
	http_filters?: {...}
	active_network_filters?: [...string]
	network_filters?: {...}
	stream_idle_timeout?:    string
	request_timeout?:        string
	drain_timeout?:          string
	delayed_close_timeout?:  string
	use_remote_address?:     true
	tracing_config?:         #TracingConfig
	access_loggers?:         #AccessLoggers

  // common.cue
	secret?:                 #Secret
	http_protocol_options?:  #HTTPProtocolOptions
	http2_protocol_options?: #HTTP2ProtocolOptions
}

#TracingConfig: {
	ingress?: true
	requestHeadersForTags?: [...string]
}

#AccessLoggers: {
	http_connection_loggers?: #Loggers
	http_upstream_loggers?:   #Loggers
}

#Loggers: {
	disabled?: true
	fileLoggers?: [...#FileAccessLog]
	hTTPGRPCAccessLoggers?: [...#HTTPGRPCAccessLog]
}

#FileAccessLog: {
	path?:   string
	format?: string
	JSONFormat?: {
		[string]: string
	}
	typedJSONFormat?: {
		[string]: string
	}
}

#HTTPGRPCAccessLog: {
	commonConfig?: #GRPCCommonConfig
	additionalRequestHeaders?: [...string]
	additionalResponseHeaders?: [...string]
	additionalResponseTrailers?: [...string]
}

#GRPCCommonConfig: {
	logName?:     string
	gRPCService?: #GRPCService
}

#GRPCService: {
	clusterName?: string
}
