package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkIngress(ctx context.Context, mesh *installv1.Mesh) error {
	ingress := &extensionsv1beta1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: "edge", Namespace: mesh.Namespace}, ingress)
	if err != nil && errors.IsNotFound(err) {
		ingress = r.mkIngressObject(mesh)
		r.Log.Info("Creating ingress", "Name", "edge", "Namespace", mesh.Namespace)
		err = r.Create(ctx, ingress)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create ingress for %s:edge", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get ingress for %s:edge", mesh.Namespace))
		return err
	}

	return nil
}

func (r *MeshReconciler) mkIngressObject(mesh *installv1.Mesh) *extensionsv1beta1.Ingress {
	ingress := &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "edge",
			Namespace: mesh.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/backend-protocol":   "https",
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"nginx.ingress.kubernetes.io/ssl-passthrough":    "true",
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					Host: "localhost",
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "edge",
										ServicePort: intstr.FromInt(10808),
									},
									Path: "/",
								},
							},
						},
					},
				},
			},
		},
	}

	ctrl.SetControllerReference(mesh, ingress, r.Scheme)
	return ingress
}
