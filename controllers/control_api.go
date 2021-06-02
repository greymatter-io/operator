package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/common"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	"github.com/bcmendoza/gm-operator/controllers/meshobjects"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MeshReconciler) mkControlAPI(ctx context.Context, mesh *installv1.Mesh) error {
	key := types.NamespacedName{
		Name:      string(gmcore.ControlApi),
		Namespace: mesh.Namespace,
	}

	if err := r.reconcile(ctx, mesh, common.DeploymentReconciler{ObjectKey: key}); err != nil {
		return err
	}
	if err := r.reconcile(ctx, mesh, common.ServiceReconciler{ObjectKey: key}); err != nil {
		return err
	}

	return nil
}

func mkMeshObjects(mesh *installv1.Mesh) error {
	addr := fmt.Sprintf("http://control-api.%s.svc.cluster.local:5555", mesh.Namespace)
	client := meshobjects.NewClient(addr)

	return client.MkMeshObjects(
		"zone-default-zone",
		[]string{"control-api:5555", "catalog:9080"},
	)
}
