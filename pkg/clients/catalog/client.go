package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	catalogv2 "github.com/greymatter-io/gm-catalog/pkg/api/client"
	"github.com/greymatter-io/gm-catalog/pkg/discovery/meshclient"
	"github.com/greymatter-io/gm-catalog/pkg/model"
	"github.com/greymatter-io/operator/pkg/clients"
)

type Client interface {
	CreateMesh(meshID, namespace string) bool
	CreateService(
		serviceID,
		meshID,
		name,
		version,
		description,
		owner,
		apiEndpoint,
		documentation,
		capability string) bool
}

func NewClient(meshVersion, addr string, logger logr.Logger) Client {
	switch meshVersion {
	case "1.3":
		return &V1Client{
			client: &http.Client{Timeout: time.Second * 3},
			addr:   addr,
			logger: logger,
		}
	default:
		return &V2Client{
			client: catalogv2.NewClient(addr),
			logger: logger,
		}
	}
}

type V2Client struct {
	client *catalogv2.Client
	logger logr.Logger
}

func (v2 *V2Client) CreateMesh(meshID, namespace string) bool {
	resp, err := v2.client.GetMesh(meshID)
	if err == nil && resp.StatusCode != http.StatusNotFound {
		return true
	}
	resp, err = v2.client.CreateMesh(meshclient.Config{
		MeshID:   meshID,
		MeshType: meshclient.GreyMatter,
		Name:     "Grey Matter Core",
		Sessions: map[string]meshclient.SessionConfig{
			"default": {
				URL:  fmt.Sprintf("control.%s.svc:50000", namespace),
				Zone: meshID,
			},
		},
	})
	if err != nil {
		v2.logger.Error(err, "failed to create mesh")
		return false
	}
	if resp.StatusCode != http.StatusOK {
		v2.logger.Error(fmt.Errorf("%s", string(resp.Data)), "failed to create mesh")
		return false
	}
	v2.logger.Info("Added Mesh to Catalog", "MeshID", meshID, "Namespace", namespace)
	return true
}

func (v2 *V2Client) CreateService(
	serviceID,
	meshID,
	name,
	version,
	description,
	owner,
	apiEndpoint,
	documentation,
	capability string) bool {
	resp, err := v2.client.GetService(meshID, serviceID)
	if err == nil && resp.StatusCode != http.StatusNotFound {
		return true
	}
	resp, err = v2.client.CreateService(model.Service{
		ServiceID:     serviceID,
		MeshID:        meshID,
		Name:          name,
		Version:       version,
		Description:   description,
		Owner:         owner,
		ApiEndpoint:   apiEndpoint,
		Documentation: documentation,
		Capability:    capability,
	})
	if err != nil {
		v2.logger.Error(err, "failed to create service")
		return false
	}
	if resp.StatusCode != http.StatusOK {
		v2.logger.Error(fmt.Errorf("%s", string(resp.Data)), "failed to create service")
		return false
	}
	v2.logger.Info("Added Service to Catalog", "ServiceID", serviceID, "MeshID", meshID)
	return true
}

type V1Client struct {
	client *http.Client
	addr   string
	logger logr.Logger
}

func (v1 *V1Client) CreateMesh(meshID, namespace string) bool {
	url := fmt.Sprintf("%s/zones/%s", v1.addr, meshID)
	if _, err := clients.Do(v1.client, http.MethodGet, url, nil); err == nil {
		return true
	}
	url = fmt.Sprintf("%s/zones", v1.addr)
	if _, err := clients.Do(v1.client, http.MethodPost, url, json.RawMessage(fmt.Sprintf(`{
		"zoneName": "%s",
		"requestCluster": "edge",
		"serverAddress": "control.%s.svc:50000"
	}`, meshID, namespace))); err != nil {
		v1.logger.Error(err, "failed to create mesh")
		return false
	}
	v1.logger.Info("Added Mesh to Catalog", "MeshID", meshID, "Namespace", namespace)
	return true
}

func (v1 *V1Client) CreateService(
	serviceID,
	meshID,
	name,
	version,
	description,
	owner,
	apiEndpoint,
	documentation,
	capability string) bool {
	url := fmt.Sprintf("%s/clusters/%s?meshID=%s", v1.addr, serviceID, meshID)
	resp, err := clients.Do(v1.client, http.MethodGet, url, nil)
	if err == nil {
		var slice []interface{}
		json.Unmarshal(resp, &slice)
		if len(slice) > 0 {
			return true
		}
	}
	url = fmt.Sprintf("%s/clusters", v1.addr)
	if _, err := clients.Do(v1.client, http.MethodPost, url, json.RawMessage(fmt.Sprintf(`{
		"clusterName": "%s",
		"zoneName": "%s",
		"name": "%s",
		"version": "%s",
		"owner": "%s",
		"documentation": "%s",
		"capability": "%s"
		}`, serviceID, meshID, name, version, owner, documentation, capability))); err != nil {
		v1.logger.Error(err, "failed to create service")
		return false
	}
	v1.logger.Info("Added Service to Catalog", "ServiceID", serviceID, "MeshID", meshID)
	return true
}
