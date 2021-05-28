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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkCatalog(ctx context.Context, mesh *installv1.Mesh, gmi gmImages) error {

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "catalog", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkCatalogAPIDeployment(mesh, gmi)
		r.Log.Info("Creating deployment", "Name", "catalog", "Namespace", mesh.Namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create deployment for %s:catalog", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get deployment for %s:catalog", mesh.Namespace))
		return err
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "catalog", Namespace: mesh.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		service = r.mkCatalogService(mesh)
		r.Log.Info("Creating service", "Name", "catalog", "Namespace", mesh.Namespace)
		err = r.Create(ctx, service)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create service for %s:catalog", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get service for %s:catalog", mesh.Namespace))
	}

	return nil
}

func (r *MeshReconciler) mkCatalogAPIDeployment(mesh *installv1.Mesh, gmi gmImages) *appsv1.Deployment {
	replicas := int32(1)
	labels := map[string]string{
		"deployment":            "catalog",
		"greymatter":            "fabric",
		"greymatter.io/control": "catalog",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "catalog",
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
					Containers: []corev1.Container{
						{
							Name:  "catalog",
							Image: fmt.Sprintf("docker.greymatter.io/release/gm-catalog:%s", gmi.Catalog),
							Env: []corev1.EnvVar{
								{Name: "CONTROL_SERVER_0_ADDRESS", Value: fmt.Sprintf("control.%s.svc.cluster.local:50000", mesh.Namespace)},
								{Name: "CONTROL_SERVER_0_REQUEST_CLUSTER_NAME", Value: "edge"},
								{Name: "CONTROL_SERVER_0_ZONE_NAME", Value: "zone-default-zone"},
								{Name: "PORT", Value: "9080"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 9080, Name: "http", Protocol: "TCP"},
							},
						},
						{
							Name:  "sidecar",
							Image: fmt.Sprintf("docker.greymatter.io/release/%s", gmi.Proxy),
							Env: []corev1.EnvVar{
								{Name: "ENVOY_ADMIN_LOG_PATH", Value: "/dev/stdout"},
								{Name: "PROXY_DYNAMIC", Value: "true"},
								{Name: "XDS_CLUSTER", Value: "catalog"},
								{Name: "XDS_HOST", Value: fmt.Sprintf("control.%s.svc", mesh.Namespace)},
								{Name: "XDS_PORT", Value: "50000"},
								{Name: "XDS_ZONE", Value: "zone-default-zone"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 10808, Name: "proxy", Protocol: "TCP"},
								{ContainerPort: 8081, Name: "metrics", Protocol: "TCP"},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	ctrl.SetControllerReference(mesh, deployment, r.Scheme)
	return deployment
}

func (r *MeshReconciler) mkCatalogService(mesh *installv1.Mesh) *corev1.Service {
	labels := map[string]string{
		"greymatter.io/control": "catalog",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "catalog",
			Namespace: mesh.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Port: 9080, TargetPort: intstr.FromInt(9080), Protocol: "TCP"},
			},
		},
	}

	ctrl.SetControllerReference(mesh, service, r.Scheme)
	return service
}
