package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
)

var base = map[SvcName]Config{
	Control: {
		Component: "fabric",
		MkEnvsMap: func(mesh *installv1.Mesh) map[string]string {
			return map[string]string{
				"GM_CONTROL_API_INSECURE":             "true",
				"GM_CONTROL_API_SSL":                  "false",
				"GM_CONTROL_API_SSLCERT":              "/etc/proxy/tls/sidecar/server.crt",
				"GM_CONTROL_API_SSLKEY":               "/etc/proxy/tls/sidecar/server.key",
				"GM_CONTROL_CONSOLE_LEVEL":            "info",
				"GM_CONTROL_API_KEY":                  "xxx",
				"GM_CONTROL_API_ZONE_NAME":            "zone-default-zone",
				"GM_CONTROL_API_HOST":                 "control-api:5555",
				"GM_CONTROL_CMD":                      "kubernetes",
				"GM_CONTROL_XDS_RESOLVE_DNS":          "true",
				"GM_CONTROL_XDS_ADS_ENABLED":          "true",
				"GM_CONTROL_KUBERNETES_CLUSTER_LABEL": "greymatter.io",
				"GM_CONTROL_KUBERNETES_PORT_NAME":     "proxy",
				"GM_CONTROL_KUBERNETES_NAMESPACES":    mesh.Namespace,
			}
		},
	},
}
