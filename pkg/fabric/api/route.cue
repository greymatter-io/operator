package api

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
		[string]: #Any
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

#Any: {
	typeUrl?: string
	value?:   bytes
}
