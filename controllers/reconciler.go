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
	// Returns the object key used to retrieve the object from the Kubernetes cluster.
	// If the object is cluster-scoped, the 'Namespace' field is an empty string.
	Key() types.NamespacedName
	// Returns an object that implements the client.Object interface (e.g. *appsv1.Deployment, *corev1.Service).
	Object() client.Object
	// Builds a new client.Object with configuration passed from a *installv1.Mesh.
	Build(*installv1.Mesh) (client.Object, error)
	// Compares the state of a client.Object with its configuration specified by a Mesh object.
	Reconciled(*installv1.Mesh, client.Object) (bool, error)
}

func (r *MeshReconciler) reconcile(ctx context.Context, mesh *installv1.Mesh, rec reconciler) (bool, error) {
	obj := rec.Object()
	key := rec.Key()

	logger := r.Log.WithValues("Name", key.Name)
	if key.Namespace != "" {
		logger = logger.WithValues("Namespace", key.Namespace)
	}

	if err := r.Get(ctx, key, obj); err != nil && errors.IsNotFound(err) {
		obj, err = rec.Build(mesh)
		if err != nil {
			logger.Error(err, "Failed to build struct")
			return false, err
		}
		ctrl.SetControllerReference(mesh, obj, r.Scheme)
		logger.Info("Attempting to Create")
		if err = r.Create(ctx, obj); err != nil {
			logger.Error(err, "Failed to Create")
			return false, err
		}
		logger.Info("Created")
		return true, nil
	} else if err != nil {
		logger.Error(err, "Failed to Get")
		return false, err
	}

	ok, err := rec.Reconciled(mesh, obj)
	if err != nil {
		logger.Error(err, "Failed to determine reconciliation status")
		return false, err
	}
	if ok {
		return false, nil
	}

	logger.Info("Attempting to Update")
	if err := r.Update(ctx, obj); err != nil {
		logger.Error(err, "Failed to Update")
		return false, err
	}
	return true, nil
}
