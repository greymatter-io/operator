package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
)

var versionOneThree = configs{
	Control: {
		ImageTag: "1.5.3",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				overlays := map[string]string{
					// todo: Add overlays here
				}
				for k, v := range overlays {
					envs[k] = v
				}
				return envs
			},
		),
	},
	ControlApi: {
		ImageTag: "1.5.4",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				overlays := map[string]string{
					// todo: Add overlays here
				}
				for k, v := range overlays {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Proxy: {
		ImageTag: "1.5.1",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				overlays := map[string]string{
					// todo: Add overlays here
				}
				for k, v := range overlays {
					envs[k] = v
				}
				return envs
			},
		),
	},
	Catalog: {
		ImageTag: "1.2.2",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				overlays := map[string]string{
					// todo: Add overlays here
				}
				for k, v := range overlays {
					envs[k] = v
				}
				return envs
			},
		),
	},
	JwtSecurity: {
		ImageTag: "1.2.0",
		Envs: mkEnvOpts(
			func(envs map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				overlays := map[string]string{
					// todo: Add overlays here
				}
				for k, v := range overlays {
					envs[k] = v
				}
				return envs
			},
		),
	},
}
