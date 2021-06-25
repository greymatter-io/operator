package gmcore

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
)

var base = configs{
	Control: {
		Component: "fabric",
		Directory: "release",
		Envs: mkEnvOpts(
			func(_ map[string]string, mesh *v1.Mesh, _ string) map[string]string {
				return map[string]string{
					"GM_CONTROL_API_INSECURE":             "true",
					"GM_CONTROL_API_SSL":                  "false",
					"GM_CONTROL_CONSOLE_LEVEL":            "info",
					"GM_CONTROL_API_ZONE_NAME":            mesh.Name,
					"GM_CONTROL_API_HOST":                 "control-api:5555",
					"GM_CONTROL_CMD":                      "kubernetes",
					"GM_CONTROL_KUBERNETES_CLUSTER_LABEL": "greymatter.io/control",
					"GM_CONTROL_KUBERNETES_PORT_NAME":     "proxy",
					"GM_CONTROL_KUBERNETES_NAMESPACES":    mesh.Namespace,
				}
			},
		),
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 50000, Name: "grpc", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 50000, TargetPort: intstr.FromInt(50000), Protocol: "TCP"},
		},
	},
	ControlApi: {
		Component: "fabric",
		Directory: "release",
		Envs: mkEnvOpts(
			func(_ map[string]string, mesh *v1.Mesh, _ string) map[string]string {
				return map[string]string{
					"GM_CONTROL_API_ADDRESS":        "0.0.0.0:5555",
					"GM_CONTROL_API_LOG_LEVEL":      "info",
					"GM_CONTROL_API_PERSISTER_TYPE": "null",
					"GM_CONTROL_API_EXPERIMENTS":    "true",
					"GM_CONTROL_API_BASE_URL":       "/services/control-api/latest/v1.0/",
					"GM_CONTROL_API_USE_TLS":        "false",
					"GM_CONTROL_API_ZONE_KEY":       mesh.Name,
					"GM_CONTROL_API_ZONE_NAME":      mesh.Name,
				}
			},
		),
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 5555, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 5555, TargetPort: intstr.FromInt(5555), Protocol: "TCP"},
		},
	},
	Proxy: {
		Component: "fabric",
		Directory: "release",
		Envs: mkEnvOpts(
			func(_ map[string]string, mesh *v1.Mesh, clusterName string) map[string]string {
				return map[string]string{
					"ENVOY_ADMIN_LOG_PATH": "/dev/stdout",
					"PROXY_DYNAMIC":        "true",
					"XDS_CLUSTER":          clusterName,
					"XDS_HOST":             fmt.Sprintf("control.%s.svc", mesh.Namespace),
					"XDS_PORT":             "50000",
					"XDS_ZONE":             mesh.Name,
				}
			},
		),
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 10808, Name: "proxy", Protocol: "TCP"},
			{ContainerPort: 8081, Name: "metrics", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Name: "proxy", Port: 10808, Protocol: "TCP"},
			{Name: "metrics", Port: 8081, Protocol: "TCP"},
		},
		Resources: &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		},
	},
	Catalog: {
		Component: "sense",
		Directory: "release",
		Envs: mkEnvOpts(
			func(_ map[string]string, mesh *v1.Mesh, _ string) map[string]string {
				return map[string]string{
					"CONTROL_SERVER_0_ADDRESS":              fmt.Sprintf("control.%s.svc:50000", mesh.Namespace),
					"CONTROL_SERVER_0_REQUEST_CLUSTER_NAME": "edge",
					"CONTROL_SERVER_0_ZONE_NAME":            mesh.Name,
					"PORT":                                  "9080",
				}
			},
		),
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 9080, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 9080, TargetPort: intstr.FromInt(9080), Protocol: "TCP"},
		},
	},
	JwtSecurity: {
		Component: "fabric",
		Directory: "release",
		Envs: mkEnvOpts(
			func(map[string]string, *v1.Mesh, string) map[string]string {
				return map[string]string{
					// TODO: add these to secret
					// the secret will need to be retrieved via controller.Get and passed as an arg
					"JWT_API_KEY": "MTIzCg==",
					"PRIVATE_KEY": "LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1JSGNBZ0VCQkVJQkhRY01yVUh5ZEFFelNnOU1vQWxneFF1a3lqQTROL2laa21ETVIvdFRkVmg3U3hNYk8xVE4KeXdzRkJDdTYvZEZXTE5rUDJGd1FFQmtqREpRZU9mc3hKZWlnQndZRks0RUVBQ09oZ1lrRGdZWUFCQUJEWklJeAp6a082cWpkWmF6ZG1xWFg1dnRFcWtodzlkcVREeTN6d0JkcXBRUmljWDRlS2lZUUQyTTJkVFJtWk0yZE9FRHh1Clhja0hzcVMxZDNtWHBpcDh2UUZHTWJCM1hRVm9DZWN0SUlLMkczRUlwWmhGZFNGdG1sa2t5U1N4angzcS9UcloKaVlRTjhJakpPbUNueUdXZ1VWUkdERURiNWlZdkZXc3dpSkljSWYyOGVRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=",
					"HTTP_PORT":   "3000",
					"ENABLE_TLS":  "false",
				}
			},
		),
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 3000, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 3000, TargetPort: intstr.FromInt(3000), Protocol: "TCP"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{MountPath: "/gm-jwt-security/etc", Name: "jwt-users"},
		},
	},
	Dashboard: {
		Component: "sense",
		Directory: "release",
		Envs: mkEnvOpts(
			func(_ map[string]string, mesh *v1.Mesh, _ string) map[string]string {
				return map[string]string{
					"BASE_URL":                     "/services/dashboard/istio/",
					"CONFIG_SERVER":                "/services/control-api/latest/v1.0",
					"DISABLE_PROMETHEUS_ROUTES_UI": "false",
					"ENABLE_INLINE_DOCS":           "true",
					"FABRIC_SERVER":                "/services/catalog/latest/",
					"OBJECTIVES_SERVER":            "/services/slo/latest/",
					"PROMETHEUS_SERVER":            "/services/prometheus/latest/api/v1/",
					"REQUEST_TIMEOUT":              "50000",
					"SERVER_SSL_CA":                "/certs/ca.crt",
					"SERVER_SSL_CERT":              "/certs/server.crt",
					"SERVER_SSL_ENABLED":           "false",
					"SERVER_SSL_KEY":               "/certs/server.key",
					"USE_PROMETHEUS":               "true",
					"AD_SERVER":                    "/services/lad/latest/",
				}
			},
		),
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 1337, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 1337, TargetPort: intstr.FromInt(1337), Protocol: "TCP"},
		},
	},
}
