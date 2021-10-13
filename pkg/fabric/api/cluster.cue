package api

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
