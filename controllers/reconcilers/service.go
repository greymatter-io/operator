package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (s Service) Build(mesh *v1.Mesh) client.Object {
	configs := gmcore.Base().Overlay(mesh.Spec.Version)
	svc := s.GmService
	svcCfg := configs[svc]

	matchLabels := map[string]string{
		"greymatter.io/control": s.ObjectKey.Name,
	}

	labels := map[string]string{
		"greymatter.io/control":         s.ObjectKey.Name,
		"greymatter.io/component":       svcCfg.Component,
		"greymatter.io/service-version": svcCfg.ImageTag,
	}
	if svc != gmcore.Control && s.ObjectKey.Name != "edge" {
		labels["greymatter.io/sidecar-version"] = configs[gmcore.Proxy].ImageTag
	}

	objectLabels := map[string]string{
		"app.kubernetes.io/name":       s.ObjectKey.Name,
		"app.kubernetes.io/version":    svcCfg.ImageTag,
		"app.kubernetes.io/part-of":    "greymatter",
		"app.kubernetes.io/managed-by": "gm-operator",
		"app.kubernetes.io/created-by": "gm-operator",
	}
	for k, v := range labels {
		objectLabels[k] = v
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ObjectKey.Name,
			Namespace: mesh.Namespace,
			Labels:    objectLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: matchLabels,
			Ports:    svcCfg.ServicePorts,
		},
	}

	if s.ServiceKind != "" {
		service.Spec.Type = s.ServiceKind
	}

	return service
}

func (s Service) Reconciled(mesh *v1.Mesh, obj client.Object) (bool, error) {
	configs := gmcore.Base().Overlay(mesh.Spec.Version)
	svc := s.GmService
	svcCfg := configs[svc]

	labels := obj.GetLabels()
	if lbl := labels["greymatter.io/service-version"]; lbl != svcCfg.ImageTag {
		return false, nil
	}
	if lbl := labels["greymatter.io/sidecar-version"]; svc != gmcore.Control &&
		s.ObjectKey.Name != "edge" &&
		lbl != configs[gmcore.Proxy].ImageTag {
		return false, nil
	}

	return true, nil
}

func (s Service) Mutate(mesh *v1.Mesh, obj client.Object) client.Object {
	return obj
}
