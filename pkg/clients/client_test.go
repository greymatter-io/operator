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

	// mc, err := newMeshClient(&v1alpha1.Mesh{
	// 	ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
	// 	Spec:       v1alpha1.MeshSpec{Zone: "zone"},
	// }, "--base64-config", conf)

	mc := newMeshClient(&v1alpha1.Mesh{
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

	objects, _ := mc.f.Service("mock", map[string]int32{"api": 5555})
	mc.controlCmds <- mkApply("domain", objects.Domain)
	mc.controlCmds <- mkApply("listener", objects.Listener)
	mc.controlCmds <- mkApply("proxy", objects.Proxy)
	mc.controlCmds <- mkApply("cluster", objects.Cluster)
	mc.controlCmds <- mkApply("route", objects.Route)
	for _, ingress := range objects.Ingresses {
		mc.controlCmds <- mkApply("cluster", ingress.Cluster)
		mc.controlCmds <- mkApply("route", ingress.Route)
	}

	time.Sleep(time.Second * 8)

	close(mc.controlCmds)
	close(mc.catalogCmds)
}

func TestCLIVersion(t *testing.T) {
	v, err := cliVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
