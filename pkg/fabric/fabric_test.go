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
		`"port":10808`,
	)(t)
}

func TestService(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Domain", testContains(service.Domain,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
		`"port":10808`,
	))
	t.Run("Listener", testContains(service.Listener,
		`"listener_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example"]`,
		`"port":10808`,
		`"active_http_filters":["gm.metrics"]`,
		`"http_filters":{"gm_metrics":{`,
	))
	t.Run("Proxy", testContains(service.Proxy,
		`"name":"example"`,
		`"proxy_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example"]`,
		`"listener_keys":["example"]`,
	))
	t.Run("Cluster", testContains(service.Clusters[0],
		`"name":"example"`,
		`"cluster_key":"example"`,
		`"zone_key":"myzone"`,
	))
	t.Run("Route", testContains(service.Routes[0],
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
		`"name":"example"`,
	))
}

func TestTCPProxyFilter(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{"greymatter.io/network-filters": "envoy.tcp_proxy"},
		map[string]int32{"tcpport": 6379})
	if err != nil {
		t.Fatal(err)
	}

	testContains(service.Listener,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"example-tcpport",`,
	)(t)
}

func TestServiceNoIngress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Ingresses); count != 0 {
		t.Fatalf("expected 0 ingress, got %d", count)
	}
}

func TestServiceOneIngress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", nil, map[string]int32{
		"api": 5555,
	})
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Ingresses); count != 1 {
		t.Fatalf("expected 1 ingress, got %d", count)
	}

	ingress, ok := service.Ingresses["example-api"]
	if !ok {
		t.Fatal("did not find example-api in ingresses")
	}

	t.Run("Cluster", testContains(ingress.Clusters[0],
		`"cluster_key":"example-api"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"127.0.0.1","port":5555}]`,
	))
	t.Run("Route", testContains(ingress.Routes[0],
		`"route_key":"example-api"`,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-api"`,
		`"route_match":{"path":"/"`,
		`"redirects":[]`,
	))
}

func TestServiceMultipleIngresses(t *testing.T) {
	f := loadMock(t)

	ports := map[string]int32{
		"api":  5555,
		"api2": 8080,
	}

	service, err := f.Service("example", nil, ports)
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Ingresses); count != 2 {
		t.Fatalf("expected 2 ingresses, got %d", count)
	}

	for name, port := range ports {
		key := fmt.Sprintf("example-%s", name)

		t.Run(key, func(t *testing.T) {
			ingress, ok := service.Ingresses[key]
			if !ok {
				t.Fatalf("did not find %s in ingresses", key)
			}

			t.Run("Cluster", testContains(ingress.Clusters[0],
				fmt.Sprintf(`"cluster_key":"%s"`, key),
				`"zone_key":"myzone"`,
				fmt.Sprintf(`"instances":[{"host":"127.0.0.1","port":%d}]`, port),
			))
			t.Run("Route", testContains(ingress.Routes[0],
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

func TestServiceOneLocalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/local-egress": "othercluster",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if service.HTTPLocalEgresses == nil {
		t.Fatal("LocalEgresses is nil")
	}

	t.Run("Domain", testContains(service.HTTPLocalEgresses.Domain,
		`"domain_key":"example-http-local-egress"`,
		`"zone_key":"myzone"`,
		`"port":10818`,
	))
	t.Run("Listener", testContains(service.HTTPLocalEgresses.Listener,
		`"listener_key":"example-http-local-egress"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-http-local-egress"]`,
		`"port":10818`,
	))
	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-http-local-egress"]`,
		`"listener_keys":["example","example-http-local-egress"]`,
	))
	t.Run("Route", testContains(service.HTTPLocalEgresses.Routes[0],
		`"route_key":"example-http-local-egress-to-othercluster"`,
		`"domain_key":"example-http-local-egress"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"othercluster"`,
		`"route_match":{"path":"/othercluster/"`,
		`"redirects":[{"from":"^/othercluster$","to":"/othercluster/"`,
	))
}

func TestServiceMultipleLocalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/local-egress": "othercluster1,othercluster2",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if service.HTTPLocalEgresses == nil {
		t.Fatal("LocalEgresses is nil")
	}

	for i, cluster := range []string{"othercluster1", "othercluster2"} {
		t.Run(fmt.Sprintf("Route:%s", cluster), testContains(service.HTTPLocalEgresses.Routes[i],
			fmt.Sprintf(`"route_key":"example-http-local-egress-to-%s"`, cluster),
			`"domain_key":"example-http-local-egress"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"%s"`, cluster),
			fmt.Sprintf(`"route_match":{"path":"/%s/"`, cluster),
			fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, cluster, cluster),
		))
	}
}

// func TestServiceOneLocalEgressTCP(t *testing.T) {
// 	f := loadMock(t)

// 	service, err := f.Service("example",
// 		map[string]string{
// 			"greymatter.io/local-egress": "tcp=othercluster",
// 		}, nil,
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if service.HTTPLocalEgresses == nil {
// 		t.Fatal("LocalEgresses is nil")
// 	}

// 	t.Run("Domain", testContains(service.HTTPLocalEgresses.Domain,
// 		`"domain_key":"example-tcp-local-egress"`,
// 		`"zone_key":"myzone"`,
// 		`"port":10818`,
// 	))
// 	t.Run("Listener", testContains(service.HTTPLocalEgresses.Listener,
// 		`"listener_key":"example-http-local-egress"`,
// 		`"zone_key":"myzone"`,
// 		`"domain_keys":["example-http-local-egress"]`,
// 		`"port":10818`,
// 	))
// 	t.Run("Proxy", testContains(service.Proxy,
// 		`"domain_keys":["example","example-http-local-egress"]`,
// 		`"listener_keys":["example","example-http-local-egress"]`,
// 	))
// 	t.Run("Route", testContains(service.HTTPLocalEgresses.Routes[0],
// 		`"route_key":"example-http-local-egress-to-othercluster"`,
// 		`"domain_key":"example-http-local-egress"`,
// 		`"zone_key":"myzone"`,
// 		`"cluster_key":"othercluster"`,
// 		`"route_match":{"path":"/othercluster/"`,
// 		`"redirects":[{"from":"^/othercluster$","to":"/othercluster/"`,
// 	))
// }

func TestServiceOneExternalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]string{
		"greymatter.io/external-egress": "google;www.google.com:80",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if service.HTTPExternalEgresses == nil {
		t.Fatal("ExternalEgresses is nil")
	}

	t.Run("Cluster", testContains(service.HTTPExternalEgresses.Clusters[0],
		`"cluster_key":"example-to-google"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"www.google.com","port":80}]`,
	))
	t.Run("Domain", testContains(service.HTTPExternalEgresses.Domain,
		`"domain_key":"example-http-external-egress"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", testContains(service.HTTPExternalEgresses.Listener,
		`"listener_key":"example-http-external-egress"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-http-external-egress"]`,
		`"port":10909`,
	))
	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-http-external-egress"]`,
		`"listener_keys":["example","example-http-external-egress"]`,
	))
	t.Run("Route", testContains(service.HTTPExternalEgresses.Routes[0],
		`"route_key":"example-http-external-egress-to-google"`,
		`"domain_key":"example-http-external-egress"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-google"`,
		`"route_match":{"path":"/google/"`,
		`"redirects":[{"from":"^/google$","to":"/google/"`,
	))
}

func TestServiceMultipleExternalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]string{
		"greymatter.io/external-egress": "google;www.google.com:80,amazon;www.amazon.com:80",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if service.HTTPExternalEgresses == nil {
		t.Fatal("ExternalEgresses is nil")
	}

	for i, cluster := range []string{"google", "amazon"} {
		t.Run(fmt.Sprintf("Cluster:%s", cluster), testContains(service.HTTPExternalEgresses.Clusters[i],
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"instances":[{"host":"www.%s.com","port":80}]`, cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", cluster), testContains(service.HTTPExternalEgresses.Routes[i],
			fmt.Sprintf(`"route_key":"example-http-external-egress-to-%s"`, cluster),
			`"domain_key":"example-http-external-egress"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, cluster),
			fmt.Sprintf(`"route_match":{"path":"/%s/"`, cluster),
			fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, cluster, cluster),
		))
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
			MeshPort: 10808,
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
