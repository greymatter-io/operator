#Metadata: {
	metadata?: [...#Metadatum]
}

#Metadatum: {
	key:   string
	value: string
}

#Secret: {
	secret_key?:  string
	secret_name?: string
	secret_validation_name?:          string
	subject_names?:                   [...string]
	ecdh_curves?:                     [...string]
	forward_client_cert_details?:     string
	set_current_client_cert_details?: #SetCurrentClientCertDetails
}

#SetCurrentClientCertDetails: {
	uri?: true
}

#SSLConfig: {
	cipherFilter?: string
	protocols?: [...string]
	certKeyPairs?: [...#CertKeyPathPair]
	requireClientCerts?: true
	trustFile?:          string
	SNI?: [...string]
	CRL?: #DataSource
}

#CertKeyPathPair: {
	certificatePath?: string
	keyPath?:         string
}

#DataSource: {
	filename?:     string
	inlineString?: string
}

#HTTPProtocolOptions: {
	allowAbsoluteURL?:     true
	acceptHTTP10?:         true
	defaultHostForHTTP10?: string
	headerKeyFormat?:      #HeaderKeyFormat
	enableTrailers?:       true
}

#HeaderKeyFormat: {
	properCaseWords?: true
}

#HTTP2ProtocolOptions: {
	hpackTableSize?:                               uint32
	maxConcurrentStreams?:                         uint32
	initialStreamWindowSize?:                      uint32
	initialConnectionWindowSize?:                  uint32
	allowConnect?:                                 true
	maxOutboundFrames?:                            uint32
	maxOutboundControlFrames?:                     uint32
	maxConsecutiveInboundFramesWithEmptyPayload?:  uint32
	maxInboundPriorityFramesPerStream?:            uint32
	maxInboundWindowUpdateFramesPerDataFrameSent?: uint32
	streamErrorOnInvalidHTTPMessaging?:            true
}

#Redirect: {
	name?:         string
	from?:         string
	to?:           string
	redirect_type?: string
	header_constraints?: [...#HeaderConstraint]
}

#HeaderConstraint: {
	name?:          string
	value?:         string
	caseSensitive?: true
	invert?:        true
}
