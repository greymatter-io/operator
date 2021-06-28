package gmcore

import v1 "github.com/greymatter.io/operator/pkg/api/v1"

var versionOneThree = Configs{
	Control: {
		Image: "docker.greymatter.io/release/gm-control:1.5.3",
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
		Image: "docker.greymatter.io/release/gm-control-api:1.5.4",
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
		Image: "docker.greymatter.io/release/gm-proxy:1.5.1",
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
		Image: "docker.greymatter.io/release/gm-catalog:1.2.2",
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
		Image: "docker.greymatter.io/release/gm-jwt-security:1.2.0",
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
		Image: "docker.greymatter.io/release/gm-dashboard:4.0.2",
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
		ImageTag: "1.1.5",
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
