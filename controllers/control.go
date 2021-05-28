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

func (r *MeshReconciler) mkControl(ctx context.Context, mesh *installv1.Mesh) error {

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "control", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkControlDeployment(mesh)
		r.Log.Info("Creating deployment", "Name", "control", "Namespace", mesh.Namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create appsv1.Deployment for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get appsv1.Deployment for %s: control", mesh.Namespace))
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "control", Namespace: mesh.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		service = r.mkControlService(mesh)
		r.Log.Info("Creating service", "Name", "control", "Namespace", mesh.Namespace)
		err = r.Create(ctx, service)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create service for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get service for %s:control", mesh.Namespace))
	}

	return nil
}

func (r *MeshReconciler) mkControlDeployment(m *installv1.Mesh) *appsv1.Deployment {
	replicas := int32(1)
	labels := map[string]string{
		"deployment":            "control",
		"greymatter":            "fabric",
		"greymatter.io/control": "control",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control",
			Namespace: m.Namespace,
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
					Containers: []corev1.Container{{
						Image:           "docker.greymatter.io/release/gm-control:1.5.3",
						Name:            "control",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							{Name: "GM_CONTROL_API_INSECURE", Value: "true"},
							{Name: "GM_CONTROL_API_SSL", Value: "true"},
							{Name: "GM_CONTROL_API_SSLCERT", Value: "/etc/proxy/tls/sidecar/server.crt"},
							{Name: "GM_CONTROL_API_SSLKEY", Value: "/etc/proxy/tls/sidecar/server.key"},
							{Name: "GM_CONTROL_CONSOLE_LEVEL", Value: "info"},
							{Name: "GM_CONTROL_API_KEY", Value: "xxx"},
							{Name: "GM_CONTROL_API_ZONE_NAME", Value: "zone-default-zone"},
							{Name: "GM_CONTROL_API_HOST", Value: "control-api:5555"},
							{Name: "GM_CONTROL_CMD", Value: "kubernetes"},
							{Name: "GM_CONTROL_XDS_RESOLVE_DNS", Value: "true"},
							{Name: "GM_CONTROL_XDS_ADS_ENABLED", Value: "true"},
							{Name: "GM_CONTROL_KUBERNETES_CLUSTER_LABEL", Value: "greymatter.io"},
							{Name: "GM_CONTROL_KUBERNETES_PORT_NAME", Value: "proxy"},
							{Name: "GM_CONTROL_KUBERNETES_NAMESPACES", Value: m.Namespace},
						},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *MeshReconciler) mkControlService(mesh *installv1.Mesh) *corev1.Service {
	labels := map[string]string{
		"greymatter.io/control": "control",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control",
			Namespace: mesh.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Port: 50000, TargetPort: intstr.FromInt(50000), Protocol: "TCP"},
			},
		},
	}

	ctrl.SetControllerReference(mesh, service, r.Scheme)
	return service
}
