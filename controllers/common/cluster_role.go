package common

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterRoleReconciler struct {
	Name string
}

func (cr ClusterRoleReconciler) Key() types.NamespacedName {
	return types.NamespacedName{Name: cr.Name}
}

func (cr ClusterRoleReconciler) Object() client.Object {
	return &rbacv1.ClusterRole{}
}

func (cr ClusterRoleReconciler) Build(mesh *installv1.Mesh) (client.Object, error) {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: cr.Name},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"list"},
			},
		},
	}, nil
}

func (cr ClusterRoleReconciler) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	return true, nil
}
