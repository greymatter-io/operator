package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MeshReconciler) mkEdge(ctx context.Context) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "edge", Namespace: namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		dep := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edge",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: 1,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"greymatter.io/control": "edge",
						"deployment":            "edge",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"greymatter.io/control": "edge",
							"deployment":            "edge",
							"greymatter":            "edge",
						},
					},
					Spec: corev1.PodSpec{
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
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "docker.secret"}
						},
					},
				},
			},
		}
	} else if err != nil {
		r.log.Error(err, "failed to get appsv1.Deployment for %s:edge", namespace)
		// TODO: Add exit
	}

	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: "edge", Namespace: namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// TODO: Create service
	} else if err != nil {
		r.Log.Error(err, "failed to get corev1.Service for %s:edge", namespace)
	}
	// Make service
	return nil
}
