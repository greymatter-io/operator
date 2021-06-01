package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/common"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	"github.com/bcmendoza/gm-operator/controllers/meshobjects"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkControlAPI(ctx context.Context, mesh *installv1.Mesh) error {
	if err := r.reconcile(ctx, mesh, gmcore.ControlApi, common.DeploymentReconciler{}); err != nil {
		return err
	}

	if err := r.reconcile(ctx, mesh, gmcore.ControlApi, common.ServiceReconciler{}); err != nil {
		return err
	}

	return nil
}

func (r *MeshReconciler) mkControlAPIDeployment(mesh *installv1.Mesh, gmi gmImages) *appsv1.Deployment {
	replicas := int32(1)

	meshLabels := map[string]string{
		"control-api-version": gmi.ControlAPI,
		"sidecar-version":     gmi.Proxy,
	}

	labels := map[string]string{
		"deployment":            "control-api",
		"greymatter":            "fabric",
		"greymatter.io/control": "control-api",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control-api",
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
					Containers: []corev1.Container{
						{
							Name:  "control-api",
							Image: fmt.Sprintf("docker.greymatter.io/release/gm-control-api:%s", gmi.ControlAPI),
							Env: []corev1.EnvVar{
								{Name: "GM_CONTROL_API_ADDRESS", Value: "0.0.0.0:5555"},
								{Name: "GM_CONTROL_API_DISABLE_VERSION_CHECK", Value: "false"},
								{Name: "GM_CONTROL_API_LOG_LEVEL", Value: "debug"},
								{Name: "GM_CONTROL_API_PERSISTER_TYPE", Value: "null"},
								{Name: "GM_CONTROL_API_EXPERIMENTS", Value: "true"},
								{Name: "GM_CONTROL_API_BASE_URL", Value: "/services/control-api/latest/v1.0/"},
								{Name: "GM_CONTROL_API_USE_TLS", Value: "false"},
								{Name: "GM_CONTROL_API_ORG_KEY", Value: "deciphernow"},
								{Name: "GM_CONTROL_API_ZONE_KEY", Value: "zone-default-zone"},
								{Name: "GM_CONTROL_API_ZONE_NAME", Value: "zone-default-zone"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5555, Name: "http", Protocol: "TCP"},
							},
						},
						{
							Name:  "sidecar",
							Image: fmt.Sprintf("docker.greymatter.io/release/gm-proxy:%s", gmi.Proxy),
							Env: []corev1.EnvVar{
								{Name: "ENVOY_ADMIN_LOG_PATH", Value: "/dev/stdout"},
								{Name: "PROXY_DYNAMIC", Value: "true"},
								{Name: "XDS_CLUSTER", Value: "control-api"},
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

func (r *MeshReconciler) mkControlAPIService(mesh *installv1.Mesh) *corev1.Service {
	labels := map[string]string{
		"greymatter.io/control": "control-api",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control-api",
			Namespace: mesh.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Port: 5555, TargetPort: intstr.FromInt(5555), Protocol: "TCP"},
			},
		},
	}

	ctrl.SetControllerReference(mesh, service, r.Scheme)
	return service
}

func mkMeshObjects(mesh *installv1.Mesh) error {
	addr := fmt.Sprintf("http://control-api.%s.svc.cluster.local:5555", mesh.Namespace)
	client := meshobjects.NewClient(addr)

	return client.MkMeshObjects(
		"zone-default-zone",
		[]string{"control-api:5555", "catalog:9080"},
	)
}
