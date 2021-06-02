package reconcilers

import (
	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterRoleBinding struct {
	Name string
}

func (crb ClusterRoleBinding) Kind() string {
	return "ClusterRoleBinding"
}

func (crb ClusterRoleBinding) Key() types.NamespacedName {
	return types.NamespacedName{Name: crb.Name}
}

func (crb ClusterRoleBinding) Object() client.Object {
	return &rbacv1.ClusterRoleBinding{}
}

func (crb ClusterRoleBinding) Build(mesh *installv1.Mesh) client.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: crb.Name},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      crb.Name,
				Namespace: mesh.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     crb.Name,
		},
	}
}

func (crb ClusterRoleBinding) Reconciled(mesh *installv1.Mesh, obj client.Object) (bool, error) {
	return true, nil
}
