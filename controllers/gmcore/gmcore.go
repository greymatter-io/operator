package gmcore

import (
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
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
	"latest":   versionOneThree,
	"1.6-beta": versionOneSixBeta,
	"1.3":      versionOneThree,
}

func Base() configs {
	return base
}

func (cs configs) Patch(gmVersion string) configs {
	patches, ok := versions[gmVersion]
	if !ok {
		patches = versions["latest"]
	}

	for svc, cfg := range cs {
		if patch, ok := patches[svc]; ok {
			if patch.Component != "" {
				cfg.Component = patch.Component
			}
			if patch.ImageTag != "" {
				cfg.ImageTag = patch.ImageTag
			}
			if cfg.Envs != nil && patch.Envs != nil {
				cfg.Envs = append(cfg.Envs, patch.Envs...)
			}
			if patch.ContainerPorts != nil {
				cfg.ContainerPorts = patch.ContainerPorts
			}
			if patch.ServicePorts != nil {
				cfg.ServicePorts = patch.ServicePorts
			}
			if patch.Resources != nil {
				cfg.Resources = patch.Resources
			}
			if patch.VolumeMounts != nil {
				cfg.VolumeMounts = patch.VolumeMounts
			}
			cs[svc] = cfg
		}
	}

	return cs
}

type envsOpts []envsOpt

type envsOpt func(map[string]string, *v1.Mesh, string) map[string]string

func mkEnvOpts(opt envsOpt) envsOpts {
	return envsOpts{opt}
}

func (eb envsOpts) Configure(mesh *v1.Mesh, clusterName string) []corev1.EnvVar {
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
