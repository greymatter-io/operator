package catalogentries

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	catalogclient "github.com/greymatter-io/gm-catalog/pkg/api/client"
	"github.com/greymatter-io/gm-catalog/pkg/discovery/meshclient"
	"github.com/greymatter-io/gm-catalog/pkg/model"
)

type Client interface {
	Ping() bool
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

func NewCatalogClient(meshVersion, addr string, logger logr.Logger) Client {
	switch meshVersion {
	case "1.3":
		return &V1Client{
			client: &http.Client{Timeout: time.Second * 5},
			addr:   addr,
			logger: logger,
		}
	default:
		return &V2Client{
			client: catalogclient.NewClient(addr),
			logger: logger,
		}
	}
}

type V2Client struct {
	client *catalogclient.Client
	logger logr.Logger
	// todo: add cache
}

func (v2 *V2Client) Ping() bool {
	resp, err := v2.client.Ping()
	v2.logger.Info("ping", "status", resp.StatusCode, "error", err)
	return err == nil && resp.StatusCode == http.StatusOK
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
	v2.logger.Info("Create Mesh Response", "status", resp.StatusCode, "error", err)
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
	// todo: add cache
}

func (v1 *V1Client) Ping() bool {
	// todo
	return true
}

func (v1 *V1Client) CreateMesh(meshID, namespace string) bool {
	// todo
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
	// todo
	return true
}
