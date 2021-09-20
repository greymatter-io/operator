// Package store uses functions from package meshobjects to persist Control API and Catalog objects
// to the gm-operator's Redis databases that are used by each mesh in the Kubernetes cluster.
// It enables Mesh CR specifications to define how a mesh should be internally configured.
package store

import (
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/meshobjects"
	corev1 "k8s.io/api/core/v1"
)

// References a db for each mesh
type Data struct {
	each map[string]db
}

type db struct {
	id        int
	templates meshobjects.ServiceTemplates
}

// Ensures a connection to Redis in the gm-operator namespace and returns *Data for
// tracking Redis database assignments for each mesh and how meshobjects in each should be configured.
func New() (*Data, error) {
	return &Data{make(map[string]db)}, nil
}

// Assigns a Redis database to the given Mesh CR (if not assigned), creates meshobject.Edge objects and
// meshobjects.ServiceTemplates that match the given Mesh CR specifications, persists the Edge objects,
// and stores the ServiceTemplates with the Redis database ID in Data.each[mesh.ObjectMeta.Name].
func (s *Data) ApplyMesh(mesh v1alpha1.Mesh) {
}

// Deletes all data from the Redis database assigned to the given Mesh CR name before unassigning the database.
func (s *Data) RemoveMesh(name string) {
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, generates meshobjects from the stored meshobjects.ServiceTemplate for each mesh and
// persists each meshobject to the Redis database assigned to each mesh.
func (s *Data) ApplyService(name string, meshes []string, containers []corev1.Container) {
	// TODO: Do not configure local objects for containers with name prefix "greymatter-dp_"
}

// Given the name of an appsv1.Deployment/StatefulSet, a list of its meshes from its `greymatter.io/mesh` label, and
// a list of corev1.Containers, deletes meshobjects generated for the service from each mesh and
// persists the deletion changes to the Redis database assigned to each mesh.
func (s *Data) RemoveService(name string, meshes []string, containers []corev1.Container) {
	// TODO: Do not attempt to unconfigure local objects for containers with name prefix "greymatter-dp_"
}
