package fabric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
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
		`"redis_connection_string":"redis://:`,
	))
	t.Run("Proxy", testContains(service.Proxy,
		`"name":"example"`,
		`"proxy_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis"]`,
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

	// 1 TCP egress is expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 1 {
		t.Fatalf("Expected 1 TCP egress but got %d", count)
	}

	t.Run("Domain", testContains(service.TCPEgresses[0].Domain,
		`"domain_key":"example-egress-tcp-to-gm-redis"`,
		`"zone_key":"myzone"`,
		`"port":10910`,
	))
	t.Run("Listener", testContains(service.TCPEgresses[0].Listener,
		`"listener_key":"example-egress-tcp-to-gm-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-tcp-to-gm-redis"]`,
		`"port":10910`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"gm-redis"`,
	))
	t.Run("Route", testContains(service.TCPEgresses[0].Routes[0],
		`"route_key":"example-to-gm-redis"`,
		`"domain_key":"example-egress-tcp-to-gm-redis"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"gm-redis"`,
		`"route_match":{"path":"/"`,
	))

	t.Run("CatalogService", testContains(service.CatalogService,
		`"mesh_id":"mymesh"`,
		`"service_id":"example"`,
		`"name":"example"`,
	))
}

func TestServiceEdge(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("edge", nil, map[string]int32{
		"proxy": 10808,
	})
	if err != nil {
		t.Fatal(err)
	}

	if count := len(service.Routes); count != 0 {
		t.Errorf("expected 0 routes but got %d", count)
	}

	if count := len(service.Ingresses.Clusters); count == 1 {
		t.Errorf("expected len(service.Ingresses.Clusters) to be 0 but got %d", count)
	}

	if count := len(service.Ingresses.Routes); count == 1 {
		t.Errorf("expected len(service.Ingresses.Routes) to be 0 but got %d", count)
	}
}

func TestServiceGMRedis(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("gm-redis", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Listener", testContains(service.Listener,
		`"listener_key":"gm-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["gm-redis"]`,
		`"port":10808`,
		`"active_http_filters":[]`,
	))
	t.Run("Proxy", testContains(service.Proxy,
		`"name":"gm-redis"`,
		`"proxy_key":"gm-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["gm-redis"]`,
		`"listener_keys":["gm-redis"]`,
	))
	t.Run("Route", testContains(service.Routes[0],
		`"route_key":"gm-redis"`,
		`"domain_key":"edge"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"gm-redis"`,
		`"route_match":{"path":"/services/gm-redis/"`,
		`"redirects":[{"from":"^/services/gm-redis$","to":"/services/gm-redis/"`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("HTTPEgresses is nil")
	}

	// No TCP egress is expected since none is needed to connect to gm-redis.
	if count := len(service.TCPEgresses); count != 0 {
		t.Errorf("Expected 0 TCP egress but got %d", count)
	}
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
		`"cluster_key":"example:5555"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"127.0.0.1","port":5555}]`,
	))

	if count := len(service.Ingresses.Routes); count != 1 {
		t.Fatalf("expected len(Ingresses.Routes) to be 1 but got %d", count)
	}

	t.Run("Route", testContains(service.Ingresses.Routes[0],
		`"route_key":"example:5555"`,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example:5555"`,
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
		key := fmt.Sprintf("example:%d", e.port)
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
			"greymatter.io/egress-http-local": `["othercluster"]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
		`"listener_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("HTTPEgresses is nil")
	}

	t.Run("Domain", testContains(service.HTTPEgresses.Domain,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", testContains(service.HTTPEgresses.Listener,
		`"listener_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-http"]`,
		`"port":10909`,
	))
	t.Run("Route", testContains(service.HTTPEgresses.Routes[0],
		`"route_key":"example-to-othercluster"`,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"othercluster"`,
		`"route_match":{"path":"/othercluster/"`,
		`"redirects":[{"from":"^/othercluster$","to":"/othercluster/"`,
	))
}

func TestServiceOneHTTPLocalEgressFromGMRedis(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("gm-redis",
		map[string]string{
			"greymatter.io/egress-http-local": `["othercluster"]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// gm-redis has an egress HTTP listener just so it can configure a metrics receiver.
	// This is because the metrics receiver capability was only added to the HTTP metrics filter.
	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["gm-redis","gm-redis-egress-http"]`,
		`"listener_keys":["gm-redis","gm-redis-egress-http"]`,
	))

	// The HTTP gm.metrics filter is only configured on egress for gm-redis.
	t.Run("Listener", testContains(service.HTTPEgresses.Listener,
		`"listener_key":"gm-redis-egress-http"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["gm-redis-egress-http"]`,
		`"port":10909`,
		`"active_http_filters":["gm.metrics"]`,
		`"http_filters":{"gm_metrics":{`,
		`"metrics_key_depth":"3"`,
		`"redis_connection_string":"redis://:`,
		`127.0.0.1:10808","push_interval_seconds`,
	))
}

func TestServiceMultipleHTTPLocalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/egress-http-local": `["othercluster1","othercluster2"]`,
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
			`"domain_key":"example-egress-http"`,
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
		"greymatter.io/egress-http-external": `[{"name":"google","host":"google.com","port":80}]`,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
		`"listener_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("ExternalEgresses is nil")
	}

	t.Run("Domain", testContains(service.HTTPEgresses.Domain,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", testContains(service.HTTPEgresses.Listener,
		`"listener_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-http"]`,
		`"port":10909`,
	))
	t.Run("Cluster", testContains(service.HTTPEgresses.Clusters[0],
		`"cluster_key":"example-to-external-google"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"google.com","port":80}]`,
	))
	t.Run("Route", testContains(service.HTTPEgresses.Routes[0],
		`"route_key":"example-to-external-google"`,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-external-google"`,
		`"route_match":{"path":"/google/"`,
		`"redirects":[{"from":"^/google$","to":"/google/"`,
	))
}

func TestServiceMultipleHTTPExternalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example", map[string]string{
		"greymatter.io/egress-http-external": `[
			{"name":"google","host":"google.com","port":80},
			{"name":"amazon","host":"amazon.com","port":80}
		]`,
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
			`"domain_key":"example-egress-http"`,
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
			"greymatter.io/egress-tcp-local": `["othercluster"]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-othercluster"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-othercluster"]`,
	))

	// 2 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 2 {
		t.Fatalf("Expected 2 TCP egresses but got %d", count)
	}

	t.Run("Domain", testContains(service.TCPEgresses[1].Domain,
		`"domain_key":"example-egress-tcp-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"port":10912`,
	))
	t.Run("Listener", testContains(service.TCPEgresses[1].Listener,
		`"listener_key":"example-egress-tcp-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-tcp-to-othercluster"]`,
		`"port":10912`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"othercluster"`,
	))
	t.Run("Route", testContains(service.TCPEgresses[1].Routes[0],
		`"route_key":"example-to-othercluster"`,
		`"domain_key":"example-egress-tcp-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"othercluster"`,
		`"route_match":{"path":"/"`,
	))
}

func TestServiceMultipleTCPLocalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/egress-tcp-local": `["c1","c2"]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-c1","example-egress-tcp-to-c2"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-c1","example-egress-tcp-to-c2"]`,
	))

	// 3 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 3 {
		t.Fatalf("Expected 3 TCP egress but got %d", count)
	}

	for i, e := range []struct {
		cluster string
		tcpPort int32
	}{
		{"c1", 10912},
		{"c2", 10913},
	} {
		t.Run(fmt.Sprintf("Domain:%s", e.cluster), testContains(service.TCPEgresses[i+1].Domain,
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"port":%d`, e.tcpPort),
		))
		t.Run(fmt.Sprintf("Listener:%s", e.cluster), testContains(service.TCPEgresses[i+1].Listener,
			fmt.Sprintf(`"listener_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"domain_keys":["example-egress-tcp-to-%s"]`, e.cluster),
			fmt.Sprintf(`"port":%d`, e.tcpPort),
			`"active_network_filters":["envoy.tcp_proxy"]`,
			`"network_filters":{"envoy_tcp_proxy":{`,
			fmt.Sprintf(`"cluster":"%s"`, e.cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", e.cluster), testContains(service.TCPEgresses[i+1].Routes[0],
			fmt.Sprintf(`"route_key":"example-to-%s"`, e.cluster),
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-%s"`, e.cluster),
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
			"greymatter.io/egress-tcp-external": `[{"name":"redis","host":"1.2.3.4","port":6379}]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-external-redis"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-external-redis"]`,
	))

	// 2 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 2 {
		t.Fatalf("Expected 2 TCP egresses but got %d", count)
	}

	t.Run("Domain", testContains(service.TCPEgresses[1].Domain,
		`"domain_key":"example-egress-tcp-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"port":10912`,
	))
	t.Run("Listener", testContains(service.TCPEgresses[1].Listener,
		`"listener_key":"example-egress-tcp-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-tcp-to-external-redis"]`,
		`"port":10912`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"example-to-external-redis"`,
	))
	t.Run("Cluster", testContains(service.TCPEgresses[1].Clusters[0],
		`"cluster_key":"example-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"1.2.3.4","port":6379}]`,
	))
	t.Run("Route", testContains(service.TCPEgresses[1].Routes[0],
		`"route_key":"example-to-external-redis"`,
		`"domain_key":"example-egress-tcp-to-external-redis"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-external-redis"`,
		`"route_match":{"path":"/"`,
	))
}

func TestServiceMultipleTCPExternalEgresses(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/egress-tcp-external": `[
				{"name":"s1","host":"1.1.1.1","port":1111},
				{"name":"s2","host":"2.2.2.2","port":2222}
			]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", testContains(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-external-s1","example-egress-tcp-to-external-s2"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-external-s1","example-egress-tcp-to-external-s2"]`,
	))

	// 3 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 3 {
		t.Fatalf("Expected 3 TCP egresses but got %d", count)
	}

	for i, e := range []struct {
		cluster string
		tcpPort int32
	}{
		{"s1", 10912},
		{"s2", 10913},
	} {
		t.Run(fmt.Sprintf("Domain:%s", e.cluster), testContains(service.TCPEgresses[i+1].Domain,
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-external-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"port":%d`, e.tcpPort),
		))
		t.Run(fmt.Sprintf("Listener:%s", e.cluster), testContains(service.TCPEgresses[i+1].Listener,
			fmt.Sprintf(`"listener_key":"example-egress-tcp-to-external-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"domain_keys":["example-egress-tcp-to-external-%s"]`, e.cluster),
			fmt.Sprintf(`"port":%d`, e.tcpPort),
			`"active_network_filters":["envoy.tcp_proxy"]`,
			`"network_filters":{"envoy_tcp_proxy":{`,
			fmt.Sprintf(`"cluster":"example-to-external-%s"`, e.cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", e.cluster), testContains(service.TCPEgresses[i+1].Routes[0],
			fmt.Sprintf(`"route_key":"example-to-external-%s"`, e.cluster),
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-external-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-external-%s"`, e.cluster),
			`"route_match":{"path":"/"`,
		))
	}
}

func TestParseFilters(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	for _, tc := range []struct {
		input    string
		expected map[string]bool
	}{
		{
			input:    "",
			expected: map[string]bool{},
		},
		{
			input:    `["one"]`,
			expected: map[string]bool{"one": true},
		},
		{
			input:    `["one","two"]`,
			expected: map[string]bool{"one": true, "two": true},
		},
		{
			input:    `[" one"," two"]`,
			expected: map[string]bool{"one": true, "two": true},
		},
		{
			input:    `["one","","two"]`,
			expected: map[string]bool{"one": true, "two": true},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			output := parseFilters(tc.input)
			if !reflect.DeepEqual(output, tc.expected) {
				t.Errorf("expected %v but got %v", tc.expected, output)
			}
		})
	}
}

func TestParseLocalEgressArgs(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	for _, tc := range []struct {
		name         string
		args         []EgressArgs
		annotation   string
		tcpPort      int32
		expectedArgs []EgressArgs
		expectedPort int32
	}{
		{
			name:         "nil and empty",
			args:         nil,
			annotation:   "",
			expectedArgs: []EgressArgs{},
		},
		{
			name:         "nil",
			args:         nil,
			annotation:   `["one"]`,
			expectedArgs: []EgressArgs{{Cluster: "one"}},
		},
		{
			name:         "non-nil",
			args:         []EgressArgs{},
			annotation:   `["one"]`,
			expectedArgs: []EgressArgs{{Cluster: "one"}},
		},
		{
			name:         "multiple",
			args:         []EgressArgs{},
			annotation:   `["one","two"]`,
			expectedArgs: []EgressArgs{{Cluster: "one"}, {Cluster: "two"}},
		},
		{
			name:         "trim",
			args:         []EgressArgs{},
			annotation:   `[" one"," two"]`,
			expectedArgs: []EgressArgs{{Cluster: "one"}, {Cluster: "two"}},
		},
		{
			name:         "adjacent commas",
			args:         []EgressArgs{},
			annotation:   `["one","","two"]`,
			expectedArgs: []EgressArgs{{Cluster: "one"}, {Cluster: "two"}},
		},
		{
			name:       "tcp",
			args:       []EgressArgs{{Cluster: "gm-redis", TCPPort: 10910}},
			annotation: `["one"]`,
			tcpPort:    10912,
			expectedArgs: []EgressArgs{
				{Cluster: "gm-redis", TCPPort: 10910},
				{Cluster: "one", TCPPort: 10912},
			},
			expectedPort: 10913,
		},
		{
			name:       "tcp multiple",
			args:       []EgressArgs{{Cluster: "gm-redis", TCPPort: 10910}},
			annotation: `["one","two"]`,
			tcpPort:    10912,
			expectedArgs: []EgressArgs{
				{Cluster: "gm-redis", TCPPort: 10910},
				{Cluster: "one", TCPPort: 10912},
				{Cluster: "two", TCPPort: 10913},
			},
			expectedPort: 10914,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, port := parseLocalEgressArgs(tc.args, tc.annotation, tc.tcpPort)
			if !reflect.DeepEqual(output, tc.expectedArgs) {
				t.Errorf("args: expected %v but got %v", tc.expectedArgs, output)
			}
			if port != tc.expectedPort {
				t.Errorf("port: expected %d but got %d", tc.expectedPort, port)
			}
		})
	}
}

func TestParseExternalEgressArgs(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	for _, tc := range []struct {
		name         string
		args         []EgressArgs
		annotation   string
		tcpPort      int32
		expectedArgs []EgressArgs
		expectedPort int32
	}{
		{
			name:         "nil and empty",
			args:         nil,
			annotation:   "",
			expectedArgs: []EgressArgs{},
		},
		{
			name: "nil",
			args: nil,
			annotation: `[
				{
					"name": "cluster",
					"host": "host",
					"port": 8080
				}
			]`,
			expectedArgs: []EgressArgs{
				{IsExternal: true, Cluster: "cluster", Host: "host", Port: 8080},
			},
		},
		{
			name: "one",
			args: []EgressArgs{},
			annotation: `[
				{
					"name": "cluster",
					"host": "host",
					"port": 8080
				}
			]`,
			expectedArgs: []EgressArgs{
				{IsExternal: true, Cluster: "cluster", Host: "host", Port: 8080},
			},
		},
		{
			name: "multiple",
			args: []EgressArgs{},
			annotation: `[
				{
					"name": "c1",
					"host": "h1",
					"port": 8080
				},
				{
					"name": "c2",
					"host": "h2",
					"port": 3000
				}
			]`,
			expectedArgs: []EgressArgs{
				{IsExternal: true, Cluster: "c1", Host: "h1", Port: 8080},
				{IsExternal: true, Cluster: "c2", Host: "h2", Port: 3000},
			},
		},
		{
			name: "tcp",
			args: []EgressArgs{
				{IsExternal: true, Cluster: "gm-redis", Host: "redis://extserver", Port: 6379, TCPPort: 10910},
			},
			annotation: `[
				{
					"name": "c1",
					"host": "h1",
					"port": 8080
				}
			]`,
			tcpPort: 10912,
			expectedArgs: []EgressArgs{
				{IsExternal: true, Cluster: "gm-redis", Host: "redis://extserver", Port: 6379, TCPPort: 10910},
				{IsExternal: true, Cluster: "c1", Host: "h1", Port: 8080, TCPPort: 10912},
			},
			expectedPort: 10913,
		},
		{
			name: "tcp multiple",
			args: []EgressArgs{
				{IsExternal: true, Cluster: "gm-redis", Host: "redis://extserver", Port: 6379, TCPPort: 10910},
			},
			annotation: `[
				{
					"name": "c1",
					"host": "h1",
					"port": 8080
				},
				{
					"name": "c2",
					"host": "h2",
					"port": 3000
				}
			]`,
			tcpPort: 10912,
			expectedArgs: []EgressArgs{
				{IsExternal: true, Cluster: "gm-redis", Host: "redis://extserver", Port: 6379, TCPPort: 10910},
				{IsExternal: true, Cluster: "c1", Host: "h1", Port: 8080, TCPPort: 10912},
				{IsExternal: true, Cluster: "c2", Host: "h2", Port: 3000, TCPPort: 10913},
			},
			expectedPort: 10914,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, port := parseExternalEgressArgs(tc.args, tc.annotation, tc.tcpPort)
			if !reflect.DeepEqual(output, tc.expectedArgs) {
				t.Errorf("args: expected %v but got %v", tc.expectedArgs, output)
			}
			if port != tc.expectedPort {
				t.Errorf("port: expected %d but got %d", tc.expectedPort, port)
			}
		})
	}
}

func loadMock(t *testing.T) *Fabric {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	if err := Init(); err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	return New((&v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mymesh"},
		Spec: v1alpha1.MeshSpec{
			Zone:           "myzone",
			ReleaseVersion: "1.7",
		},
	}).Options(""))
}

func testContains(obj json.RawMessage, subs ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		if len(obj) == 0 {
			t.Fatal("json is empty")
		}
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
