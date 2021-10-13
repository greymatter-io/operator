package api

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
