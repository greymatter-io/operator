package controllers

import (
	"context"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type reconciler interface {
	// Returns a string describing the kind of the object being reconciled
	Kind() string
	// Returns the object key used to retrieve the object from the Kubernetes cluster.
	// If the object is cluster-scoped, the 'Namespace' field is an empty string.
	Key() types.NamespacedName
	// Returns an object that implements the client.Object interface (e.g. *appsv1.Deployment, *corev1.Service).
	Object() client.Object
	// Builds a new client.Object with configuration passed from a *installv1.Mesh.
	Build(*installv1.Mesh) client.Object
	// Compares the state of a client.Object with its configuration specified by a Mesh object.
	Reconciled(*installv1.Mesh, client.Object) (bool, error)
	// Mutates an existing client.Object with configuration passed from a *installv1.Mesh
	Mutate(*installv1.Mesh, client.Object) client.Object
}

func (r *MeshReconciler) reconcile(ctx context.Context, mesh *installv1.Mesh, rec reconciler) error {
	key := rec.Key()

	logger := r.Log.
		WithValues("ReconcileID", ctx.Value(reconcileId("id"))).
		WithValues("Kind", rec.Kind()).
		WithValues("Name", key.Name)
	if key.Namespace != "" {
		logger = logger.WithValues("Namespace", key.Namespace)
	}

	obj := rec.Object()
	if err := r.Get(ctx, key, obj); err != nil && errors.IsNotFound(err) {
		obj = rec.Build(mesh)
		ctrl.SetControllerReference(mesh, obj, r.Scheme)
		if err = r.Create(ctx, obj); err != nil {
			logger.Error(err, "Create Failed")
			return err
		}
		logger.Info("Created")
		return nil
	} else if err != nil {
		logger.Error(err, "Get Failed")
		return err
	}

	ok, err := rec.Reconciled(mesh, obj)
	if err != nil {
		logger.Error(err, "Eval Failed")
		return err
	}
	if ok {
		return nil
	}

	obj = rec.Mutate(mesh, obj)
	if err := r.Update(ctx, obj); err != nil {
		logger.Error(err, "Update Failed")
		return err
	}
	logger.Info("Updated")
	return nil
}
