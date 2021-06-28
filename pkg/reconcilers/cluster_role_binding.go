package reconcilers

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/greymatter.io/operator/pkg/api/v1"
	"github.com/greymatter.io/operator/pkg/gmcore"
)

type ClusterRoleBinding struct {
	Name string
}

func (crb ClusterRoleBinding) Kind() string {
	return "rbacv1.ClusterRoleBinding"
}

func (crb ClusterRoleBinding) Key() types.NamespacedName {
	return types.NamespacedName{Name: crb.Name}
}

func (crb ClusterRoleBinding) Object() client.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: crb.Name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     crb.Name,
		},
	}
}

func (crb ClusterRoleBinding) Reconcile(mesh *v1.Mesh, _ gmcore.Configs, obj client.Object) (client.Object, bool) {
	binding := obj.(*rbacv1.ClusterRoleBinding)

	for _, subject := range binding.Subjects {
		if subject.Name == crb.Name && subject.Namespace == mesh.Namespace {
			return binding, false
		}
	}

	binding.Subjects = append(binding.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      crb.Name,
		Namespace: mesh.Namespace,
	})

	return binding, true
}
