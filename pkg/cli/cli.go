// Package cli executes greymatter CLI commands to configure mesh behavior
// in Control and Catalog APIs in each install namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package cli

import (
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
}

// Returns *CLI for storing clients used to execute greymatter CLI commands.
func New() (*CLI, error) {
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

	return &CLI{
		RWMutex: &sync.RWMutex{},
		clients: make(map[string]*client),
	}, nil
}

// Initializes or updates a client.
func (cs *CLI) ConfigureMeshClient(mesh *v1alpha1.Mesh) {
	cs.Lock()
	defer cs.Unlock()

	// Close an existing cmds channel if updating
	if cl, ok := cs.clients[mesh.Name]; ok {
		close(cl.controlCmds)
		close(cl.catalogCmds)
	}

	// for CLI 4
	// conf := fmt.Sprintf(`
	// [api]
	// host = "http://control-api.%s.svc:5555/v1.0"
	// [catalog]
	// host = "http://catalog.%s.svc:8080"
	// mesh = "%s"
	// `, mesh.Namespace, mesh.Namespace, mesh.Name)
	// conf = base64.StdEncoding.EncodeToString([]byte(conf))

	cs.clients[mesh.Name] = newClient(mesh,
		// todo: add --base64-config flag
		fmt.Sprintf("--api.host control.%s.svc:5555", mesh.Namespace),
		fmt.Sprintf("--catalog.host catalog.%s.svc:8080", mesh.Namespace),
		fmt.Sprintf("--catalog.mesh %s", mesh.Name),
	)
}

// Closes a client's cmds channels before deleting the client.
func (cs *CLI) RemoveMeshClient(name string) {
	cs.Lock()
	defer cs.Unlock()

	cl, ok := cs.clients[name]
	if !ok {
		return
	}
	close(cl.controlCmds)
	close(cl.catalogCmds)
	delete(cs.clients, name)
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, generates fabric from the stored fabric.ServiceTemplate for each mesh and
// persists each meshobject to the Redis database assigned to each mesh.
func (cs *CLI) ConfigureService(mesh, workload string, containers []corev1.Container) {
	cs.RLock()
	defer cs.RUnlock()

	cl, ok := cs.clients[mesh]
	if !ok {
		logger.Error(fmt.Errorf("unknown mesh"), "failed to configure mesh objects", "Mesh", mesh, "Workload", workload)
	}

	ingresses := make(map[string]int32)
	for _, container := range containers {
		for _, port := range container.Ports {
			if port.Name != "" {
				ingresses[port.Name] = port.ContainerPort
			}
		}
	}

	if len(ingresses) == 0 {
		logger.Error(fmt.Errorf("no named container ports"), "failed to configure mesh objects", "Mesh", mesh, "Workload", workload)
	}

	objects, err := cl.f.Service(workload, ingresses)
	if err != nil {
		logger.Error(err, "failed to generate mesh objects", "Mesh", mesh, "Workload", workload)
	}

	cl.controlCmds <- mkApply("domain", objects.Domain)
	cl.controlCmds <- mkApply("listener", objects.Listener)
	cl.controlCmds <- mkApply("proxy", objects.Proxy)
	cl.controlCmds <- mkApply("cluster", objects.Cluster)
	cl.controlCmds <- mkApply("route", objects.Route)
	for _, ingress := range objects.Ingresses {
		cl.controlCmds <- mkApply("cluster", ingress.Cluster)
		cl.controlCmds <- mkApply("route", ingress.Route)
	}

	logger.Info("configured mesh objects", "Mesh", mesh, "Workload", workload)
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, deletes fabric generated for the service from each mesh and
// persists the deletion changes to the Redis database assigned to each mesh.
func (cs *CLI) RemoveService(mesh, service string, containers []corev1.Container) {

}
