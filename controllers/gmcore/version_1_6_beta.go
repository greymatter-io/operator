package gmcore

import v1 "github.com/bcmendoza/gm-operator/api/v1"

var versionOneSixBeta = configs{
	Control: {
		Directory: "development",
		ImageTag:  "1.6.0-dev",
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
		Directory: "development",
		ImageTag:  "1.6.0-dev",
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
		Directory: "development",
		ImageTag:  "1.6.1-dev",
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
		Directory: "development",
		ImageTag:  "latest",
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
		Directory: "development",
		ImageTag:  "latest",
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
}
