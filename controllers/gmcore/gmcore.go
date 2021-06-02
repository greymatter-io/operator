package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type Service string

const (
	ControlApi  Service = "control-api"
	Control     Service = "control"
	Proxy       Service = "proxy"
	Catalog     Service = "catalog"
	JwtSecurity Service = "jwt-security"
)

type Config struct {
	Component      string
	ImageTag       string
	MkEnvsMap      func(*installv1.Mesh, string) map[string]string
	ContainerPorts []corev1.ContainerPort
	ServicePorts   []corev1.ServicePort
	Resources      *corev1.ResourceRequirements
}

var configs = map[string]map[Service]Config{
	"latest": versionOneThree,
	"1.3":    versionOneThree,
	"1.2":    versionOneTwo,
}

func Configs(gmVersion string) map[Service]Config {
	if cs, ok := configs[gmVersion]; ok {
		return cs
	}
	return configs["latest"]
}
