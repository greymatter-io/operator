package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
)

var versionOneThree = configs{
	Control: {
		ImageTag: "1.5.3",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	ControlApi: {
		ImageTag: "1.5.4",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	Proxy: {
		ImageTag: "1.5.1",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	Catalog: {
		ImageTag: "1.2.2",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
	JwtSecurity: {
		ImageTag: "1.2.0",
		Envs: mkEnvOpts(
			func(_ map[string]string, _ *installv1.Mesh, _ string) map[string]string {
				return map[string]string{
					// todo: Add overlays here
				}
			},
		),
	},
}
