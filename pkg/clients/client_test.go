package clients

import (
	"testing"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestNewClient(t *testing.T) {
	// t.Skip() // only works locally in integrated env

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	fabric.Init()

	// conf := `
	// [api]
	// host = "http://localhost:5555/v1.0"
	// [catalog]
	// host = "http://localhost:8181"
	// mesh = "mesh"
	// `

	// conf = base64.StdEncoding.EncodeToString([]byte(conf))

	// cl, err := newMeshClient(&v1alpha1.Mesh{
	// 	ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
	// 	Spec:       v1alpha1.MeshSpec{Zone: "zone"},
	// }, "--base64-config", conf)

	cl, err := newMeshClient(&v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
		Spec:       v1alpha1.MeshSpec{Zone: "zone"},
	},
		"--api.host localhost:5555",
		"--catalog.host localhost:8181",
		"--catalog.mesh mesh",
	)
	if err != nil {
		t.Fatal(err)
	}

	cl.cmds <- cmd{args: "version"}
	time.Sleep(time.Second * 1)

	close(cl.cmds)
}

func TestCLIVersion(t *testing.T) {
	t.Skip() // temp while version in flux

	v, err := cliVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
