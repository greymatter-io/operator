package gmcore

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type configs map[Service]config

type Service string

const (
	ControlApi  Service = "control-api"
	Control     Service = "control"
	Proxy       Service = "proxy"
	Catalog     Service = "catalog"
	JwtSecurity Service = "jwt-security"
)

type config struct {
	Component      string
	Directory      string
	ImageTag       string
	Envs           envsOpts
	ContainerPorts []corev1.ContainerPort
	ServicePorts   []corev1.ServicePort
	VolumeMounts   []corev1.VolumeMount
	Resources      *corev1.ResourceRequirements
}

var versions = map[string]configs{
	"latest": versionOneThree,
	"1.3":    versionOneThree,
}

func Base() configs {
	return base
}

func (cs configs) Overlay(gmVersion string) configs {
	overlays, ok := versions[gmVersion]
	if !ok {
		overlays = versions["latest"]
	}

	for svc, cfg := range cs {
		if overlay, ok := overlays[svc]; ok {
			if overlay.Component != "" {
				cfg.Component = overlay.Component
			}
			if overlay.ImageTag != "" {
				cfg.ImageTag = overlay.ImageTag
			}
			if cfg.Envs != nil && overlay.Envs != nil {
				cfg.Envs = append(cfg.Envs, overlay.Envs...)
			}
			if overlay.ContainerPorts != nil {
				cfg.ContainerPorts = overlay.ContainerPorts
			}
			if overlay.ServicePorts != nil {
				cfg.ServicePorts = overlay.ServicePorts
			}
			if overlay.Resources != nil {
				cfg.Resources = overlay.Resources
			}
			cs[svc] = cfg
		}
	}

	return cs
}

type envsOpts []envsOpt

type envsOpt func(map[string]string, *installv1.Mesh, string) map[string]string

func mkEnvOpts(opt envsOpt) envsOpts {
	return envsOpts{opt}
}

func (eb envsOpts) Configure(mesh *installv1.Mesh, clusterName string) []corev1.EnvVar {
	envsMap := make(map[string]string)
	for _, fn := range eb {
		envsMap = fn(envsMap, mesh, clusterName)
	}
	var envs []corev1.EnvVar
	for k, v := range envsMap {
		envs = append(envs, corev1.EnvVar{Name: k, Value: v})
	}
	return envs
}
