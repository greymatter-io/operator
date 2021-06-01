package controllers

import (
	"context"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type reconciler interface {
	Object() client.Object
	Build(*installv1.Mesh, gmcore.SvcName) (client.Object, error)
	Reconciled(*installv1.Mesh, client.Object) (bool, error)
}

func (r *MeshReconciler) reconcile(ctx context.Context, mesh *installv1.Mesh, svc gmcore.SvcName, rec reconciler) error {
	obj := rec.Object()

	logger := r.Log.
		WithName(obj.GetObjectKind().GroupVersionKind().String()).
		WithValues("Name", string(svc)).
		WithValues("Namespace", mesh.Namespace)

	key := types.NamespacedName{Name: string(svc), Namespace: mesh.Namespace}

	if err := r.Get(ctx, key, obj); err != nil && errors.IsNotFound(err) {
		obj, err = rec.Build(mesh, svc)
		if err != nil {
			logger.Error(err, "Failed to build struct")
			return err
		}
		logger.Info("Attempting to Create")
		if err = r.Create(ctx, obj); err != nil {
			logger.Error(err, "Failed to Create")
			return err
		}
		logger.Info("Created")
		return nil
	} else if err != nil {
		logger.Error(err, "Failed to Get")
		return err
	}

	ok, err := rec.Reconciled(mesh, obj)
	if err != nil {
		logger.Error(err, "Failed to determine reconciliation status")
		return err
	}
	if ok {
		return nil
	}

	logger.Info("Attempting to Update")
	if err := r.Update(ctx, obj); err != nil {
		logger.Error(err, "Failed to Update")
		return err
	}

	return nil
}
