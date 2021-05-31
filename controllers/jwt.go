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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkJwtSecurity(ctx context.Context, mesh *installv1.Mesh, gmi gmImages) error {

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "jwt-security", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkJwtSecurityDeployment(mesh, gmi)
		r.Log.Info("Creating deployment", "Name", "jwt-security", "Namespace", mesh.Namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create deployment for %s:jwt-security", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get deployment for %s:jwt-security", mesh.Namespace))
		return err
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "jwt-security", Namespace: mesh.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		service = r.mkJwtSecurityService(mesh)
		r.Log.Info("Creating service", "Name", "jwt-security", "Namespace", mesh.Namespace)
		err = r.Create(ctx, service)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create service for %s:jwt-security", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get service for %s:jwt-security", mesh.Namespace))
	}

	return nil
}

func (r *MeshReconciler) mkJwtSecurityDeployment(mesh *installv1.Mesh, gmi gmImages) *appsv1.Deployment {
	replicas := int32(1)

	meshLabels := map[string]string{
		"jwt-security-version": gmi.JwtSecurity,
		"proxy-version":        gmi.Proxy,
	}

	labels := map[string]string{
		"deployment":            "jwt-security",
		"greymatter":            "fabric",
		"greymatter.io/control": "jwt-security",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jwt-security",
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
						{Name: "docker.secret"},
					},
					// 		Containers: []corev1.Container{
					// 			{
					// 				Name:            "jwt-security",
					// 				Image:           "docker.greymatter.io/release/gm-jwt-security:1.2.0",
					// 				ImagePullPolicy: corev1.PullIfNotPresent,
					// 				Env: []corev1.EnvVar{
					// 					{Name: "ENABLE_TLS", Value: "false"},
					// 					{Name: "REDIS_DB", Value: 0},
					// 					{Name: "REDIS_HOST", Value: fmt.Sprintf("jwt-redis.%s.svc", mesh.Namespace)},
					// 					{Name: "REDIS_PORT", Value: 6379},
					// 					{Name: "ZEROLOG_LEVEL", Value: "info"},
					// 					{Name: "HTTPS_PORT", Value: ""},
					// 					// TODO: Add Env that have ValueFrom SecretKey
					// 				},
					// 				Resources: corev1.ResourceRequirements{
					// 					Limits: corev1.ResourceList{
					// 						corev1.ResourceCPU:    resource.MustParse("200m"),
					// 						corev1.ResourceMemory: resource.MustParse("512Mi"),
					// 					},
					// 					Requests: corev1.ResourceList{
					// 						corev1.ResourceCPU:    resource.MustParse("100m"),
					// 						corev1.ResourceMemory: resource.MustParse("64Mi"),
					// 			},
					// 		},
					// 	},
					// },
				},
			},
		},
	}

	ctrl.SetControllerReference(mesh, deployment, r.Scheme)
	return deployment
}

func (r *MeshReconciler) mkJwtSecurityService(mesh *installv1.Mesh) *corev1.Service {
	labels := map[string]string{
		"greymatter.io/control": "jwt-security",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jwt-security",
			Namespace: mesh.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Port: 3000, TargetPort: intstr.FromInt(3000), Protocol: "TCP"},
			},
		},
	}

	ctrl.SetControllerReference(mesh, service, r.Scheme)
	return service
}

// TODO: jwt-users ConfigMap
// TODO: jwt-redis Deployment
// TODO: jwt-redis Service
