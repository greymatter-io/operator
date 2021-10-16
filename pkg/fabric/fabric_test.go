package fabric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cueutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEdgeDomain(t *testing.T) {
	f := loadMock(t)

	testContains(f.EdgeDomain(),
		`"domain_key":"edge"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	)(t)
}

func TestService(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]int32{
		"api": 5555,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Domain", testContains(service.Domain,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", testContains(service.Listener,
		`"listener_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example"]`,
		`"port":10909`,
	))
	t.Run("Proxy", testContains(service.Proxy,
		`"name":"example"`,
		`"proxy_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example"]`,
		`"listener_keys":["example"]`,
	))
	t.Run("Cluster", testContains(service.Cluster,
		`"name":"example"`,
		`"cluster_key":"example"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Route", testContains(service.Route,
		`"route_key":"example"`,
		`"domain_key":"edge"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example"`,
		`"route_match":{"path":"/services/example/"`,
		`"redirects":[{"from":"^/services/example$","to":"/services/example/"`,
	))
	t.Run("CatalogService", testContains(service.CatalogService,
		`"mesh_id":"mymesh"`,
		`"service_id":"example"`,
	))
}

func TestServiceNoIngress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", nil)
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Ingresses); count != 0 {
		t.Fatalf("expected 0 ingress, got %d", count)
	}
}

func TestServiceOneIngress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]int32{
		"api": 5555,
	})
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Ingresses); count != 1 {
		t.Fatalf("expected 1 ingress, got %d", count)
	}

	ingress, ok := service.Ingresses["example-5555"]
	if !ok {
		t.Fatal("did not find example-5555 in ingresses")
	}

	t.Run("example-5555", func(t *testing.T) {
		t.Run("Cluster", testContains(ingress.Cluster,
			`"cluster_key":"example-5555"`,
			`"zone_key":"myzone"`,
			`"instances":[{"host":"127.0.0.1","port":5555}]`,
		))
		t.Run("Route", testContains(ingress.Route,
			`"route_key":"example-5555"`,
			`"domain_key":"example"`,
			`"zone_key":"myzone"`,
			`"cluster_key":"example-5555"`,
			`"route_match":{"path":"/"`,
			`"redirects":[]`,
		))
	})
}

func TestServiceMultipleIngresses(t *testing.T) {
	f := loadMock(t)

	ports := map[string]int32{
		"api":  5555,
		"api2": 8080,
	}

	service, err := f.Service("example", ports)
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Ingresses); count != 2 {
		t.Fatalf("expected 2 ingresses, got %d", count)
	}

	for name, port := range ports {
		key := fmt.Sprintf("example-%d", port)

		t.Run(key, func(t *testing.T) {
			ingress, ok := service.Ingresses[key]
			if !ok {
				t.Fatalf("did not find %s in ingresses", key)
			}

			t.Run("Cluster", testContains(ingress.Cluster,
				fmt.Sprintf(`"cluster_key":"%s"`, key),
				`"zone_key":"myzone"`,
				fmt.Sprintf(`"instances":[{"host":"127.0.0.1","port":%d}]`, port),
			))
			t.Run("Route", testContains(ingress.Route,
				fmt.Sprintf(`"route_key":"%s"`, key),
				`"domain_key":"example"`,
				`"zone_key":"myzone"`,
				fmt.Sprintf(`"cluster_key":"%s"`, key),
				fmt.Sprintf(`"route_match":{"path":"/%s/"`, name),
				fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, name, name),
			))
		})
	}
}

func loadMock(t *testing.T) *Fabric {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	if err := Init(); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	return New(&v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mymesh"},
		Spec: v1alpha1.MeshSpec{
			Zone:     "myzone",
			MeshPort: 10909,
		},
	})
}

func testContains(obj json.RawMessage, subs ...string) func(t *testing.T) {
	return func(t *testing.T) {
		for _, sub := range subs {
			if !bytes.Contains(obj, json.RawMessage(sub)) {
				t.Errorf("did not contain substring '%s'", sub)
			}
		}
		if t.Failed() {
			prettyPrint(obj)
		}
	}
}

//lint:ignore U1000 print util
func prettyPrint(raw json.RawMessage) {
	b := new(bytes.Buffer)
	json.Indent(b, raw, "", "\t")
	fmt.Println(b.String())
}
