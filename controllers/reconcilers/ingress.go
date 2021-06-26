package reconcilers

import (
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
)

type Ingress struct {
	ObjectKey types.NamespacedName
}

func (i Ingress) Kind() string {
	return "Ingress"
}

func (i Ingress) Key() types.NamespacedName {
	return i.ObjectKey
}

func (i Ingress) Object() client.Object {
	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.ObjectKey.Name,
			Namespace: i.ObjectKey.Namespace,
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
}

func (i Ingress) Reconcile(mesh *v1.Mesh, _ gmcore.Configs, obj client.Object) client.Object {
	return obj
}
