package common

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceReconciler struct{}

func (sr ServiceReconciler) Object() client.Object {
	return &corev1.Service{}
}

func (sr ServiceReconciler) Build(mesh *installv1.Mesh, svc gmcore.SvcName) (client.Object, error) {
	configs := gmcore.Configs(mesh.Spec.Version)

	labels := map[string]string{
		"greymatter.io/component":       configs[svc].Component,
		"greymatter.io/service-version": configs[svc].ImageTag,
		"greymatter.io/control":         string(svc),
	}
	if svc != gmcore.Control {
		labels["greymatter.io/sidecar-version"] = configs[gmcore.Proxy].ImageTag
	}

	objectLabels := map[string]string{
		"app.kubernetes.io/name":       string(svc),
		"app.kubernetes.io/version":    configs[svc].ImageTag,
		"app.kubernetes.io/part-of":    "greymatter",
		"app.kubernetes.io/managed-by": "gm-operator",
		"app.kubernetes.io/created-by": "gm-operator",
	}
	for k, v := range labels {
		objectLabels[k] = v
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(svc),
			Namespace: mesh.Namespace,
			Labels:    objectLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: objectLabels,
			Ports:    configs[svc].ServicePorts,
		},
	}

	return service, nil
}

func (sr ServiceReconciler) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	configs := gmcore.Configs(mesh.Spec.Version)

	svc, err := gmcore.ServiceName(obj.GetName())
	if err != nil {
		return false, err
	}

	labels := obj.GetLabels()
	if lbl := labels["greymatter.io/service-version"]; lbl != configs[svc].ImageTag {
		return false, nil
	}
	if lbl := labels["greymatter.io/sidecar-version"]; svc != gmcore.Control && lbl != configs[gmcore.Proxy].ImageTag {
		return false, nil
	}

	return true, nil
}
