package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkControlAPI(ctx context.Context, mesh *installv1.Mesh) error {

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "control-api", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkControlAPIDeployment(mesh)
		r.Log.Info("Creating deployment", "Name", "control-api", "Namespace", mesh.Namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create appsv1.Deployment for %s:control-api", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get appsv1.Deployment for %s:control-api", mesh.Namespace))
		return err
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "control-api", Namespace: mesh.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// TODO: Create service
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get corev1.Service for %s:control-api", mesh.Namespace))
	}

	// TODO: Configure mesh objects (send requests to service)
	// Check if objects exist; if not, create them

	return nil
}

func (r *MeshReconciler) mkControlAPIDeployment(mesh *installv1.Mesh) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control-api",
			Namespace: mesh.Namespace,
		},
	}

	ctrl.SetControllerReference(mesh, deployment, r.Scheme)
	return deployment
}
