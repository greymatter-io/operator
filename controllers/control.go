package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/common"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkControl(ctx context.Context, mesh *installv1.Mesh) error {

	// Create RBAC for pod access across cluster
	rbacName := "control-pods"
	if err := r.reconcile(ctx, mesh, common.ClusterRoleReconciler{Name: rbacName}); err != nil {
		return err
	}
	sarKey := types.NamespacedName{Name: rbacName, Namespace: mesh.Namespace}
	if err := r.reconcile(ctx, mesh, common.ServiceAccountReconciler{ObjectKey: sarKey}); err != nil {
		return err
	}

	// TODO: The ClusterRoleBinding should be updated with added subjects per namespace; the CRB is only created once!
	// If another mesh is deployed into another namespace, this will break.
	if err := r.reconcile(ctx, mesh, common.ClusterRoleBindingReconciler{Name: rbacName}); err != nil {
		return err
	}

	key := types.NamespacedName{
		Name:      string(gmcore.Control),
		Namespace: mesh.Namespace,
	}

	if err := r.reconcile(ctx, mesh, common.DeploymentReconciler{ObjectKey: key}); err != nil {
		return err
	}
	if err := r.reconcile(ctx, mesh, common.ServiceReconciler{ObjectKey: key}); err != nil {
		return err
	}

	return nil
}

func (r *MeshReconciler) mkControlDeployment(mesh *installv1.Mesh, gmi gmImages) *appsv1.Deployment {
	replicas := int32(1)

	meshLabels := map[string]string{
		"control-version": gmi.Control,
	}

	labels := map[string]string{
		"deployment":            "control",
		"greymatter":            "fabric",
		"greymatter.io/control": "control",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control",
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
					ServiceAccountName: "control-pods",
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: mesh.Spec.ImagePullSecret},
					},
					Containers: []corev1.Container{{
						Name:  "control",
						Image: fmt.Sprintf("docker.greymatter.io/release/gm-control:%s", gmi.Control),
						Env: []corev1.EnvVar{
							{Name: "GM_CONTROL_API_INSECURE", Value: "true"},
							{Name: "GM_CONTROL_API_SSL", Value: "false"},
							{Name: "GM_CONTROL_API_SSLCERT", Value: "/etc/proxy/tls/sidecar/server.crt"},
							{Name: "GM_CONTROL_API_SSLKEY", Value: "/etc/proxy/tls/sidecar/server.key"},
							{Name: "GM_CONTROL_CONSOLE_LEVEL", Value: "debug"},
							{Name: "GM_CONTROL_API_KEY", Value: "xxx"},
							{Name: "GM_CONTROL_API_ZONE_NAME", Value: "zone-default-zone"},
							{Name: "GM_CONTROL_API_HOST", Value: "control-api:5555"},
							{Name: "GM_CONTROL_CMD", Value: "kubernetes"},
							{Name: "GM_CONTROL_XDS_RESOLVE_DNS", Value: "true"},
							{Name: "GM_CONTROL_XDS_ADS_ENABLED", Value: "true"},
							{Name: "GM_CONTROL_KUBERNETES_CLUSTER_LABEL", Value: "greymatter.io/control"},
							{Name: "GM_CONTROL_KUBERNETES_PORT_NAME", Value: "proxy"},
							{Name: "GM_CONTROL_KUBERNETES_NAMESPACES", Value: mesh.Namespace},
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{ContainerPort: 50000, Name: "grpc", Protocol: "TCP"},
						},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(mesh, dep, r.Scheme)
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
