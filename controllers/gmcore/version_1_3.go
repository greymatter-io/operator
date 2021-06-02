package gmcore

import (
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var versionOneThree = map[SvcName]Config{
	Control: {
		Component: base[Control].Component,
		ImageTag:  "1.5.3",
		MkEnvsMap: func(mesh *installv1.Mesh, svc SvcName) map[string]string {
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
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 50000, Name: "grpc", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 50000, TargetPort: intstr.FromInt(50000), Protocol: "TCP"},
		},
	},
	ControlApi: {
		Component: "fabric",
		ImageTag:  "1.5.4",
		MkEnvsMap: func(mesh *installv1.Mesh, svc SvcName) map[string]string {
			return map[string]string{
				"GM_CONTROL_API_ADDRESS":               "0.0.0.0:5555",
				"GM_CONTROL_API_DISABLE_VERSION_CHECK": "false",
				"GM_CONTROL_API_LOG_LEVEL":             "debug",
				"GM_CONTROL_API_PERSISTER_TYPE":        "null",
				"GM_CONTROL_API_EXPERIMENTS":           "true",
				"GM_CONTROL_API_BASE_URL":              "/services/control-api/latest/v1.0/",
				"GM_CONTROL_API_USE_TLS":               "false",
				"GM_CONTROL_API_ORG_KEY":               "deciphernow",
				"GM_CONTROL_API_ZONE_KEY":              "zone-default-zone",
				"GM_CONTROL_API_ZONE_NAME":             "zone-default-zone",
			}
		},
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 5555, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 5555, TargetPort: intstr.FromInt(5555), Protocol: "TCP"},
		},
	},
	Proxy: {
		Component: "fabric",
		ImageTag:  "1.5.1",
		MkEnvsMap: func(mesh *installv1.Mesh, svc SvcName) map[string]string {
			return map[string]string{
				"ENVOY_ADMIN_LOG_PATH": "/dev/stdout",
				"PROXY_DYNAMIC":        "true",
				"XDS_CLUSTER":          string(svc),
				"XDS_HOST":             fmt.Sprintf("control.%s.svc", mesh.Namespace),
				"XDS_PORT":             "50000",
				"XDS_ZONE":             "zone-default-zone",
			}
		},
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
		ImageTag:  "1.2.2",
		MkEnvsMap: func(mesh *installv1.Mesh, svc SvcName) map[string]string {
			return map[string]string{
				"CONTROL_SERVER_0_ADDRESS":              fmt.Sprintf("control.%s.svc.cluster.local:50000", mesh.Namespace),
				"CONTROL_SERVER_0_REQUEST_CLUSTER_NAME": "edge",
				"CONTROL_SERVER_0_ZONE_NAME":            "zone-default-zone",
				"PORT":                                  "9080",
			}
		},
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 9080, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 9080, TargetPort: intstr.FromInt(9080), Protocol: "TCP"},
		},
	},
	JwtSecurity: {
		Component: "fabric",
		ImageTag:  "1.2.0",
		MkEnvsMap: func(mesh *installv1.Mesh, svc SvcName) map[string]string {
			return map[string]string{
				"ENABLE_TLS": "false",
				"REDIS_DB":   "0",
				"REDIS_HOST": fmt.Sprintf("jwt-redis.%s.svc", mesh.Namespace),
				"REDIS_PORT": "6379",
				"HTTPS_PORT": "3000",
			}
		},
		ContainerPorts: []corev1.ContainerPort{
			{ContainerPort: 3000, Name: "http", Protocol: "TCP"},
		},
		ServicePorts: []corev1.ServicePort{
			{Port: 3000, TargetPort: intstr.FromInt(3000), Protocol: "TCP"},
		},
	},
}
