package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/pkg/api/v1"
	"github.com/bcmendoza/gm-operator/pkg/gmcore"
)

type Service struct {
	GmService   gmcore.Service
	ObjectKey   types.NamespacedName
	ServiceKind corev1.ServiceType
}

func (s Service) Kind() string {
	return "corev1.Service"
}

func (s Service) Key() types.NamespacedName {
	return s.ObjectKey
}

func (s Service) Object() client.Object {
	return &corev1.Service{}
}

func (s Service) Reconcile(mesh *v1.Mesh, configs gmcore.Configs, obj client.Object) (client.Object, bool) {
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

	prev := obj.(*corev1.Service)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            s.ObjectKey.Name,
			Namespace:       mesh.Namespace,
			ResourceVersion: prev.ResourceVersion,
			Labels:          objectLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: matchLabels,
			Ports:    svcCfg.ServicePorts,
		},
	}

	if s.ServiceKind != "" {
		service.Spec.Type = s.ServiceKind
	}

	return service, false
}
