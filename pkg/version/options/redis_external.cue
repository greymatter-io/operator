host: string
port: string
password: string
db: string

control_api: {
	env: {
		"GM_CONTROL_API_REDIS_HOST": "\(host)"
		"GM_CONTROL_API_REDIS_PORT": "\(port)"
		"GM_CONTROL_API_REDIS_PASS": "\(password)"
		"GM_CONTROL_API_REDIS_DB": "\(db)"
	}
}
catalog: {
	env: {
		"REDIS_HOST": "\(host)"
		"REDIS_PORT": "\(port)"
		"REDIS_PASS": "\(password)"
		"REDIS_DB": "\(db)"
	}
}
jwt_security: {
	env: {
		"REDIS_HOST": "\(host)"
		"REDIS_PORT": "\(port)"
		"REDIS_PASS": "\(password)"
		"REDIS_DB": "\(db)"
	}
}
