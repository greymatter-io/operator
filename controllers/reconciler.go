package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/bcmendoza/gm-operator/api/v1"
)

type reconciler interface {
	// Returns a string describing the kind of the object being reconciled
	Kind() string
	// Returns the object key used to retrieve the object from the Kubernetes cluster.
	// If the object is cluster-scoped, the 'Namespace' field is an empty string.
	Key() types.NamespacedName
	// Returns an object that implements the client.Object interface (e.g. *appsv1.Deployment).
	Object() client.Object
	// Builds a new client.Object with configuration passed from a *v1.Mesh.
	Build(*v1.Mesh) client.Object
	// Compares the state of a client.Object with its configuration specified by a Mesh object.
	Reconciled(*v1.Mesh, client.Object) (bool, error)
	// Mutates an existing client.Object with configuration passed from a *v1.Mesh
	Mutate(*v1.Mesh, client.Object) client.Object
}

func apply(ctx context.Context, controller *MeshController, mesh *v1.Mesh, r reconciler) error {
	key := r.Key()

	logger := controller.Log.
		WithValues("ReconcileID", ctx.Value(struct{}{})).
		WithValues("Kind", r.Kind()).
		WithValues("Name", key.Name)
	if key.Namespace != "" {
		logger = logger.WithValues("Namespace", key.Namespace)
	}

	obj := r.Object()
	if err := controller.Get(ctx, key, obj); err != nil && errors.IsNotFound(err) {
		obj = r.Build(mesh)
		ctrl.SetControllerReference(mesh, obj, controller.Scheme)
		if err = controller.Create(ctx, obj); err != nil {
			logger.Error(err, "Create Failed")
			return err
		}
		logger.Info("Created")
		return nil
	} else if err != nil {
		logger.Error(err, "Get Failed")
		return err
	}

	ok, err := r.Reconciled(mesh, obj)
	if err != nil {
		logger.Error(err, "Eval Failed")
		return err
	}
	if ok {
		return nil
	}

	obj = r.Mutate(mesh, obj)
	if err := controller.Update(ctx, obj); err != nil {
		logger.Error(err, "Update Failed")
		return err
	}
	logger.Info("Updated")
	return nil
}
