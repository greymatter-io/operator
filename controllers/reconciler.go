package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
)

type reconciler interface {
	// Returns a string describing the kind of the object being reconciled
	Kind() string
	// Returns the object key used to retrieve the object from the Kubernetes cluster.
	// If the object is cluster-scoped, the 'Namespace' field is an empty string.
	Key() types.NamespacedName
	// Returns an object that implements the client.Object interface (e.g. *appsv1.Deployment).
	Object() client.Object
	// Builds a new client.Object with configuration from a *v1.Mesh and gmcore.Configs.
	Build(*v1.Mesh, gmcore.Configs) client.Object
	// Compares the state of a client.Object with its desired configuration from a *v1.Mesh and gmcore.Configs.
	Reconciled(*v1.Mesh, gmcore.Configs, client.Object) (bool, error)
	// Mutates an existing client.Object with configuration from a *v1.Mesh and gmcore.Configs.
	Mutate(*v1.Mesh, gmcore.Configs, client.Object) client.Object
}

func apply(ctx context.Context, controller *MeshController, mesh *v1.Mesh, configs gmcore.Configs, r reconciler) error {
	key := r.Key()

	logger := controller.Log.
		WithValues("Kind", r.Kind()).
		WithValues("Name", key.Name)
	if key.Namespace != "" {
		logger = logger.WithValues("Namespace", key.Namespace)
	}

	obj := r.Object()
	if err := controller.Get(ctx, key, obj); err != nil && errors.IsNotFound(err) {
		obj = r.Build(mesh, configs)
		ctrl.SetControllerReference(mesh, obj, controller.Scheme)
		if err = controller.Create(ctx, obj); err != nil {
			logger.Error(err, "Create failed")
			return err
		}
		logger.Info("Created")
		return nil
	} else if err != nil {
		logger.Error(err, "Get failed")
		return err
	}

	ok, err := r.Reconciled(mesh, configs, obj)
	if err != nil {
		logger.Error(err, "Reconciled failed")
		return err
	}
	if ok {
		return nil
	}

	obj = r.Mutate(mesh, configs, obj)
	if err := controller.Update(ctx, obj); err != nil {
		logger.Error(err, "Update failed")
		return err
	}
	logger.Info("Updated")

	return nil
}
