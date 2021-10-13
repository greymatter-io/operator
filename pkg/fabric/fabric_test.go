package fabric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEdge(t *testing.T) {
	f := loadMock(t)
	edge := f.Edge()

	t.Run("Proxy", testObject(edge.Proxy,
		`"proxy_key":"edge"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Domain", testObject(edge.Domain,
		`"domain_key":"edge"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Listener", testObject(edge.Listener,
		`"listener_key":"edge"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Cluster", testObject(edge.Cluster,
		`"cluster_key":"edge"`,
		`"zone_key":"myzone"`,
	))
}

func TestService(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]int32{
		"api": 5555,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testObject(service.Proxy,
		`"proxy_key":"example"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Domain", testObject(service.Domain,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Listener", testObject(service.Listener,
		`"listener_key":"example"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Cluster", testObject(service.Cluster,
		`"cluster_key":"example"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Route", testObject(service.Route,
		`"route_key":"example"`,
		`"zone_key":"myzone"`,
	))
}

func TestServiceOnePort(t *testing.T) {
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
		t.Run("Cluster", testObject(ingress.Cluster,
			`"cluster_key":"example-5555"`,
			`"zone_key":"myzone"`,
		))
		t.Run("Route", testObject(ingress.Route,
			`"route_key":"example-5555"`,
			`"zone_key":"myzone"`,
			`"route_match":{"path":"/"`,
		))
	})
}

func TestServiceMultiplePorts(t *testing.T) {
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

			t.Run("Cluster", testObject(ingress.Cluster,
				fmt.Sprintf(`"cluster_key":"%s"`, key),
				`"zone_key":"myzone"`,
			))
			t.Run("Route", testObject(ingress.Route,
				fmt.Sprintf(`"route_key":"%s"`, key),
				`"zone_key":"myzone"`,
				fmt.Sprintf(`"route_match":{"path":"/%s/"`, name),
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

	f, err := New("myzone", 10909)
	if err != nil {
		t.Fatal(err)
	}

	return f
}

func testObject(obj json.RawMessage, subs ...string) func(t *testing.T) {
	return func(t *testing.T) {
		for _, sub := range subs {
			if !bytes.Contains(obj, json.RawMessage(sub)) {
				t.Fatalf("did not contain substring '%s'", sub)
			}
		}
	}
}

//lint:ignore U1000 print util
func prettyPrint(raw json.RawMessage) {
	b := new(bytes.Buffer)
	json.Indent(b, raw, "", "\t")
	fmt.Println(b.String())
}
