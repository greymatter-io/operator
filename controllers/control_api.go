package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkControlAPI(ctx context.Context, mesh *installv1.Mesh, namespace string) error {

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "control", Namespace: namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkControlAPIDeployment(mesh)
		r.Log.Info("Creating deployment", "Name", "control", "Namespace", namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, "failed to create appsv1.Deployment for %s:control", namespace)
			return err
		}
	} else if err != nil {
		r.Log.Error(err, "failed to get appsv1.Deployment for %s:control", namespace)
		return err
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "control", Namespace: namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// TODO: Create service
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get corev1.Service for %s:control", namespace))
	}

	// TODO: Configure mesh objects (send requests to service)
	// Check if objects exist; if not, create them

	return nil
}

func (r *MeshReconciler) mkControlAPIDeployment(mesh *installv1.Mesh) *appsv1.Deployment {
	deployment := &appsv1.Deployment{}

	ctrl.SetControllerReference(mesh, deployment, r.Scheme)
	return deployment
}
