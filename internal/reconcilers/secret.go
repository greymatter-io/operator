package reconcilers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/internal/gmcore"
)

type Secret struct {
	ObjectKey     types.NamespacedName
	ObjectLiteral *corev1.Secret
}

func (s Secret) Kind() string {
	return "Secret"
}

func (s Secret) Key() types.NamespacedName {
	return s.ObjectKey
}

func (s Secret) Object() client.Object {
	return s.ObjectLiteral
}

func (s Secret) Reconcile(mesh *v1.Mesh, _ gmcore.Configs, obj client.Object) client.Object {
	return obj
}
