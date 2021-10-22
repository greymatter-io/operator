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
		`"metrics_key_depth":"3"`,
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
		map[string]int32{"redis": 6379})
	if err != nil {
		t.Fatal(err)
	}

	testContains(service.Listener,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"example-6379"`,
	)(t)
}

func TestServiceNoIngress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if service.Ingresses == nil {
		t.Fatal("Ingresses should not be nil")
	}

	if count := len(service.Ingresses.Clusters); count != 0 {
		t.Errorf("expected len(Ingresses.Clusters) to be 0 but got %d", count)
	}

	if count := len(service.Ingresses.Routes); count != 0 {
		t.Errorf("expected len(Ingresses.Routes) to be 0 but got %d", count)
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

	if service.Ingresses == nil {
		t.Fatal("Ingresses should not be nil")
	}

	if count := len(service.Ingresses.Clusters); count != 1 {
		t.Fatalf("expected len(Ingresses.Clusters) to be 1 but got %d", count)
	}

	t.Run("Cluster", testContains(service.Ingresses.Clusters[0],
		`"cluster_key":"example-5555"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"127.0.0.1","port":5555}]`,
	))

	if count := len(service.Ingresses.Routes); count != 1 {
		t.Fatalf("expected len(Ingresses.Routes) to be 1 but got %d", count)
	}

	t.Run("Route", testContains(service.Ingresses.Routes[0],
		`"route_key":"example-5555"`,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-5555"`,
		`"route_match":{"path":"/"`,
		`"redirects":[]`,
	))
}

func TestServiceMultipleIngresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", nil, map[string]int32{
		"api":  5555,
		"api2": 8080,
	})
	if err != nil {
		t.Fatal(err)
	}

	if service.Ingresses == nil {
		t.Fatal("Ingresses should not be nil")
	}

	if count := len(service.Ingresses.Clusters); count != 2 {
		t.Fatalf("expected len(Ingresses.Clusters) to be 2 but got %d", count)
	}

	if count := len(service.Ingresses.Routes); count != 2 {
		t.Fatalf("expected len(Ingresses.Routes) to be 2 but got %d", count)
	}

	for i, e := range []struct {
		cluster string
		port    int32
	}{
		{"api", 5555},
		{"api2", 8080},
	} {
		key := fmt.Sprintf("example-%d", e.port)
		t.Run("Cluster", testContains(service.Ingresses.Clusters[i],
			fmt.Sprintf(`"cluster_key":"%s"`, key),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"instances":[{"host":"127.0.0.1","port":%d}]`, e.port),
		))
		t.Run("Route", testContains(service.Ingresses.Routes[i],
			fmt.Sprintf(`"route_key":"%s"`, key),
			`"domain_key":"example"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"%s"`, key),
			fmt.Sprintf(`"route_match":{"path":"/%s/"`, e.cluster),
			fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, e.cluster, e.cluster),
		))
	}
}

func TestServiceOneHTTPLocalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/http-local-egress": "othercluster",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-http-egress"]`,
		`"listener_keys":["example","example-http-egress"]`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("HTTPEgresses is nil")
	}

	t.Run("Domain", testContains(service.HTTPEgresses.Domain,
		`"domain_key":"example-http-egress"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", testContains(service.HTTPEgresses.Listener,
		`"listener_key":"example-http-egress"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-http-egress"]`,
		`"port":10909`,
	))
	t.Run("Route", testContains(service.HTTPEgresses.Routes[0],
		`"route_key":"example-to-othercluster"`,
		`"domain_key":"example-http-egress"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"othercluster"`,
		`"route_match":{"path":"/othercluster/"`,
		`"redirects":[{"from":"^/othercluster$","to":"/othercluster/"`,
	))
}

func TestServiceMultipleHTTPLocalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/http-local-egress": "othercluster1,othercluster2",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if service.HTTPEgresses == nil {
		t.Fatal("HTTPEgresses is nil")
	}

	for i, cluster := range []string{"othercluster1", "othercluster2"} {
		t.Run(fmt.Sprintf("Route:%s", cluster), testContains(service.HTTPEgresses.Routes[i],
			fmt.Sprintf(`"route_key":"example-to-%s"`, cluster),
			`"domain_key":"example-http-egress"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"%s"`, cluster),
			fmt.Sprintf(`"route_match":{"path":"/%s/"`, cluster),
			fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, cluster, cluster),
		))
	}
}

func TestServiceOneHTTPExternalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]string{
		"greymatter.io/http-external-egress": "google;google.com:80",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-http-egress"]`,
		`"listener_keys":["example","example-http-egress"]`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("ExternalEgresses is nil")
	}

	t.Run("Domain", testContains(service.HTTPEgresses.Domain,
		`"domain_key":"example-http-egress"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", testContains(service.HTTPEgresses.Listener,
		`"listener_key":"example-http-egress"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-http-egress"]`,
		`"port":10909`,
	))
	t.Run("Cluster", testContains(service.HTTPEgresses.Clusters[0],
		`"cluster_key":"example-to-external-google"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"google.com","port":80}]`,
	))
	t.Run("Route", testContains(service.HTTPEgresses.Routes[0],
		`"route_key":"example-to-external-google"`,
		`"domain_key":"example-http-egress"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-external-google"`,
		`"route_match":{"path":"/google/"`,
		`"redirects":[{"from":"^/google$","to":"/google/"`,
	))
}

func TestServiceMultipleHTTPExternalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]string{
		"greymatter.io/http-external-egress": "google;google.com:80,amazon;amazon.com:80",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if service.HTTPEgresses == nil {
		t.Fatal("ExternalEgresses is nil")
	}

	for i, cluster := range []string{"google", "amazon"} {
		t.Run(fmt.Sprintf("Cluster:%s", cluster), testContains(service.HTTPEgresses.Clusters[i],
			fmt.Sprintf(`"cluster_key":"example-to-external-%s"`, cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"instances":[{"host":"%s.com","port":80}]`, cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", cluster), testContains(service.HTTPEgresses.Routes[i],
			fmt.Sprintf(`"route_key":"example-to-external-%s"`, cluster),
			`"domain_key":"example-http-egress"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-external-%s"`, cluster),
			fmt.Sprintf(`"route_match":{"path":"/%s/"`, cluster),
			fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, cluster, cluster),
		))
	}
}

func TestServiceOneTCPLocalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/tcp-local-egress": "othercluster",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-tcp-egress-to-othercluster"]`,
		`"listener_keys":["example","example-tcp-egress-to-othercluster"]`,
	))

	if count := len(service.TCPEgresses); count != 1 {
		t.Fatalf("Expected 1 TCP egress but got %d", count)
	}

	t.Run("Domain", testContains(service.TCPEgresses[0].Domain,
		`"domain_key":"example-tcp-egress-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"port":10910`,
	))
	t.Run("Listener", testContains(service.TCPEgresses[0].Listener,
		`"listener_key":"example-tcp-egress-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-tcp-egress-to-othercluster"]`,
		`"port":10910`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"othercluster"`,
	))
	t.Run("Route", testContains(service.TCPEgresses[0].Routes[0],
		`"route_key":"example-to-othercluster"`,
		`"domain_key":"example-tcp-egress-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"othercluster"`,
		`"route_match":{"path":"/"`,
	))
}

func TestServiceMultipleTCPLocalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/tcp-local-egress": "c1,c2",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-tcp-egress-to-c1","example-tcp-egress-to-c2"]`,
		`"listener_keys":["example","example-tcp-egress-to-c1","example-tcp-egress-to-c2"]`,
	))

	if count := len(service.TCPEgresses); count != 2 {
		t.Fatalf("Expected 1 TCP egress but got %d", count)
	}

	for i, e := range []struct {
		cluster string
		tcpPort int32
	}{
		{"c1", 10910},
		{"c2", 10911},
	} {
		t.Run(fmt.Sprintf("Domain:%s", e.cluster), testContains(service.TCPEgresses[i].Domain,
			fmt.Sprintf(`"domain_key":"example-tcp-egress-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"port":%d`, e.tcpPort),
		))
		t.Run(fmt.Sprintf("Listener:%s", e.cluster), testContains(service.TCPEgresses[i].Listener,
			fmt.Sprintf(`"listener_key":"example-tcp-egress-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"domain_keys":["example-tcp-egress-to-%s"]`, e.cluster),
			fmt.Sprintf(`"port":%d`, e.tcpPort),
			`"active_network_filters":["envoy.tcp_proxy"]`,
			`"network_filters":{"envoy_tcp_proxy":{`,
			fmt.Sprintf(`"cluster":"%s"`, e.cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", e.cluster), testContains(service.TCPEgresses[i].Routes[0],
			fmt.Sprintf(`"route_key":"example-to-%s"`, e.cluster),
			fmt.Sprintf(`"domain_key":"example-tcp-egress-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"%s"`, e.cluster),
			`"route_match":{"path":"/"`,
		))
	}
}

func TestServiceOneTCPExternalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/tcp-external-egress": "redis;1.2.3.4:6379",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-tcp-egress-to-external-redis"]`,
		`"listener_keys":["example","example-tcp-egress-to-external-redis"]`,
	))

	if count := len(service.TCPEgresses); count != 1 {
		t.Fatalf("Expected 1 TCP egress but got %d", count)
	}

	t.Run("Domain", testContains(service.TCPEgresses[0].Domain,
		`"domain_key":"example-tcp-egress-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"port":10910`,
	))
	t.Run("Listener", testContains(service.TCPEgresses[0].Listener,
		`"listener_key":"example-tcp-egress-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-tcp-egress-to-external-redis"]`,
		`"port":10910`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"example-to-external-redis"`,
	))
	t.Run("Cluster", testContains(service.TCPEgresses[0].Clusters[0],
		`"cluster_key":"example-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"1.2.3.4","port":6379}]`,
	))
	t.Run("Route", testContains(service.TCPEgresses[0].Routes[0],
		`"route_key":"example-to-external-redis"`,
		`"domain_key":"example-tcp-egress-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-external-redis"`,
		`"route_match":{"path":"/"`,
	))
}

func TestServiceMultipleTCPExternalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/tcp-external-egress": "s1;1.1.1.1:1111,s2;2.2.2.2:2222",
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-tcp-egress-to-external-s1","example-tcp-egress-to-external-s2"]`,
		`"listener_keys":["example","example-tcp-egress-to-external-s1","example-tcp-egress-to-external-s2"]`,
	))

	if count := len(service.TCPEgresses); count != 2 {
		t.Fatalf("Expected 1 TCP egress but got %d", count)
	}

	for i, e := range []struct {
		cluster string
		tcpPort int32
	}{
		{"s1", 10910},
		{"s2", 10911},
	} {
		t.Run(fmt.Sprintf("Domain:%s", e.cluster), testContains(service.TCPEgresses[i].Domain,
			fmt.Sprintf(`"domain_key":"example-tcp-egress-to-external-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"port":%d`, e.tcpPort),
		))
		t.Run(fmt.Sprintf("Listener:%s", e.cluster), testContains(service.TCPEgresses[i].Listener,
			fmt.Sprintf(`"listener_key":"example-tcp-egress-to-external-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"domain_keys":["example-tcp-egress-to-external-%s"]`, e.cluster),
			fmt.Sprintf(`"port":%d`, e.tcpPort),
			`"active_network_filters":["envoy.tcp_proxy"]`,
			`"network_filters":{"envoy_tcp_proxy":{`,
			fmt.Sprintf(`"cluster":"example-to-external-%s"`, e.cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", e.cluster), testContains(service.TCPEgresses[i].Routes[0],
			fmt.Sprintf(`"route_key":"example-to-external-%s"`, e.cluster),
			fmt.Sprintf(`"domain_key":"example-tcp-egress-to-external-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-external-%s"`, e.cluster),
			`"route_match":{"path":"/"`,
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
			Zone: "myzone",
		},
	})
}

func testContains(obj json.RawMessage, subs ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
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
func prettyPrint(raws ...json.RawMessage) {
	for _, raw := range raws {
		b := new(bytes.Buffer)
		json.Indent(b, raw, "", "\t")
		fmt.Println(b.String())
	}
}
