package installer

import (
	"testing"

	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestInjectPKI(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	manifests, err := getSpireManifests()
	if err != nil {
		t.Fatal(err)
	}

	secret := manifests.Secret

	cs, err := cfsslsrv.New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := cs.Start(); err != nil {
		t.Fatal(err)
	}

	secret, err = injectPKI(secret, cs)
	if err != nil {
		t.Fatal(err)
	}

	for _, k := range []string{"root.crt", "intermediate.crt", "intermediate.key"} {
		_, ok := secret.StringData[k]
		if !ok {
			t.Errorf("did not find %s", k)
		}
	}
}

func TestGetSpireManifests(t *testing.T) {
	manifests, err := getSpireManifests()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		got  interface{}
		want interface{}
	}{
		{manifests.Namespace.Name, "spire"},
		{manifests.Secret.Name, "server-ca"},
		{manifests.Service.Name, "server"},
		{manifests.StatefulSet.Name, "server"},
		{manifests.StatefulSet.Spec.VolumeClaimTemplates[0].Name, "server-data"},
		{manifests.DaemonSet.Name, "agent"},
		{manifests.Role.Name, "server"},
		{manifests.RoleBinding.Name, "server"},
		{manifests.RoleBinding.RoleRef.Name, "server"},
		{manifests.RoleBinding.Subjects[0].Name, "server"},
		{manifests.ServiceAccounts[0].Name, "server"},
		{manifests.ServiceAccounts[1].Name, "agent"},
		{manifests.ClusterRoles[0].Name, "spire-server"},
		{manifests.ClusterRoles[1].Name, "spire-agent"},
		{manifests.ClusterRoleBindings[0].Name, "spire-server"},
		{manifests.ClusterRoleBindings[0].RoleRef.Name, "spire-server"},
		{manifests.ClusterRoleBindings[0].Subjects[0].Name, "server"},
		{manifests.ClusterRoleBindings[1].Name, "spire-agent"},
		{manifests.ClusterRoleBindings[1].RoleRef.Name, "spire-agent"},
		{manifests.ClusterRoleBindings[1].Subjects[0].Name, "agent"},
		{manifests.ConfigMaps[0].Name, "server-config"},
		{manifests.ConfigMaps[1].Name, "server-bundle"},
		{manifests.ConfigMaps[2].Name, "agent-config"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %s; want %s", tc.got, tc.want)
		}
	}
}
