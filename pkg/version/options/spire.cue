proxy: {
	volumes: {
		"spire-socket": {
			hostPath: {
				path: "/run/spire/socket"
				type: "DirectoryOrCreate"
			}
		}
	}
	volumeMounts: {
		"spire-socket": {
			mountPath: "/run/spire/socket"
		}
	}
	env: {
		"SPIRE_PATH": "/run/spire/socket/agent.sock"
	}
}
