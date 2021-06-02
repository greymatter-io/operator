package gmcore

import (
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type SvcName string

const (
	ControlApi  SvcName = "control-api"
	Control     SvcName = "control"
	Proxy       SvcName = "proxy"
	Catalog     SvcName = "catalog"
	JwtSecurity SvcName = "jwt-security"
)

var serviceNames = map[string]SvcName{
	"control-api":  ControlApi,
	"control":      Control,
	"proxy":        Proxy,
	"catalog":      Catalog,
	"jwt-security": JwtSecurity,
}

func ServiceName(s string) (SvcName, error) {
	if svcName, ok := serviceNames[s]; ok {
		return svcName, nil
	}
	return "", fmt.Errorf("%s is not a valid Grey Matter core service name", s)
}

type Config struct {
	Component      string
	ImageTag       string
	MkEnvsMap      func(*installv1.Mesh, SvcName) map[string]string
	ContainerPorts []corev1.ContainerPort
	ServicePorts   []corev1.ServicePort
	Resources      *corev1.ResourceRequirements
}

var configs = map[string]map[SvcName]Config{
	"latest": versionOneThree,
	"1.3":    versionOneThree,
	"1.2":    versionOneTwo,
}

func Configs(gmVersion string) map[SvcName]Config {
	if cs, ok := configs[gmVersion]; ok {
		return cs
	}
	return configs["latest"]
}
