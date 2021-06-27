package gmcore

import v1 "github.com/bcmendoza/gm-operator/pkg/api/v1"

var versionOneSix = Configs{
	Control: {
		ImageTag: "1.6.0",
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
		ImageTag: "1.6.0",
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
		ImageTag:  "1.6.0",
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
		ImageTag:  "2.0.0",
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
		ImageTag:  "1.3.0",
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
		Directory: "development",
		ImageTag:  "5.0.0",
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
