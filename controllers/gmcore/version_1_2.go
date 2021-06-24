package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
)

var versionOneTwo = configs{
	Control: {
		ImageTag: "1.4.2",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	ControlApi: {
		ImageTag: "1.4.1",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	Proxy: {
		ImageTag: "1.4.0",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	Catalog: {
		ImageTag: "1.0.7",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	JwtSecurity: {
		ImageTag: "1.1.1",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
}
