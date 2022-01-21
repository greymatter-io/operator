// Package cli executes greymatter CLI commands to configure mesh behavior
// in Control and Catalog APIs in each install namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("cli")
)

// CLI exposes methods for configuring clients that execute greymatter CLI commands.
type CLI struct {
	*sync.RWMutex
	clients     map[string]*client
	mTLSEnabled bool
}

// New returns a new *CLI instance.
// It receives a context for cleaning up goroutines started by the *CLI.
func New(ctx context.Context, mTLSEnabled bool) (*CLI, error) {
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
		RWMutex:     &sync.RWMutex{},
		clients:     make(map[string]*client),
		mTLSEnabled: mTLSEnabled,
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
// for reaching Control and Catalog for the given Mesh CR.
func (c *CLI) ConfigureMeshClient(mesh *v1alpha1.Mesh) {
	conf := mkCLIConfig(
		fmt.Sprintf("http://control.%s.svc.cluster.local:5555", mesh.Spec.InstallNamespace),
		fmt.Sprintf("http://catalog.%s.svc.cluster.local:8080", mesh.Spec.InstallNamespace),
		mesh.Name,
	)
	flags := []string{"--base64-config", conf}

	if err := c.configureMeshClient(mesh, flags...); err != nil {
		logger.Error(err, "failed to configure client", "Mesh", mesh.Name)
	}
}

func mkCLIConfig(apiHost, catalogHost, catalogMesh string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`
	[api]
	host = "%s"
	[catalog]
	host = "%s"
	mesh = "%s"
	`, apiHost, catalogHost, catalogMesh)))
}

func (c *CLI) configureMeshClient(mesh *v1alpha1.Mesh, flags ...string) error {
	c.Lock()
	defer c.Unlock()

	// Close an existing cmds channel if updating
	if cl, ok := c.clients[mesh.Name]; ok {
		logger.Info("Updating mesh client", "Mesh", mesh.Name)
		cl.cancel()
	} else {
		logger.Info("Initializing mesh client", "Mesh", mesh.Name)
	}

	cl, err := newClient(mesh, flags...)
	if err != nil {
		return err
	}

	c.clients[mesh.Name] = cl

	return nil
}

// RemoveMeshClient cleans up a client's goroutines before removing it from the *CLI.
func (c *CLI) RemoveMeshClient(name string) {
	c.Lock()
	defer c.Unlock()

	cl, ok := c.clients[name]
	if !ok {
		return
	}

	logger.Info("Removing mesh client", "Mesh", name)

	cl.cancel()
	delete(c.clients, name)
}

// ConfigureService applies fabric objects that add a workload to the mesh specified
// given the workload's annotations and a list of its corev1.Containers.
func (c *CLI) ConfigureService(mesh, workload string, obj runtime.Object) {
	c.RLock()
	defer c.RUnlock()

	cl, ok := c.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	// TODO: Handle removals of containers and annotations.
	// We'll need to be able to pass previous annotations and containers, or even a diff with actions for what to add/edit/remove.

	objects, err := cl.f.Service(workload, obj)
	if err != nil {
		logger.Error(err, "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	logger.Info("loading fabric objects", "Mesh", mesh, "Workload", workload)

	if workload != "edge" {
		cl.controlCmds <- mkApply(mesh, "domain", objects.Domain)
	}
	cl.controlCmds <- mkApply(mesh, "listener", objects.Listener)
	for _, cluster := range objects.Clusters {
		cl.controlCmds <- mkApply(mesh, "cluster", cluster)
	}
	for _, route := range objects.Routes {
		cl.controlCmds <- mkApply(mesh, "route", route)
	}

	if objects.Ingresses != nil && len(objects.Ingresses.Routes) > 0 {
		for _, cluster := range objects.Ingresses.Clusters {
			cl.controlCmds <- mkApply(mesh, "cluster", cluster)
		}
		for _, route := range objects.Ingresses.Routes {
			cl.controlCmds <- mkApply(mesh, "route", route)
		}
	}

	if objects.HTTPEgresses != nil && len(objects.HTTPEgresses.Routes) > 0 {
		cl.controlCmds <- mkApply(mesh, "domain", objects.HTTPEgresses.Domain)
		cl.controlCmds <- mkApply(mesh, "listener", objects.HTTPEgresses.Listener)
		for _, cluster := range objects.HTTPEgresses.Clusters {
			cl.controlCmds <- mkApply(mesh, "cluster", cluster)
		}
		for _, route := range objects.HTTPEgresses.Routes {
			cl.controlCmds <- mkApply(mesh, "route", route)
		}
	}
	for _, egress := range objects.TCPEgresses {
		cl.controlCmds <- mkApply(mesh, "domain", egress.Domain)
		cl.controlCmds <- mkApply(mesh, "listener", egress.Listener)
		for _, cluster := range egress.Clusters {
			cl.controlCmds <- mkApply(mesh, "cluster", cluster)
		}
		for _, route := range egress.Routes {
			cl.controlCmds <- mkApply(mesh, "route", route)
		}
	}

	if c.mTLSEnabled {
		for _, listener := range objects.LocalEgresses {
			cl.controlCmds <- mkInjectSVID(mesh, fmt.Sprintf("%s.%s", mesh, workload), listener)
		}
	}

	cl.controlCmds <- mkApply(mesh, "proxy", objects.Proxy)

	// Some services should not have a corresponding Catalog service entry
	if len(objects.CatalogService) != 0 {
		cl.catalogCmds <- mkApply(mesh, "catalogservice", objects.CatalogService)
	}
}

// RemoveService removes fabric objects, disconnecting the workload from the mesh specified,
// along with all ingress and egress cluster routes derived from the given annotations and containers.
func (c *CLI) RemoveService(mesh, workload string, obj runtime.Object) {
	c.RLock()
	defer c.RUnlock()

	cl, ok := c.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to remove fabric objects for workload", "Mesh", mesh, "Workload", workload)
	}

	objects, err := cl.f.Service(workload, obj)
	if err != nil {
		logger.Error(err, "failed to configure fabric objects for workload", "Mesh", mesh, "Workload", workload)
		return
	}

	logger.Info("removing fabric objects", "Mesh", mesh, "Workload", workload)

	cl.controlCmds <- mkDelete(mesh, "domain", objects.Domain)
	cl.controlCmds <- mkDelete(mesh, "listener", objects.Listener)
	for _, cluster := range objects.Clusters {
		cl.controlCmds <- mkDelete(mesh, "cluster", cluster)
	}
	for _, route := range objects.Routes {
		cl.controlCmds <- mkDelete(mesh, "route", route)
	}
	if objects.Ingresses != nil && len(objects.Ingresses.Routes) > 0 {
		for _, cluster := range objects.Ingresses.Clusters {
			cl.controlCmds <- mkDelete(mesh, "cluster", cluster)
		}
		for _, route := range objects.Ingresses.Routes {
			cl.controlCmds <- mkDelete(mesh, "route", route)
		}
	}
	if objects.HTTPEgresses != nil && len(objects.HTTPEgresses.Routes) > 0 {
		cl.controlCmds <- mkDelete(mesh, "domain", objects.HTTPEgresses.Domain)
		cl.controlCmds <- mkDelete(mesh, "listener", objects.HTTPEgresses.Listener)
		for _, cluster := range objects.HTTPEgresses.Clusters {
			cl.controlCmds <- mkDelete(mesh, "cluster", cluster)
		}
		for _, route := range objects.HTTPEgresses.Routes {
			cl.controlCmds <- mkDelete(mesh, "route", route)
		}
	}
	for _, egress := range objects.TCPEgresses {
		cl.controlCmds <- mkDelete(mesh, "domain", egress.Domain)
		cl.controlCmds <- mkDelete(mesh, "listener", egress.Listener)
		for _, cluster := range egress.Clusters {
			cl.controlCmds <- mkDelete(mesh, "cluster", cluster)
		}
		for _, route := range egress.Routes {
			cl.controlCmds <- mkDelete(mesh, "route", route)
		}
	}
	cl.controlCmds <- mkDelete(mesh, "proxy", objects.Proxy)
	cl.catalogCmds <- mkDelete(mesh, "catalogservice", objects.CatalogService)
}
