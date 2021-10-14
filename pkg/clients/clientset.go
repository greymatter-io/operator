// Package clients executes greymatter CLI commands to configure mesh capabilities
// in Control and Catalog APIs in each 'system' namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package clients

import (
	"fmt"
	"sync"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.clients")
)

type Clientset struct {
	*sync.RWMutex
	meshClients map[string]*meshClient
}

// Returns *Clientset for storing clients to configure Control and Catalog APIs in the system namespace of each mesh.
func New() (*Clientset, error) {
	v, err := cliVersion()
	if err != nil {
		logger.Error(err, "Failed to initialize greymatter CLI")
		return nil, err
	}

	logger.Info("Using greymatter CLI", "Version", v)

	if err := fabric.Init(); err != nil {
		logger.Error(err, "Failed to initialize Fabric templates")
		return nil, err
	}

	return &Clientset{
		RWMutex:     &sync.RWMutex{},
		meshClients: make(map[string]*meshClient),
	}, nil
}

// Initializes or updates a meshClient.
func (cs *Clientset) ConfigureMeshClient(mesh *v1alpha1.Mesh) {
	cs.Lock()
	defer cs.Unlock()

	// Close an existing cmds channel if updating
	if mc, ok := cs.meshClients[mesh.Name]; ok {
		close(mc.controlCmds)
		close(mc.catalogCmds)
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

	cs.meshClients[mesh.Name] = newMeshClient(mesh, // todo: add --base64-config flag
		fmt.Sprintf("--api.host control.%s.svc:5555", mesh.Namespace),
		fmt.Sprintf("--catalog.host catalog.%s.svc:8080", mesh.Namespace),
		fmt.Sprintf("--catalog.mesh %s", mesh.Name),
	)
}

// Closes a client's cmds channels before deleting the client.
func (cs *Clientset) RemoveMeshClient(name string) {
	cs.Lock()
	defer cs.Unlock()

	mc, ok := cs.meshClients[name]
	if !ok {
		return
	}
	close(mc.controlCmds)
	close(mc.catalogCmds)
	delete(cs.meshClients, name)
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, generates fabric from the stored fabric.ServiceTemplate for each mesh and
// persists each meshobject to the Redis database assigned to each mesh.
func (cs *Clientset) ApplyService(name string, meshes []string, containers []corev1.Container) {
	// TODO: Do not configure local objects for containerPorts with name "proxy"
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, deletes fabric generated for the service from each mesh and
// persists the deletion changes to the Redis database assigned to each mesh.
func (cs *Clientset) RemoveService(name string, meshes []string, containers []corev1.Container) {
	// TODO: Do not attempt to unconfigure local objects for containerPorts with name "proxy"
}
