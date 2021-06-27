package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/pkg/gmcore"
)

type ConfigMap struct {
	ObjectKey types.NamespacedName
	Data      map[string]string
}

func (cm ConfigMap) Kind() string {
	return "corev1.ConfigMap"
}

func (cm ConfigMap) Key() types.NamespacedName {
	return cm.ObjectKey
}

func (cm ConfigMap) Object() client.Object {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.ObjectKey.Name,
			Namespace: cm.ObjectKey.Namespace,
		},
		Data: cm.Data,
	}
}

// todo: allow for updating jwt-users
func (cm ConfigMap) Reconcile(_ *v1.Mesh, _ gmcore.Configs, obj client.Object) (client.Object, bool) {
	return obj, false
}
