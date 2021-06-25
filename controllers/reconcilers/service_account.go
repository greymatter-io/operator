package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
)

type ServiceAccount struct {
	ObjectKey types.NamespacedName
}

func (sa ServiceAccount) Kind() string {
	return "ServiceAccount"
}

func (sa ServiceAccount) Key() types.NamespacedName {
	return sa.ObjectKey
}

func (sa ServiceAccount) Object() client.Object {
	return &corev1.ServiceAccount{}
}

func (sa ServiceAccount) Build(mesh *v1.Mesh) client.Object {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sa.ObjectKey.Name,
			Namespace: sa.ObjectKey.Namespace,
		},
	}
}

func (sa ServiceAccount) Reconciled(mesh *v1.Mesh, obj client.Object) (bool, error) {
	return true, nil
}

func (sa ServiceAccount) Mutate(mesh *v1.Mesh, obj client.Object) client.Object {
	return obj
}
