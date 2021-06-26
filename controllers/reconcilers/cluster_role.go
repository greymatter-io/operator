package reconcilers

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
)

type ClusterRole struct {
	Name  string
	Rules []rbacv1.PolicyRule
}

func (cr ClusterRole) Kind() string {
	return "ClusterRole"
}

func (cr ClusterRole) Key() types.NamespacedName {
	return types.NamespacedName{Name: cr.Name}
}

func (cr ClusterRole) Object() client.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name,
		},
		Rules: cr.Rules,
	}
}

func (cr ClusterRole) Reconcile(mesh *v1.Mesh, _ gmcore.Configs, obj client.Object) client.Object {
	return obj
}
