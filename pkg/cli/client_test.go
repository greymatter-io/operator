package cli

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

	// cl, err := newClient(&v1alpha1.Mesh{
	// 	ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
	// 	Spec:       v1alpha1.MeshSpec{Zone: "zone"},
	// }, "--base64-config", conf)

	cl := newClient(&v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
		Spec: v1alpha1.MeshSpec{
			Zone:     "zone",
			MeshPort: 10808,
		},
	},
		"--api.host localhost:5555",
		"--catalog.host localhost:8181",
		"--catalog.mesh mesh",
	)

	objects, _ := cl.f.Service("mock", map[string]int32{"api": 5555})
	cl.controlCmds <- mkApply("domain", objects.Domain)
	cl.controlCmds <- mkApply("listener", objects.Listener)
	cl.controlCmds <- mkApply("proxy", objects.Proxy)
	cl.controlCmds <- mkApply("cluster", objects.Cluster)
	cl.controlCmds <- mkApply("route", objects.Route)
	for _, ingress := range objects.Ingresses {
		cl.controlCmds <- mkApply("cluster", ingress.Cluster)
		cl.controlCmds <- mkApply("route", ingress.Route)
	}

	time.Sleep(time.Second * 5)

	close(cl.controlCmds)
	close(cl.catalogCmds)
}

func TestCLIVersion(t *testing.T) {
	v, err := cliversion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
