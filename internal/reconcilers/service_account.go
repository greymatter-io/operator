package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/internal/gmcore"
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
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sa.ObjectKey.Name,
			Namespace: sa.ObjectKey.Namespace,
		},
	}
}

func (sa ServiceAccount) Reconcile(mesh *v1.Mesh, _ gmcore.Configs, obj client.Object) client.Object {
	return obj
}
