package meshobjects

import (
	"fmt"
)

func (c *Client) MkZone(zoneKey string) error {
	template := `{"zone_key":"%s","name":"%s"}`
	return c.GetOrMake("zone", zoneKey, raw(template, zoneKey, zoneKey))
}

func (c *Client) MkProxy(zoneKey, clusterName string) error {
	key := fmt.Sprintf("%s.%s", zoneKey, clusterName)

	domain := `{
		"zone_key":"%s",
		"domain_key":"%s",
		"name":"*",
		"port":10808
	}`
	object := raw(domain, zoneKey, key)
	if err := c.GetOrMake("domain", key, object); err != nil {
		return err
	}

	listener := `{
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
	}`
	object = raw(listener, zoneKey, key, key, clusterName)
	if err := c.GetOrMake("listener", key, object); err != nil {
		return err
	}

	proxy := `{
		"zone_key":"%s",
		"proxy_key":"%s",
		"domain_keys":["%s"],
		"listener_keys":["%s"],
		"name":"%s"
	}`
	object = raw(proxy, zoneKey, key, key, key, clusterName)
	if err := c.GetOrMake("proxy", key, object); err != nil {
		return err
	}

	cluster := `{
		"zone_key":"%s",
		"cluster_key":"%s",
		"name":"%s"
	}`
	object = raw(cluster, zoneKey, key, clusterName)
	if err := c.GetOrMake("cluster", key, object); err != nil {
		return err
	}

	return nil
}

func (c *Client) MkService(zoneKey, clusterName, port string) error {
	sidecarKey := fmt.Sprintf("%s.%s", zoneKey, clusterName)
	serviceKey := fmt.Sprintf("%s.service", sidecarKey)
	serviceClusterName := fmt.Sprintf("%s.service", clusterName)
	edgeKey := fmt.Sprintf("%s.edge", zoneKey)

	cluster := `{
		"zone_key":"%s",
		"cluster_key":"%s",
		"name":"%s",
		"instances":[
			{
				"host":"127.0.0.1",
				"port":%s
			}
		]
	}`
	object := raw(cluster, zoneKey, serviceKey, serviceClusterName, port)
	if err := c.GetOrMake("cluster", serviceKey, object); err != nil {
		return err
	}

	switch clusterName {
	case "dashboard":
		if err := c.mkRoute(
			zoneKey,
			sidecarKey+".a",
			edgeKey,
			"/",
			"prefix",
			"",
			sidecarKey,
		); err != nil {
			return err
		}
	default:
		if err := c.mkRoute(
			zoneKey,
			sidecarKey+".a",
			edgeKey,
			fmt.Sprintf("/services/%s/latest", clusterName),
			"exact",
			fmt.Sprintf("/services/%s/latest/", clusterName),
			sidecarKey,
		); err != nil {
			return err
		}

		if err := c.mkRoute(
			zoneKey,
			sidecarKey+".b",
			edgeKey,
			fmt.Sprintf("/services/%s/latest/", clusterName),
			"prefix",
			"/",
			sidecarKey,
		); err != nil {
			return err
		}
	}

	if err := c.mkRoute(
		zoneKey,
		sidecarKey+".c",
		sidecarKey,
		"/",
		"prefix",
		"/",
		serviceKey,
	); err != nil {
		return err
	}

	return nil
}

func (c *Client) mkRoute(zoneKey, routeKey, domainKey, path, matchType, rewrite, clusterKey string) error {
	template := `{
		"zone_key":"%s",
		"route_key":"%s",
		"domain_key":"%s",
		"route_match": {
			"path": "%s",
			"match_type": "%s"
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
	}`
	object := raw(template, zoneKey, routeKey, domainKey, path, matchType, rewrite, clusterKey)
	return c.GetOrMake("route", routeKey, object)
}
