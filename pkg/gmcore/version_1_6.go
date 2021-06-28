package gmcore

import v1 "github.com/greymatter.io/operator/pkg/api/v1"

var versionOneSix = Configs{
	Control: {
		Image: "docker.greymatter.io/release/gm-control:1.6.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					// todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	ControlApi: {
		Image: "docker.greymatter.io/release/gm-control-api:1.6.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					// todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Proxy: {
		Image: "docker.greymatter.io/development/gm-proxy:1.6.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					// todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Catalog: {
		Image: "docker.greymatter.io/development/gm-catalog:2.0.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					// todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	JwtSecurity: {
		Image: "docker.greymatter.io/development/gm-jwt-security:1.3.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					// todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Dashboard: {
		Image: "docker.greymatter.io/development/gm-dashboard:5.0.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					// todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Slo: {
		Image: "docker.greymatter.io/release/gm-slo:1.2.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					//todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Postgres: {
		Image: "docker.io/centos/postgresql-10-centos7",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *v1.Mesh, _ string) map[string]string {
				patches := map[string]string{
					//todo: Add patches here
				}
				for k, v := range patches {
					envs[k] = v
				}
				return envs
			},
		),
	},
}
