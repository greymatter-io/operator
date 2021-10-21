// Package cli executes greymatter CLI commands to configure mesh behavior
// in Control and Catalog APIs in each install namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package cli

import (
	"context"
	"fmt"
	"sync"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("cli")
)

type CLI struct {
	*sync.RWMutex
	clients map[string]*client
	ctx     context.Context
}

// Returns *CLI for storing clients used to execute greymatter CLI commands.
func New(ctx context.Context) (*CLI, error) {
	v, err := cliversion()
	if err != nil {
		logger.Error(err, "Failed to initialize greymatter CLI")
		return nil, err
	}

	logger.Info("Using greymatter CLI", "Version", v)

	if err := fabric.Init(); err != nil {
		logger.Error(err, "Failed to initialize Fabric templates")
		return nil, err
	}

	gmcli := &CLI{
		RWMutex: &sync.RWMutex{},
		clients: make(map[string]*client),
		ctx:     ctx,
	}

	// Cancel all client goroutines if package context is done.
	go func(c *CLI) {
		<-ctx.Done()
		c.RLock()
		defer c.RUnlock()
		for _, cl := range c.clients {
			cl.cancel()
		}
	}(gmcli)

	return gmcli, nil
}

// Initializes or updates a client with flags pointing to Control and Catalog for the given mesh.
func (c *CLI) ConfigureMeshClient(mesh *v1alpha1.Mesh) {

	// for CLI 4
	// conf := fmt.Sprintf(`
	// [api]
	// host = "http://control-api.%s.svc:5555/v1.0"
	// [catalog]
	// host = "http://catalog.%s.svc:8080"
	// mesh = "%s"
	// `, mesh.Namespace, mesh.Namespace, mesh.Name)
	// conf = base64.StdEncoding.EncodeToString([]byte(conf))
	// flags := []string{"--base64-config", conf}

	flags := []string{
		fmt.Sprintf("--api.host control.%s.svc:5555", mesh.Namespace),
		fmt.Sprintf("--catalog.host catalog.%s.svc:8080", mesh.Namespace),
		fmt.Sprintf("--catalog.mesh %s", mesh.Name),
	}

	c.configureMeshClient(mesh, flags...)
}

// Initializes or updates a client.
func (c *CLI) configureMeshClient(mesh *v1alpha1.Mesh, flags ...string) {
	c.Lock()
	defer c.Unlock()

	// Close an existing cmds channel if updating
	if cl, ok := c.clients[mesh.Name]; ok {
		cl.cancel()
	}

	c.clients[mesh.Name] = newClient(mesh, flags...)
}

// Closes a client's cmds channels before deleting the client.
func (c *CLI) RemoveMeshClient(name string) {
	c.Lock()
	defer c.Unlock()

	cl, ok := c.clients[name]
	if !ok {
		return
	}

	cl.cancel()
	delete(c.clients, name)
}

// Configures mesh objects given a mesh name, an appsv1.Deployment/StatefulSet name, and a list of corev1.Containers.
// TODO: Remove ingresses if container ports are modified.
// This may require passing a changelog vs containers.
func (c *CLI) ConfigureService(mesh, workload string, annotations map[string]string, containers []corev1.Container) {
	c.RLock()
	defer c.RUnlock()

	cl, ok := c.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	ingresses := make(map[string]int32)
	for _, container := range containers {
		for _, port := range container.Ports {
			if port.Name != "" {
				ingresses[port.Name] = port.ContainerPort
			}
		}
	}

	objects, err := cl.f.Service(workload, annotations, ingresses)
	if err != nil {
		logger.Error(err, "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	cl.controlCmds <- mkApply("domain", objects.Domain)
	cl.controlCmds <- mkApply("listener", objects.Listener)
	for _, cluster := range objects.Clusters {
		cl.controlCmds <- mkApply("cluster", cluster)
	}
	for _, route := range objects.Routes {
		cl.controlCmds <- mkApply("route", route)
	}
	for _, ingress := range objects.Ingresses {
		for _, cluster := range ingress.Clusters {
			cl.controlCmds <- mkApply("cluster", cluster)
		}
		for _, route := range ingress.Routes {
			cl.controlCmds <- mkApply("route", route)
		}
	}
	if objects.HTTPEgresses != nil && len(objects.HTTPEgresses.Routes) > 0 {
		cl.controlCmds <- mkApply("domain", objects.HTTPEgresses.Domain)
		cl.controlCmds <- mkApply("listener", objects.HTTPEgresses.Listener)
		for _, cluster := range objects.HTTPEgresses.Clusters {
			cl.controlCmds <- mkApply("cluster", cluster)
		}
		for _, route := range objects.HTTPEgresses.Routes {
			cl.controlCmds <- mkApply("route", route)
		}
	}
	for _, egress := range objects.TCPEgresses {
		cl.controlCmds <- mkApply("domain", egress.Domain)
		cl.controlCmds <- mkApply("listener", egress.Listener)
		for _, cluster := range egress.Clusters {
			cl.controlCmds <- mkApply("cluster", cluster)
		}
		for _, route := range egress.Routes {
			cl.controlCmds <- mkApply("route", route)
		}
	}
	cl.controlCmds <- mkApply("proxy", objects.Proxy)
	cl.catalogCmds <- mkApply("catalog-service", objects.CatalogService)
	// cl.catalogCmds <- mkApply("catalogservice", objects.CatalogService)
}

// Removes mesh objects given a mesh name, an appsv1.Deployment/StatefulSet name, and a list of corev1.Containers.
func (c *CLI) RemoveService(mesh, workload string, containers []corev1.Container) {
	c.RLock()
	defer c.RUnlock()

	cl, ok := c.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to remove fabric objects for workload", "Mesh", mesh, "Workload", workload)
	}

	ingresses := make(map[int32]struct{})
	for _, container := range containers {
		for _, port := range container.Ports {
			if port.Name != "" {
				ingresses[port.ContainerPort] = struct{}{}
			}
		}
	}

	cl.controlCmds <- mkDelete("domain", workload)
	cl.controlCmds <- mkDelete("listener", workload)
	cl.controlCmds <- mkDelete("proxy", workload)
	cl.controlCmds <- mkDelete("cluster", workload)
	cl.controlCmds <- mkDelete("route", workload)
	for port := range ingresses {
		key := fmt.Sprintf("%s-%d", workload, port)
		cl.controlCmds <- mkDelete("cluster", key)
		cl.controlCmds <- mkDelete("route", key)
	}
	cl.catalogCmds <- mkDelete("catalog-service", workload)
	// cl.catalogCmds <- mkDelete("catalogservice", workload)
}
