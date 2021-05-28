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
)

func (r *MeshReconciler) mkJwt(ctx context.Context, mesh *installv1.Mesh) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "jwt-security", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkJwtSecurityDeployment(mesh)
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

	return nil
}

func (r *MeshReconciler) mkJwtSecurityDeployment(mesh *installv1.Mesh) *appsv1.Deployment {
	replicas := int32(1)
	labels := map[string]string{
		"greymatter.io/control": "jwt-security",
		"deployment":            "jwt-security",
		"greymatter":            "fabric",
	}
	annotations := map[string]string{
		"checksum/config": "e83092f5a470f06cac5b4f8c0c888698e937e0bac1ce00e8c897a863c66215a",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jwt-security",
			Namespace: mesh.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: int32(10),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "docker.secret"},
					},
					DNSPolicy:     corev1.DNSClusterFirst,
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:            "jwt-security",
							Image:           "docker.greymatter.io/release/gm-jwt-security:1.2.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "ENABLE_TLS", Value: "false"},
								{Name: "REDIS_DB", Value: 0},
								{Name: "REDIS_HOST", Value: fmt.Sprintf("jwt-redis.%s.svc", mesh.Namespace)},
								{Name: "REDIS_PORT", Value: 6379},
								{Name: "ZEROLOG_LEVEL", Value: "info"},
								{Name: "HTTPS_PORT", Value: ""},
								// TODO: Add Env that have ValueFrom SecretKey
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
			},
		},
	}
}

// TODO: finish the jwt-security deployment
// TODO: jwt-users ConfigMap
// TODO: jwt-redis Deployment
// TODO: jwt-redis Service
