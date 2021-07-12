package control

import "github.com/go-logr/logr"

type Client interface {
	CreateSidecar(meshID, clusterName string) bool
	CreateService(meshID, clusterName, port string) bool
}

func NewClient(meshVersion, addr string, logger logr.Logger) Client {
	return &APIClient{
		logger: logger,
	}
}

type APIClient struct {
	// imported api client
	logger logr.Logger
}

func (api *APIClient) CreateSidecar(meshID, clusterName string) bool {
	// todo
	return true
}

func (api *APIClient) CreateService(meshID, clusterName, port string) bool {
	// todo
	return true
}
