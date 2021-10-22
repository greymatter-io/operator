// catalogservice

#CatalogService: {
  mesh_id: string
  service_id: string
	name: *ServiceName | string
	api_endpoint?: string
	description?: string
}

// Cluster

#Cluster: {
	name:        string
	cluster_key: name
	zone_key:    string
	require_tls?: true
	instances?: [...#Instance]
	health_checks?: [...#HealthCheck]
	outlier_detection?: #OutlierDetection
	circuit_breakers?:  #CircuitBreakersThresholds
	ring_hash_lb_conf?:      #RingHashLbConfig
	original_dst_lb_conf?:   #OriginalDstLbConfig
	least_request_lb_conf?:  #LeastRequestLbConfig
	common_lb_conf?:         #CommonLbConfig

	// common.cue
	secret?:     #Secret
	ssl_config?: #SSLConfig
	http_protocol_options?:  #HTTPProtocolOptions
	http2_protocol_options?: #HTTP2ProtocolOptions
}

#Instance: {
	host: string
	port: int32
	metadata?: [...#Metadata]
}

#HealthCheck: {
	timeoutMsec?:               int64
	intervalMsec?:              int64
	intervalJitterMsec?:        int64
	unhealthyThreshold?:        int64
	healthyThreshold?:          int64
	reuseConnection?:           true
	noTrafficIntervalMsec?:     int64
	unhealthyIntervalMsec?:     int64
	unhealthyEdgeIntervalMsec?: int64
	healthyEdgeIntervalMsec?:   int64
	healthChecker?:             #HealthChecker
}

#HealthChecker: {
	HTTPHealthCheck?: #HTTPHealthCheck
	TCPHealthCheck?:  #TCPHealthCheck
}

#TCPHealthCheck: {
	send?: string
	receive?: [...string]
}

#HTTPHealthCheck: {
	host?:        string
	path?:        string
	serviceName?: string
	request_headers_to_add?: [...#Metadata]
}

#OutlierDetection: {
	intervalMsec?:                       int64
	baseEjectionTimeMsec?:               int64
	maxEjectionPercent?:                 int64
	consecutive5xx?:                     int64
	enforcingConsecutive5xx?:            int64
	enforcingSuccessRate?:               int64
	successRateMinimumHosts?:            int64
	successRateRequestVolume?:           int64
	successRateStdevFactor?:             int64
	consecutiveGatewayFailure?:          int64
	enforcingConsecutiveGatewayFailure?: int64
}

#CircuitBreakersThresholds: #CircuitBreakers & {
	high?: #CircuitBreakers
}

#CircuitBreakers: {
	maxConnections?:     int64
	maxPendingRequests?: int64
	maxRequests?:        int64
	maxRetries?:         int64
	maxConnectionPools?: int64
	trackRemaining?:     true
}

#RingHashLbConfig: {
	minimumRingSize?: uint64
	hashFunc?:        uint32
	maximumRingSize?: uint64
}

#OriginalDstLbConfig: {
	useHTTPHeader?: true
}

#LeastRequestLbConfig: {
	choiceCount?: uint32
}

#CommonLbConfig: {
	healthyPanicThreshold?:           #Percent
	zoneAwareLbConf?:                 #ZoneAwareLbConfig
	localityWeightedLbConf?:          #LocalityWeightedLbConfig
	consistentHashingLbConf?:         #ConsistentHashingLbConfig
	updateMergeWindow?:               #Duration
	ignoreNewHostsUntilFirstHc?:      true
	closeConnectionsOnHostSetChange?: true
}

#Percent: {
	value?: float64
}

#ZoneAwareLbConfig: {
	routingEnabled?:     #Percent
	minClusterSize?:     uint64
	failTrafficOnPanic?: true
}

#LocalityWeightedLbConfig: {
}

#ConsistentHashingLbConfig: {
	useHostnameForHashing?: true
}

#Duration: {
	seconds?: int64
	nanos?:   int32
}

// Route

#Route: {
	route_key:  string
	domain_key: string
	zone_key:   string
	prefix_rewrite?:    string | *""
	cohort_seed?:       string | *""
	high_priority?:    true
	timeout?:      string
	idle_timeout?: string
	rules:             [...#Rule]
	route_match:      #RouteMatch
	response_data?:     #ResponseData | *null
	retry_policy?:      #RetryPolicy | *null
	filter_metadata?: {
		[string]: #Metadata
	}
	filter_configs?: {
		[string]: {...}
	}
	request_headers_to_add?: [...#Metadatum]
	response_headers_to_add?: [...#Metadatum]
	request_headers_to_remove?: [...string]
	response_headers_to_remove?: [...string]

  // common.cue
	redirects: [...#Redirect]
}

#RouteMatch: {
	path:      string
	match_type: string
}

#Rule: {
	ruleKey?: string
	methods?: [...string]
	matches?: [...#Match]
	constraints?: #Constraints
	cohort_seed?: string
}

#Match: {
	kind?:     string
	behavior?: string
	from?:     #Metadatum
	to?:       #Metadatum
}

#Constraints: {
	light: [...#Constraint]
	dark?:  [...#Constraint]
	tap?:   [...#Constraint]
}

#Constraint: {
	cluster_key:    string
	metadata?:       [...#Metadata]
	properties?:     [...#Metadata]
	response_data?:  #ResponseData
	weight: uint32
}

#ResponseData: {
	headers?: [...#HeaderDatum]
	cookies?: [...#CookieDatum]
}

#HeaderDatum: {
	responseDatum?: #ResponseDatum
}

#ResponseDatum: {
	name?:           string
	value?:          string
	valueIsLiteral?: true
}

#CookieDatum: {
	responseDatum?: #ResponseDatum
	expiresInSec?:  uint32
	domain?:        string
	path?:          string
	secure?:        true
	httpOnly?:      true
	sameSite?:      string
}

#RetryPolicy: {
	numRetries?:                    int64
	perTryTimeoutMsec?:             int64
	timeoutMsec?:                   int64
	retryOn?:                       string
	retryPriority?:                 string
	retryHostPredicate?:            string
	hostSelectionRetryMaxAttempts?: int64
	retriableStatusCodes?:          int64
	retryBackOff?:                  #BackOff
	retriableHeaders?:              #HeaderMatcher
	retriableRequestHeaders?:       #HeaderMatcher
}

#BackOff: {
	baseInterval?: string
	maxInterval?:  string
}

#HeaderMatcher: {
	name?:           string
	exactMatch?:     string
	regexMatch?:     string
	safeRegexMatch?: #RegexMatcher
	rangeMatch?:     #RangeMatch
	presentMatch?:   true
	prefixMatch?:    string
	suffixMatch?:    string
	invertMatch?:    true
}

#RegexMatcher: {
	googleRE2?: #GoogleRe2
	regex?:     string
}

#GoogleRe2: {
	maxProgramSize?: int64
}

#RangeMatch: {
	start?: int64
	end?:   int64
}

// Domain

#Domain: {
	domain_key:  string
	zone_key:    string
	name:        string | *"*"
	port:        int32
	force_https?: true
	cors_config?: #CorsConfig
	aliases?:     [...string]

  // common.cue
	ssl_config?: #SSLConfig
	redirects?:   [...#Redirect]
	custom_headers?: [...#Metadatum]
}

#CorsConfig: {
	allowedOrigins?: [...#AllowOriginStringMatchItem]
	allowCredentials?: true
	exposedHeaders?: [...string]
	maxAge?: int64
	allowedMethods?: [...string]
	allowedHeaders?: [...string]
}

#AllowOriginStringMatchItem: {
	matchType?: string
	value?:     string
}

// Listener

#Listener: {
	name:         string
	listener_key: name
	zone_key:     string
	ip:           "0.0.0.0"
	port:         int32
	protocol:			"http_auto"
	domain_keys:	[...string]
	active_http_filters?: [...string]
	http_filters?: #HTTPFilters
	active_network_filters?: [...string]
	network_filters?: #NetworkFilters
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

#HTTPFilters: {
	gm_metrics: {
		metrics_host: string
		metrics_port: int
    metrics_dashboard_uri_path: string
    metrics_prometheus_uri_path: string
    metrics_ring_buffer_size: int
    prometheus_system_metrics_interval_seconds: int
    metrics_key_function: string
		metrics_key_depth: string
	}
}

#NetworkFilters: {
	envoy_tcp_proxy?: {
		cluster: string
		stat_prefix: string
	}
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

// Proxy

#Proxy: {
	name:      string
	proxy_key: name
	zone_key:  string
	domain_keys: [...string]
	listener_keys: [...string]
	upgrades?: string
}
