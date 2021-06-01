package common

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type GmCore string

const (
	ControlApi  GmCore = "control-api"
	Control     GmCore = "control"
	Proxy       GmCore = "proxy"
	Catalog     GmCore = "catalog"
	JwtSecurity GmCore = "jwt-security"
)

const ImagePullPolicy = corev1.PullIfNotPresent

type GmCoreConfig struct {
	component      string
	imageTag       string
	mkEnvsMap      func(*installv1.Mesh) map[string]string
	containerPorts []corev1.ContainerPort
	servicePorts   []corev1.ServicePort
	resources      *corev1.ResourceRequirements
}

var GmCoreConfigs = map[string]map[GmCore]GmCoreConfig{
	"1.3": versionOneThree,
	"1.2": versionOneTwo,
}
