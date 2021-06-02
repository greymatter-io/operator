package common

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAccountReconciler struct {
	ObjectKey types.NamespacedName
}

func (sar ServiceAccountReconciler) Key() types.NamespacedName {
	return sar.ObjectKey
}

func (sar ServiceAccountReconciler) Object() client.Object {
	return &corev1.ServiceAccount{}
}

func (sar ServiceAccountReconciler) Build(mesh *installv1.Mesh) (client.Object, error) {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sar.ObjectKey.Name,
			Namespace: sar.ObjectKey.Namespace,
		},
	}, nil
}

func (sar ServiceAccountReconciler) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	return true, nil
}
