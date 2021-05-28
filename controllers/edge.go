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

func (r *MeshReconciler) mkEdge(ctx context.Context, mesh *installv1.Mesh) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "edge", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkEdgeDeployment(mesh)
		r.Log.Info("Creating deployment", "Name", "edge", "Namespace", mesh.Namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create deployment for %s:edge", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get deployment for %s:edge", mesh.Namespace))
		return err
	}

	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "edge", Namespace: mesh.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// TODO: Create service
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get corev1.Service for %s:edge", mesh.Namespace))
	}

	return nil
}

func (r *MeshReconciler) mkEdgeDeployment(mesh *installv1.Mesh) *appsv1.Deployment {
	replicas := int32(1)
	labels := map[string]string{
		"greymatter.io/control": "edge",
		"deployment":            "edge",
		"greymatter":            "edge",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "edge",
			Namespace: mesh.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "docker.secret"},
					},
					DNSPolicy:     corev1.DNSClusterFirst,
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{{
						Name:            "edge",
						Image:           "docker.greymatter.io/development/gm-proxy:1.6.0-rc.1",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							{Name: "ENVOY_ADMIN_LOG_PATH", Value: "/dev/stdout"},
							{Name: "PROXY_DYNAMIC", Value: "true"},
							{Name: "XDS_CLUSTER", Value: "edge"},
							{Name: "XDS_HOST", Value: fmt.Sprintf("control.%s.svc", mesh.Namespace)},
							{Name: "XDS_PORT", Value: "50000"},
							{Name: "XDS_ZONE", Value: "zone-default-zone"},
						},
						Ports: []corev1.ContainerPort{
							{ContainerPort: 10808, Name: "proxy", Protocol: "TCP"},
							{ContainerPort: 8081, Name: "metrics", Protocol: "TCP"},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(mesh, deployment, r.Scheme)
	return deployment
}
