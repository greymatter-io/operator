Namespace: string
password: string

redis: {
	env: {
		"REDIS_PASSWORD": =~ "^.{16}$" & "\(password)"
	}
}
control_api: {
	env: {
		"GM_CONTROL_API_REDIS_HOST": "mesh-redis.\(Namespace).svc.cluster.local"
		"GM_CONTROL_API_REDIS_PORT": "6379"
		"GM_CONTROL_API_REDIS_PASS": password
		"GM_CONTROL_API_REDIS_DB": "0"
	}
}
catalog: {
	env: {
		"REDIS_HOST": "mesh-redis.\(Namespace).svc.cluster.local"
		"REDIS_PORT": "6379"
		"REDIS_PASS": password
		"REDIS_DB": "0"
	}
}
jwt_security: {
	env: {
		"REDIS_HOST": "mesh-redis.\(Namespace).svc.cluster.local"
		"REDIS_PORT": "6379"
		"REDIS_PASS": password
		"REDIS_DB": "0"
	}
}
