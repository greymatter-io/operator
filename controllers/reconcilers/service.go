package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
)

type Service struct {
	GmService   gmcore.Service
	ObjectKey   types.NamespacedName
	ServiceKind corev1.ServiceType
}

func (s Service) Kind() string {
	return "Service"
}

func (s Service) Key() types.NamespacedName {
	return s.ObjectKey
}

func (s Service) Object() client.Object {
	return &corev1.Service{}
}

func (s Service) Reconcile(mesh *v1.Mesh, configs gmcore.Configs, obj client.Object) client.Object {
	svc := s.GmService
	svcCfg := configs[svc]

	matchLabels := map[string]string{
		"greymatter.io/control": s.ObjectKey.Name,
	}

	labels := map[string]string{
		"greymatter.io/control":   s.ObjectKey.Name,
		"greymatter.io/component": svcCfg.Component,
	}

	objectLabels := map[string]string{
		"app.kubernetes.io/name":       s.ObjectKey.Name,
		"app.kubernetes.io/part-of":    "greymatter",
		"app.kubernetes.io/managed-by": "gm-operator",
		"app.kubernetes.io/created-by": "gm-operator",
	}
	for k, v := range labels {
		objectLabels[k] = v
	}

	service := obj.(*corev1.Service)

	service.ObjectMeta.Name = s.ObjectKey.Name
	service.ObjectMeta.Namespace = mesh.Namespace
	service.ObjectMeta.Labels = mesh.Labels

	service.Spec.Selector = matchLabels
	service.Spec.Ports = svcCfg.ServicePorts

	if s.ServiceKind != "" {
		service.Spec.Type = s.ServiceKind
	}

	return service
}
