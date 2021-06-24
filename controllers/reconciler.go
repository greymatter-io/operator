package controllers

import (
	"context"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type reconciler interface {
	// Returns a string describing the kind of the object being reconciled
	Kind() string
	// Returns the object key used to retrieve the object from the Kubernetes cluster.
	// If the object is cluster-scoped, the 'Namespace' field is an empty string.
	Key() types.NamespacedName
	// Returns an object that implements the client.Object interface (e.g. *appsv1.Deployment, *corev1.Service).
	Object() ctrlclient.Object
	// Builds a new ctrlclient.Object with configuration passed from a *installv1.Mesh.
	Build(*installv1.Mesh) ctrlclient.Object
	// Compares the state of a ctrlclient.Object with its configuration specified by a Mesh object.
	Reconciled(*installv1.Mesh, ctrlclient.Object) (bool, error)
	// Mutates an existing ctrlclient.Object with configuration passed from a *installv1.Mesh
	Mutate(*installv1.Mesh, ctrlclient.Object) ctrlclient.Object
}

func reconcile(ctx context.Context, client *MeshController, r reconciler, mesh *installv1.Mesh) error {
	key := r.Key()

	logger := client.Log.
		WithValues("ReconcileID", ctx.Value(reconcileId("id"))).
		WithValues("Kind", r.Kind()).
		WithValues("Name", key.Name)
	if key.Namespace != "" {
		logger = logger.WithValues("Namespace", key.Namespace)
	}

	obj := r.Object()
	if err := client.Get(ctx, key, obj); err != nil && errors.IsNotFound(err) {
		obj = r.Build(mesh)
		ctrl.SetControllerReference(mesh, obj, client.Scheme)
		if err = client.Create(ctx, obj); err != nil {
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
	if err := client.Update(ctx, obj); err != nil {
		logger.Error(err, "Update Failed")
		return err
	}
	logger.Info("Updated")
	return nil
}
