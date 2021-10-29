// Package cli executes greymatter CLI commands to configure mesh behavior
// in Control and Catalog APIs in each install namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package cli

import (
	"context"
	"fmt"
	"sync"

	"cuelang.org/go/cue"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("cli")
)

// CLI exposes methods for configuring clients that execute greymatter CLI commands.
type CLI struct {
	*sync.RWMutex
	clients map[string]*client
}

// New returns a new *CLI instance.
// It receives a context for cleaning up goroutines started by the *CLI.
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

// ConfigureMeshClient initializes or updates a client with flags specifying connection options
// for reaching Control and Catalog for the given mesh and its configuration options.
func (c *CLI) ConfigureMeshClient(mesh *v1alpha1.Mesh, options []cue.Value) {

	// for CLI 4
	// conf := fmt.Sprintf(`
	// [api]
	// host = "http://control.%s.svc.cluster.local:5555/v1.0"
	// [catalog]
	// host = "http://catalog.%s.svc.cluster.local:8080"
	// mesh = "%s"
	// `, mesh.Spec.InstallNamespace, mesh.Spec.InstallNamespace, mesh.Name)
	// conf = base64.StdEncoding.EncodeToString([]byte(conf))
	// flags := []string{"--base64-config", conf}

	flags := []string{
		fmt.Sprintf("--api.host control.%s.svc.cluster.local:5555", mesh.Spec.InstallNamespace),
		fmt.Sprintf("--catalog.host catalog.%s.svc.cluster.local:8080", mesh.Spec.InstallNamespace),
		fmt.Sprintf("--catalog.mesh %s", mesh.Name),
	}

	c.configureMeshClient(mesh, options, flags...)
}

func (c *CLI) configureMeshClient(mesh *v1alpha1.Mesh, options []cue.Value, flags ...string) {
	c.Lock()
	defer c.Unlock()

	// Close an existing cmds channel if updating
	if cl, ok := c.clients[mesh.Name]; ok {
		cl.cancel()
	}

	logger.Info("Initializing fabric objects", "Mesh", mesh.Name)

	c.clients[mesh.Name] = newClient(mesh, options, flags...)
}

// RemoveMeshClient cleans up a client's goroutines before removing it from the *CLI.
func (c *CLI) RemoveMeshClient(name string) {
	c.Lock()
	defer c.Unlock()

	cl, ok := c.clients[name]
	if !ok {
		return
	}

	logger.Info("Removing all fabric objects", "Mesh", name)

	cl.cancel()
	delete(c.clients, name)
}

// ConfigureService applies fabric objects that add a workload to the mesh specified
// given the workload's annotations and a list of its corev1.Containers.
func (c *CLI) ConfigureService(mesh, workload string, annotations map[string]string, containers []corev1.Container) {
	c.RLock()
	defer c.RUnlock()

	cl, ok := c.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	// TODO: Handle removals of containers and annotations.
	// We'll need to be able to pass previous annotations and containers, or even a diff with actions for what to add/edit/remove.

	ingresses := make(map[string]int32)
	for _, container := range containers {
		for _, port := range container.Ports {
			if port.Name != "" || len(container.Ports) == 1 {
				ingresses[port.Name] = port.ContainerPort
			}
		}
	}

	objects, err := cl.f.Service(workload, annotations, ingresses)
	if err != nil {
		logger.Error(err, "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	if workload != "edge" {
		cl.controlCmds <- mkApply("domain", objects.Domain)
	}
	cl.controlCmds <- mkApply("listener", objects.Listener)
	for _, cluster := range objects.Clusters {
		cl.controlCmds <- mkApply("cluster", cluster)
	}
	for _, route := range objects.Routes {
		cl.controlCmds <- mkApply("route", route)
	}
	if objects.Ingresses != nil && len(objects.Ingresses.Routes) > 0 {
		logger.Info("configuring ingresses", "Mesh", mesh, "Workload", workload)
		for _, cluster := range objects.Ingresses.Clusters {
			cl.controlCmds <- mkApply("cluster", cluster)
		}
		for _, route := range objects.Ingresses.Routes {
			cl.controlCmds <- mkApply("route", route)
		}
	}
	if objects.HTTPEgresses != nil && len(objects.HTTPEgresses.Routes) > 0 {
		logger.Info("configuring HTTP egresses", "Mesh", mesh, "Workload", workload)
		cl.controlCmds <- mkApply("domain", objects.HTTPEgresses.Domain)
		cl.controlCmds <- mkApply("listener", objects.HTTPEgresses.Listener)
		for _, cluster := range objects.HTTPEgresses.Clusters {
			cl.controlCmds <- mkApply("cluster", cluster)
		}
		for _, route := range objects.HTTPEgresses.Routes {
			cl.controlCmds <- mkApply("route", route)
		}
	}
	logger.Info("configuring TCP egresses", "Mesh", mesh, "Workload", workload)
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

// RemoveService removes fabric objects, disconnecting the workload from the mesh specified,
// along with all ingress and egress cluster routes derived from the given annotations and containers.
func (c *CLI) RemoveService(mesh, workload string, annotations map[string]string, containers []corev1.Container) {
	c.RLock()
	defer c.RUnlock()

	cl, ok := c.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to remove fabric objects for workload", "Mesh", mesh, "Workload", workload)
	}

	ingresses := make(map[string]int32)
	for _, container := range containers {
		for _, port := range container.Ports {
			if port.Name != "" || len(container.Ports) == 1 {
				ingresses[port.Name] = port.ContainerPort
			}
		}
	}

	objects, err := cl.f.Service(workload, annotations, ingresses)
	if err != nil {
		logger.Error(err, "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	logger.Info("removing fabric objects", "Mesh", mesh, "Workload", workload)
	cl.controlCmds <- mkDelete("domain", objects.Domain)
	cl.controlCmds <- mkDelete("listener", objects.Listener)
	for _, cluster := range objects.Clusters {
		cl.controlCmds <- mkDelete("cluster", cluster)
	}
	for _, route := range objects.Routes {
		cl.controlCmds <- mkDelete("route", route)
	}
	if objects.Ingresses != nil && len(objects.Ingresses.Routes) > 0 {
		logger.Info("removing ingresses", "Mesh", mesh, "Workload", workload)
		for _, cluster := range objects.Ingresses.Clusters {
			cl.controlCmds <- mkDelete("cluster", cluster)
		}
		for _, route := range objects.Ingresses.Routes {
			cl.controlCmds <- mkDelete("route", route)
		}
	}
	if objects.HTTPEgresses != nil && len(objects.HTTPEgresses.Routes) > 0 {
		logger.Info("removing HTTP egresses", "Mesh", mesh, "Workload", workload)
		cl.controlCmds <- mkDelete("domain", objects.HTTPEgresses.Domain)
		cl.controlCmds <- mkDelete("listener", objects.HTTPEgresses.Listener)
		for _, cluster := range objects.HTTPEgresses.Clusters {
			cl.controlCmds <- mkDelete("cluster", cluster)
		}
		for _, route := range objects.HTTPEgresses.Routes {
			cl.controlCmds <- mkDelete("route", route)
		}
	}
	logger.Info("removing TCP egresses", "Mesh", mesh, "Workload", workload)
	for _, egress := range objects.TCPEgresses {
		cl.controlCmds <- mkDelete("domain", egress.Domain)
		cl.controlCmds <- mkDelete("listener", egress.Listener)
		for _, cluster := range egress.Clusters {
			cl.controlCmds <- mkDelete("cluster", cluster)
		}
		for _, route := range egress.Routes {
			cl.controlCmds <- mkDelete("route", route)
		}
	}
	cl.controlCmds <- mkDelete("proxy", objects.Proxy)
	cl.catalogCmds <- mkDelete("catalog-service", objects.CatalogService)
	// cl.catalogCmds <- mkDelete("catalogservice", objects.CatalogService)
}
