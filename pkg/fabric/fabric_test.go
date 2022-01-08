package fabric

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/assert"
	"github.com/greymatter-io/operator/pkg/cueutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEdgeDomain(t *testing.T) {
	f := loadMock(t)

	assert.JSONHasSubstrings(f.EdgeDomain(),
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

	t.Run("CatalogService", assert.JSONHasSubstrings(service.CatalogService,
		`"mesh_id":"mymesh"`,
		`"service_id":"example"`,
		`"name":"example"`,
	))

	t.Run("Domain", assert.JSONHasSubstrings(service.Domain,
		`"domain_key":"example"`,
		`"zone_key":"myzone"`,
		`"port":10808`,
	))
	t.Run("Listener", assert.JSONHasSubstrings(service.Listener,
		`"listener_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example"]`,
		`"port":10808`,
		`"active_http_filters":["gm.metrics","gm.observables"]`,
		`"http_filters":{"gm_metrics":{`,
		`"metrics_key_depth":"3"`,
		`"redis_connection_string":"redis://`,
		`"secret_name":"spiffe://greymatter.io/mymesh.example"`,
		`"subject_names":["spiffe://greymatter.io/mymesh.edge"]`,
	))
	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"name":"example"`,
		`"proxy_key":"example"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis"]`,
	))
	t.Run("Cluster", assert.JSONHasSubstrings(service.Clusters[0],
		`"name":"example"`,
		`"cluster_key":"example"`,
		`"zone_key":"myzone"`,
		`"secret_name":"spiffe://greymatter.io/mymesh.edge"`,
		`"subject_names":["spiffe://greymatter.io/mymesh.example"]`,
	))
	t.Run("Route", assert.JSONHasSubstrings(service.Routes[0],
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

	t.Run("gm-redis:Domain", assert.JSONHasSubstrings(service.TCPEgresses[0].Domain,
		`"domain_key":"example-egress-tcp-to-gm-redis"`,
		`"zone_key":"myzone"`,
		`"port":10910`,
	))
	t.Run("gm-redis:Listener", assert.JSONHasSubstrings(service.TCPEgresses[0].Listener,
		`"listener_key":"example-egress-tcp-to-gm-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-tcp-to-gm-redis"]`,
		`"port":10910`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"gm-redis"`,
	))
	t.Run("gm-redis:Cluster", assert.JSONHasSubstrings(service.TCPEgresses[0].Clusters[0],
		`"name":"gm-redis"`,
		`"cluster_key":"example-to-gm-redis"`,
		`"zone_key":"myzone"`,
	))
	t.Run("gm-redis:Route", assert.JSONHasSubstrings(service.TCPEgresses[0].Routes[0],
		`"route_key":"example-to-gm-redis"`,
		`"domain_key":"example-egress-tcp-to-gm-redis"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-gm-redis"`,
		`"route_match":{"path":"/"`,
	))

	if len(service.LocalEgresses) != 1 || service.LocalEgresses[0] != "gm-redis" {
		t.Errorf("expected 1 local egress, 'gm-redis', but got %v", service.LocalEgresses)
	}
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

	t.Run("Listener", assert.JSONHasSubstrings(service.Listener,
		`"listener_key":"gm-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["gm-redis"]`,
		`"port":10808`,
	))
	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"name":"gm-redis"`,
		`"proxy_key":"gm-redis"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["gm-redis"]`,
		`"listener_keys":["gm-redis"]`,
	))
	t.Run("Route", assert.JSONHasSubstrings(service.Routes[0],
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

	if count := len(service.LocalEgresses); count != 0 {
		t.Errorf("expected 0 local egresses but got %d", count)
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

	t.Run("Cluster", assert.JSONHasSubstrings(service.Ingresses.Clusters[0],
		`"cluster_key":"example:5555"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"127.0.0.1","port":5555}]`,
	))

	if count := len(service.Ingresses.Routes); count != 1 {
		t.Fatalf("expected len(Ingresses.Routes) to be 1 but got %d", count)
	}

	t.Run("Route", assert.JSONHasSubstrings(service.Ingresses.Routes[0],
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
		t.Run("Cluster", assert.JSONHasSubstrings(service.Ingresses.Clusters[i],
			fmt.Sprintf(`"cluster_key":"%s"`, key),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"instances":[{"host":"127.0.0.1","port":%d}]`, e.port),
		))
		t.Run("Route", assert.JSONHasSubstrings(service.Ingresses.Routes[i],
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

	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"domain_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
		`"listener_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("HTTPEgresses is nil")
	}

	t.Run("Domain", assert.JSONHasSubstrings(service.HTTPEgresses.Domain,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", assert.JSONHasSubstrings(service.HTTPEgresses.Listener,
		`"listener_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-http"]`,
		`"port":10909`,
	))
	t.Run("Cluster", assert.JSONHasSubstrings(service.HTTPEgresses.Clusters[0],
		`"name":"othercluster"`,
		`"cluster_key":"example-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"secret_name":"spiffe://greymatter.io/mymesh.example"`,
		`"subject_names":["spiffe://greymatter.io/mymesh.othercluster"]`,
	))
	t.Run("Route", assert.JSONHasSubstrings(service.HTTPEgresses.Routes[0],
		`"route_key":"example-to-othercluster"`,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-othercluster"`,
		`"route_match":{"path":"/othercluster/"`,
		`"redirects":[{"from":"^/othercluster$","to":"/othercluster/"`,
	))

	if !strings.Contains(strings.Join(service.LocalEgresses, "|"), "othercluster") {
		t.Errorf("expected 'othercluster' to be a local egress, but got %v", service.LocalEgresses)
	}
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

	localEgresses := strings.Join(service.LocalEgresses, "|")

	for i, name := range []string{"othercluster1", "othercluster2"} {
		t.Run(fmt.Sprintf("Cluster:%s", name), assert.JSONHasSubstrings(service.HTTPEgresses.Clusters[i],
			fmt.Sprintf(`"name":"%s"`, name),
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, name),
			`"zone_key":"myzone"`,
			`"secret_name":"spiffe://greymatter.io/mymesh.example"`,
			fmt.Sprintf(`"subject_names":["spiffe://greymatter.io/mymesh.%s"]`, name),
		))
		t.Run(fmt.Sprintf("Route:%s", name), assert.JSONHasSubstrings(service.HTTPEgresses.Routes[i],
			fmt.Sprintf(`"route_key":"example-to-%s"`, name),
			`"domain_key":"example-egress-http"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, name),
			fmt.Sprintf(`"route_match":{"path":"/%s/"`, name),
			fmt.Sprintf(`"redirects":[{"from":"^/%s$","to":"/%s/"`, name, name),
		))
		if !strings.Contains(localEgresses, name) {
			t.Errorf("expected '%s' to be a local egress, but got %v", name, service.LocalEgresses)
		}
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

	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"domain_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
		`"listener_keys":["example","example-egress-http","example-egress-tcp-to-gm-redis"]`,
	))

	if service.HTTPEgresses == nil {
		t.Fatal("ExternalEgresses is nil")
	}

	t.Run("Domain", assert.JSONHasSubstrings(service.HTTPEgresses.Domain,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"port":10909`,
	))
	t.Run("Listener", assert.JSONHasSubstrings(service.HTTPEgresses.Listener,
		`"listener_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-http"]`,
		`"port":10909`,
	))
	t.Run("Cluster", assert.JSONHasSubstrings(service.HTTPEgresses.Clusters[0],
		`"cluster_key":"example-to-google"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"google.com","port":80}]`,
	))
	t.Run("Route", assert.JSONHasSubstrings(service.HTTPEgresses.Routes[0],
		`"route_key":"example-to-google"`,
		`"domain_key":"example-egress-http"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-google"`,
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
		t.Run(fmt.Sprintf("Cluster:%s", cluster), assert.JSONHasSubstrings(service.HTTPEgresses.Clusters[i],
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"instances":[{"host":"%s.com","port":80}]`, cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", cluster), assert.JSONHasSubstrings(service.HTTPEgresses.Routes[i],
			fmt.Sprintf(`"route_key":"example-to-%s"`, cluster),
			`"domain_key":"example-egress-http"`,
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, cluster),
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

	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-othercluster"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-othercluster"]`,
	))

	// 2 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 2 {
		t.Fatalf("Expected 2 TCP egresses but got %d", count)
	}

	t.Run("Domain", assert.JSONHasSubstrings(service.TCPEgresses[1].Domain,
		`"domain_key":"example-egress-tcp-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"port":10912`,
	))
	t.Run("Listener", assert.JSONHasSubstrings(service.TCPEgresses[1].Listener,
		`"listener_key":"example-egress-tcp-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-tcp-to-othercluster"]`,
		`"port":10912`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"othercluster"`,
	))
	t.Run("Cluster", assert.JSONHasSubstrings(service.TCPEgresses[1].Clusters[0],
		`"name":"othercluster"`,
		`"cluster_key":"example-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"secret_name":"spiffe://greymatter.io/mymesh.example"`,
		`"subject_names":["spiffe://greymatter.io/mymesh.othercluster"]`,
	))
	t.Run("Route", assert.JSONHasSubstrings(service.TCPEgresses[1].Routes[0],
		`"route_key":"example-to-othercluster"`,
		`"domain_key":"example-egress-tcp-to-othercluster"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-othercluster"`,
		`"route_match":{"path":"/"`,
	))

	if !strings.Contains(strings.Join(service.LocalEgresses, "|"), "othercluster") {
		t.Errorf("expected 'othercluster' to be a local egress, but got %v", service.LocalEgresses)
	}
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

	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-c1","example-egress-tcp-to-c2"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-c1","example-egress-tcp-to-c2"]`,
	))

	// 3 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 3 {
		t.Fatalf("Expected 3 TCP egress but got %d", count)
	}

	localEgresses := strings.Join(service.LocalEgresses, "|")

	for i, e := range []struct {
		cluster string
		tcpPort int32
	}{
		{"c1", 10912},
		{"c2", 10913},
	} {
		t.Run(fmt.Sprintf("Domain:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Domain,
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"port":%d`, e.tcpPort),
		))
		t.Run(fmt.Sprintf("Listener:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Listener,
			fmt.Sprintf(`"listener_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"domain_keys":["example-egress-tcp-to-%s"]`, e.cluster),
			fmt.Sprintf(`"port":%d`, e.tcpPort),
			`"active_network_filters":["envoy.tcp_proxy"]`,
			`"network_filters":{"envoy_tcp_proxy":{`,
			fmt.Sprintf(`"cluster":"%s"`, e.cluster),
		))
		t.Run(fmt.Sprintf("Cluster:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Clusters[0],
			fmt.Sprintf(`"name":"%s"`, e.cluster),
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			`"secret_name":"spiffe://greymatter.io/mymesh.example"`,
			fmt.Sprintf(`"subject_names":["spiffe://greymatter.io/mymesh.%s"]`, e.cluster),
		))
		t.Run(fmt.Sprintf("Route:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Routes[0],
			fmt.Sprintf(`"route_key":"example-to-%s"`, e.cluster),
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, e.cluster),
			`"route_match":{"path":"/"`,
		))
		if !strings.Contains(localEgresses, e.cluster) {
			t.Errorf("expected '%s' to be a local egress, but got %v", e.cluster, service.LocalEgresses)
		}
	}
}

func TestServiceOneTCPExternalEgress(t *testing.T) {
	f := loadMock(t)

	service, err := f.Service("example",
		map[string]string{
			"greymatter.io/egress-tcp-external": `[{"name":"svc","host":"1.2.3.4","port":80}]`,
		}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-svc"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-svc"]`,
	))

	// 2 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 2 {
		t.Fatalf("Expected 2 TCP egresses but got %d", count)
	}

	t.Run("Domain", assert.JSONHasSubstrings(service.TCPEgresses[1].Domain,
		`"domain_key":"example-egress-tcp-to-svc"`,
		`"zone_key":"myzone"`,
		`"port":10912`,
	))
	t.Run("Listener", assert.JSONHasSubstrings(service.TCPEgresses[1].Listener,
		`"listener_key":"example-egress-tcp-to-svc"`,
		`"zone_key":"myzone"`,
		`"domain_keys":["example-egress-tcp-to-svc"]`,
		`"port":10912`,
		`"active_network_filters":["envoy.tcp_proxy"]`,
		`"network_filters":{"envoy_tcp_proxy":{`,
		`"cluster":"svc"`,
	))
	t.Run("Cluster", assert.JSONHasSubstrings(service.TCPEgresses[1].Clusters[0],
		`"cluster_key":"example-to-svc"`,
		`"zone_key":"myzone"`,
		`"instances":[{"host":"1.2.3.4","port":80}]`,
	))
	t.Run("Route", assert.JSONHasSubstrings(service.TCPEgresses[1].Routes[0],
		`"route_key":"example-to-svc"`,
		`"domain_key":"example-egress-tcp-to-svc"`,
		`"zone_key":"myzone"`,
		`"cluster_key":"example-to-svc"`,
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

	t.Run("Proxy", assert.JSONHasSubstrings(service.Proxy,
		`"domain_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-s1","example-egress-tcp-to-s2"]`,
		`"listener_keys":["example","example-egress-tcp-to-gm-redis","example-egress-tcp-to-s1","example-egress-tcp-to-s2"]`,
	))

	// 3 TCP egresses are expected since `-egress-tcp-to-gm-redis` is added by default.
	if count := len(service.TCPEgresses); count != 3 {
		t.Fatalf("Expected 3 TCP egresses but got %d", count)
	}

	for i, e := range []struct {
		cluster string
		tcpPort int32
		host    string
		port    int32
	}{
		{"s1", 10912, "1.1.1.1", 1111},
		{"s2", 10913, "2.2.2.2", 2222},
	} {
		t.Run(fmt.Sprintf("Domain:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Domain,
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"port":%d`, e.tcpPort),
		))
		t.Run(fmt.Sprintf("Listener:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Listener,
			fmt.Sprintf(`"listener_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"domain_keys":["example-egress-tcp-to-%s"]`, e.cluster),
			fmt.Sprintf(`"port":%d`, e.tcpPort),
			`"active_network_filters":["envoy.tcp_proxy"]`,
			`"network_filters":{"envoy_tcp_proxy":{`,
			fmt.Sprintf(`"cluster":"%s"`, e.cluster),
		))
		t.Run("Cluster", assert.JSONHasSubstrings(service.TCPEgresses[i+1].Clusters[0],
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"instances":[{"host":"%s","port":%d}]`, e.host, e.port),
		))
		t.Run(fmt.Sprintf("Route:%s", e.cluster), assert.JSONHasSubstrings(service.TCPEgresses[i+1].Routes[0],
			fmt.Sprintf(`"route_key":"example-to-%s"`, e.cluster),
			fmt.Sprintf(`"domain_key":"example-egress-tcp-to-%s"`, e.cluster),
			`"zone_key":"myzone"`,
			fmt.Sprintf(`"cluster_key":"example-to-%s"`, e.cluster),
			`"route_match":{"path":"/"`,
		))
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
