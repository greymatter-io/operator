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
	envs: {
		"SPIRE_PATH": "/run/spire/socket/agent.sock"
	}
}
