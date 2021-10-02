namespace: string
password: string

redis: {
	envs: {
		"REDIS_PASSWORD": =~ "^.{16}$" & "\(password)"
	}
}
control_api: {
	envs: {
		"GM_CONTROL_API_REDIS_HOST": "mesh-redis.\(namespace).svc.cluster.local"
		"GM_CONTROL_API_REDIS_PORT": "6379"
		"GM_CONTROL_API_REDIS_PASS": password
		"GM_CONTROL_API_REDIS_DB": "0"
	}
}
catalog: {
	envs: {
		"REDIS_HOST": "mesh-redis.\(namespace).svc.cluster.local"
		"REDIS_PORT": "6379"
		"REDIS_PASS": password
		"REDIS_DB": "0"
	}
}
jwt_security: {
	envs: {
		"REDIS_HOST": "mesh-redis.\(namespace).svc.cluster.local"
		"REDIS_PORT": "6379"
		"REDIS_PASS": password
		"REDIS_DB": "0"
	}
}
