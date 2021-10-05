// Package clients executes greymatter CLI commands to configure mesh capabilities
// in Control and Catalog APIs in each 'system' namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package clients

import (
	"fmt"

	"github.com/greymatter-io/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("pkg.clients")
)

type Clientset struct {
	meshes map[string]*client
}

// Returns *Clientset for storing clients to configure Control and Catalog APIs in the system namespace of each mesh.
func New() (*Clientset, error) {
	v, err := version()
	if err != nil {
		return nil, fmt.Errorf("unable to execute greymatter CLI commands: %w", err)
	}

	logger.Info("Using greymatter CLI", "Version", v)

	return &Clientset{make(map[string]*client)}, nil
}

// If the given Mesh is new, initializes a client.
// Generates fabric.ServiceTemplates from the given Mesh, storing them to be used when configuring services.
func (cs *Clientset) ApplyMesh(mesh v1alpha1.Mesh) {
	cl, ok := cs.meshes[mesh.Name]
	if !ok {
		cl = newClient()
	}
	cl.tmpl = mesh.GenerateServiceTemplates()
	cs.meshes[mesh.Name] = cl
}

// Closes a client's cmds channel before deleting the client.
func (cs *Clientset) RemoveMesh(name string) {
	cl, ok := cs.meshes[name]
	if !ok {
		return
	}
	close(cl.cmds)
	delete(cs.meshes, name)
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, generates fabric from the stored fabric.ServiceTemplate for each mesh and
// persists each meshobject to the Redis database assigned to each mesh.
func (cs *Clientset) ApplyService(name string, meshes []string, containers []corev1.Container) {
	// TODO: Do not configure local objects for containerPorts with name "gm-proxy"
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, deletes fabric generated for the service from each mesh and
// persists the deletion changes to the Redis database assigned to each mesh.
func (cs *Clientset) RemoveService(name string, meshes []string, containers []corev1.Container) {
	// TODO: Do not attempt to unconfigure local objects for containerPorts with name "gm-proxy"
}
