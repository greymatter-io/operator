package gmcore

import v1 "github.com/bcmendoza/gm-operator/api/v1"

var versionOneThree = configs{
	Control: {
		ImageTag: "1.5.3",
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
		ImageTag: "1.5.4",
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
		ImageTag: "1.5.1",
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
		ImageTag: "1.2.2",
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
		ImageTag: "1.2.0",
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
