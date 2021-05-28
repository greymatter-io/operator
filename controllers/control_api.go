package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MeshReconciler) mkControlAPI(ctx context.Context, namespace string) error {

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "control", Namespace: namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		// TODO: Create deployment
	} else if err != nil {
		r.Log.Error(err, "failed to get appsv1.Deployment for %s:control", namespace)
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "control", Namespace: namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// TODO: Create service
	} else if err != nil {
		r.Log.Error(err, "failed to get corev1.Service for %s:control", namespace)
	}

	// TODO: Configure mesh objects (send requests to service)
	// Check if objects exist; if not, create them

	return nil
}
