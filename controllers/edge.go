package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkEdge(ctx context.Context, mesh *installv1.Mesh, gmi gmImages) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "edge", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkEdgeDeployment(mesh, gmi)
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
		service = r.mkEdgeService(mesh)
		r.Log.Info("Creating service", "Name", "edge", "Namespace", mesh.Namespace)
		err = r.Create(ctx, service)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create service for %s:edge", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get service for %s:edge", mesh.Namespace))
		return err
	}

	return nil
}

func (r *MeshReconciler) mkEdgeDeployment(mesh *installv1.Mesh, gmi gmImages) *appsv1.Deployment {
	replicas := int32(1)

	meshLabels := map[string]string{
		"sidecar-version": gmi.Proxy,
	}

	labels := map[string]string{
		"greymatter.io/control": "edge",
		"deployment":            "edge",
		"greymatter":            "edge",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "edge",
			Namespace: mesh.Namespace,
			Labels:    meshLabels,
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
						{Name: mesh.Spec.ImagePullSecret},
					},
					Containers: []corev1.Container{{
						Name:            "edge",
						Image:           fmt.Sprintf("docker.greymatter.io/release/gm-proxy:%s", gmi.Proxy),
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
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(mesh, deployment, r.Scheme)
	return deployment
}

func (r *MeshReconciler) mkEdgeService(mesh *installv1.Mesh) *corev1.Service {
	labels := map[string]string{
		"greymatter.io/control": "edge",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "edge",
			Namespace: mesh.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "proxy", Port: 10808, Protocol: "TCP"},
				{Name: "metrics", Port: 8081, Protocol: "TCP"},
			},
			Selector: labels,
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	ctrl.SetControllerReference(mesh, service, r.Scheme)
	return service
}
