package reconcilers

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMap struct {
	ObjectKey types.NamespacedName
	Data      map[string]string
}

func (cm ConfigMap) Kind() string {
	return "ConfigMap"
}

func (cm ConfigMap) Key() types.NamespacedName {
	return cm.ObjectKey
}

func (cm ConfigMap) Object() client.Object {
	return &corev1.ConfigMap{}
}

func (cm ConfigMap) Build(mesh *installv1.Mesh) client.Object {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.ObjectKey.Name,
			Namespace: mesh.Namespace,
		},
		Data: cm.Data,
	}
}

func (cm ConfigMap) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	return true, nil
}

func (cm ConfigMap) Mutate(mesh *installv1.Mesh, obj client.Object) client.Object {
	return obj
}
