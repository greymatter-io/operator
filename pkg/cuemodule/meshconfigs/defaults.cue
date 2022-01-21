package meshconfigs

import "greymatter.io/operator/greymatter-cue/greymatter"

// Identify core service versions for each Grey Matter release
_ServiceVersions: {
	if mesh.spec.release_version == "1.7" {
		edge:        "1.7.0"
		control:     "1.7.0"
		catalog:     "3.0.0"
		dashboard:   "6.0.0"
		jwtsecurity: "1.3.0"
	}
	if mesh.spec.release_version == "1.6" {
		edge:        "1.6.3"
		control:     "1.6.5"
		catalog:     "2.0.1"
		dashboard:   "5.1.1"
		jwtsecurity: "1.3.0"
	}
}

#Domain: greymatter.#Domain & {
	zone_key: mesh.spec.zone
	name:     "*"
}

#Proxy: greymatter.#Proxy & {
	zone_key:  mesh.spec.zone
	proxy_key: workload.metadata.name
	name:      proxy_key
}

#Listener: greymatter.#Listener & {
	listener_key: string
	name:         listener_key
	zone_key:     mesh.spec.zone
	ip:           "0.0.0.0"
	protocol:     "http_auto"
}

#Cluster: greymatter.#Cluster & {
	cluster_key: string
	name:        *cluster_key | string
	zone_key:    mesh.spec.zone
	require_tls: true
}

#Route: greymatter.#Route & {
	zone_key: mesh.spec.zone
}

edgeDomain: #Domain

service: {
	proxy:    #Proxy
	domain:   #Domain & {domain_key:     workload.metadata.name}
	listener: #Listener & {listener_key: workload.metadata.name}
	clusters: [...#Cluster]
	routes: [...#Route]

	ingresses: {
		clusters: [...#Cluster]
		routes: [...#Route]
	}

	if len(HTTPEgresses) > 0 {
		httpEgresses: {
			domain:   #Domain
			listener: #Listener
			clusters: [...#Cluster]
			routes: [...#Route]
		}
	}

	tcpEgresses: [...{
		domain:   #Domain
		listener: #Listener
		clusters: [...#Cluster]
		routes: [...#Route]
	}]
}
