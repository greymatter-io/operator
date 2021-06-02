package reconcilers

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAccount struct {
	ObjectKey types.NamespacedName
}

func (sa ServiceAccount) Key() types.NamespacedName {
	return sa.ObjectKey
}

func (sa ServiceAccount) Object() client.Object {
	return &corev1.ServiceAccount{}
}

func (sa ServiceAccount) Build(mesh *installv1.Mesh) (client.Object, error) {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sa.ObjectKey.Name,
			Namespace: sa.ObjectKey.Namespace,
		},
	}, nil
}

func (sa ServiceAccount) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	return true, nil
}
