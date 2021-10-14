package clients

import (
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// note: only works locally, this is temporary
func TestNewClient(t *testing.T) {
	t.Skip()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	fabric.Init()

	cl, err := newClient(&v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
		Spec:       v1alpha1.MeshSpec{Zone: "zone"},
	},
		"--config", "/tmp",
		"--api.url", "http://localhost:5555/v1.0",
		"--catalog.url", "http://localhost:8181",
	)
	if err != nil {
		t.Fatal(err)
	}

	close(cl.cmds)
}

func TestCLIVersion(t *testing.T) {
	v, err := cliVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
