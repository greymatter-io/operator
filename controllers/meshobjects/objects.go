package meshobjects

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c *Client) MkMeshObjects(zone string, clusters []string) error {
	zoneObject := fmt.Sprintf(`{"zone_key":"%s","name":"%s"}`, zone, zone)
	zoneBytes := json.RawMessage(zoneObject)
	if err := c.Make("zone", zone, zoneBytes); err != nil {
		return err
	}

	if err := c.mkSidecarObjects(zone, "edge"); err != nil {
		return err
	}

ClusterLoop:
	for _, cluster := range clusters {
		split := strings.Split(cluster, ":")
		if len(split) != 2 {
			continue ClusterLoop
		}
		if err := c.mkSidecarObjects(zone, split[0]); err != nil {
			return err
		}
		if err := c.mkServiceObjects(zone, split[0], split[1]); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) mkSidecarObjects(zone, cluster string) error {
	key := fmt.Sprintf("%s-%s", zone, cluster)

	domainObject := fmt.Sprintf(`{"zone_key":"%s","domain_key":"%s","name":"*","port":10808}`, zone, key)
	domainBytes := json.RawMessage(domainObject)
	if err := c.Make("domain", key, domainBytes); err != nil {
		return err
	}

	listenerObject := fmt.Sprintf(`{
		"zone_key":"%s",
		"listener_key":"%s",
		"domain_keys":["%s"],
		"name":"%s",
		"ip":"0.0.0.0",
		"port":10808,
		"protocol":"http_auto",
		"active_http_filters": ["gm.metrics"],
		"http_filters": {
			"gm_metrics": {
				"metrics_port": 8081,
				"metrics_host": "0.0.0.0",
				"metrics_dashboard_uri_path": "/metrics",
				"metrics_prometheus_uri_path": "/prometheus",
				"metrics_ring_buffer_size": 4096,
				"prometheus_system_metrics_interval_seconds": 15,
				"metrics_key_function": "none"
			}
		}
	}`, zone, key, key, cluster)
	listenerBytes := json.RawMessage(listenerObject)
	if err := c.Make("listener", key, listenerBytes); err != nil {
		return err
	}

	proxyObject := fmt.Sprintf(`{
		"zone_key":"%s",
		"proxy_key":"%s",
		"domain_keys":["%s"],
		"listener_keys":["%s"],
		"name":"%s"
	}`, zone, key, key, key, cluster)
	proxyBytes := json.RawMessage(proxyObject)
	if err := c.Make("proxy", key, proxyBytes); err != nil {
		return err
	}

	clusterObject := fmt.Sprintf(`{"zone_key":"%s","cluster_key":"%s","name":"%s"}`, zone, key, cluster)
	clusterBytes := json.RawMessage(clusterObject)
	if err := c.Make("cluster", key, clusterBytes); err != nil {
		return err
	}

	return nil
}

func (c *Client) mkServiceObjects(zone, cluster, port string) error {
	sidecarKey := fmt.Sprintf("%s-%s", zone, cluster)
	serviceKey := fmt.Sprintf("service-%s", sidecarKey)
	serviceCluster := fmt.Sprintf("service-%s", cluster)
	edgeKey := fmt.Sprintf("%s-edge", zone)

	clusterObject := fmt.Sprintf(`{
		"zone_key":"%s",
		"cluster_key":"%s",
		"name":"%s",
		"instances":[
			{
				"host":"127.0.0.1",
				"port":%s
			}
		]
	}`, zone, serviceKey, serviceCluster, port)
	clusterBytes := json.RawMessage(clusterObject)
	if err := c.Make("cluster", serviceCluster, clusterBytes); err != nil {
		return err
	}

	if err := c.mkRoute(zone,
		fmt.Sprintf("%s-a", sidecarKey),
		edgeKey,
		fmt.Sprintf("/services/%s/latest", cluster),
		fmt.Sprintf("/services/%s/latest/", cluster),
		sidecarKey,
	); err != nil {
		return err
	}

	if err := c.mkRoute(zone,
		fmt.Sprintf("%s-b", sidecarKey),
		edgeKey,
		fmt.Sprintf("/services/%s/latest/", cluster),
		"/",
		sidecarKey,
	); err != nil {
		return err
	}

	if err := c.mkRoute(zone,
		fmt.Sprintf("%s-c", sidecarKey),
		sidecarKey,
		"/",
		"/",
		serviceKey,
	); err != nil {
		return err
	}

	return nil
}

func (c *Client) mkRoute(zone, key, dk, path, rewrite, ck string) error {
	routeObject := fmt.Sprintf(`{
		"zone_key":"%s",
		"route_key":"%s",
		"domain_key":"%s",
		"route_match": {
			"path": "%s",
			"match_type": "prefix"
		},
		"prefix_rewrite":"%s",
		"rules": [
			{
				"constraints": {
					"light": [
						{
							"cluster_key": "%s",
							"weight": 1
						}
					]
				}
			}
		]
	}`, zone, key, dk, path, rewrite, ck)
	routeBytes := json.RawMessage(routeObject)
	if err := c.Make("route", key, routeBytes); err != nil {
		return err
	}

	return nil
}
