package api

#Proxy: {
	name:      string
	proxy_key: name
	zone_key:  string
	domain_keys: [...string]
	listener_keys: [...string]
	upgrades?: string
	listeners?: [...#Listener]
}
