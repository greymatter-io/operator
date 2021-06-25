package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
)

var versionOneSixBeta = configs{
	Control: {
		Directory: "development",
		ImageTag:  "1.6.0-dev",
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
		Directory: "development",
		ImageTag:  "1.6.0-dev",
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
		Directory: "development",
		ImageTag:  "1.6.1-dev",
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
		Directory: "development",
		ImageTag:  "latest",
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
		Directory: "development",
		ImageTag:  "latest",
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
