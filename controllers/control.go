package controllers

import (
	"context"
	"fmt"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MeshReconciler) mkControl(ctx context.Context, mesh *installv1.Mesh, gmi gmImages) error {

	// Create RBAC for pod access across cluster, starting with role
	role := &rbacv1.ClusterRole{}
	if err := r.Get(ctx, types.NamespacedName{Name: "control-pods"}, role); err != nil && errors.IsNotFound(err) {
		role = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "control-pods",
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"pods"},
					Verbs:     []string{"list"},
				},
			},
		}
		ctrl.SetControllerReference(mesh, role, r.Scheme)
		r.Log.Info("Creating clusterrole", "Name", "control-pods")
		if err := r.Create(ctx, role); err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create rbacv1.ClusterRole for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get rbacv1.ClusterRole for %s:control", mesh.Namespace))
	}

	account := &corev1.ServiceAccount{}
	if err := r.Get(ctx, types.NamespacedName{Name: "control-pods", Namespace: mesh.Namespace}, account); err != nil && errors.IsNotFound(err) {
		account = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-pods",
				Namespace: mesh.Namespace,
			},
		}
		ctrl.SetControllerReference(mesh, account, r.Scheme)
		r.Log.Info("Creating serviceaccount", "Name", "control-pods", "Namespace", mesh.Namespace)
		if err := r.Create(ctx, account); err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create corev1.ServiceAccount for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get corev1.ServiceAccount for %s:control", mesh.Namespace))
	}

	// TODO: The ClusterRoleBinding should be updated with added subjects per namespace; the CRB is only created once!
	// If another mesh is deployed into another namespace, this will break.
	binding := &rbacv1.ClusterRoleBinding{}
	if err := r.Get(ctx, types.NamespacedName{Name: "control-pods"}, binding); err != nil && errors.IsNotFound(err) {
		binding = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "control-pods",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "control-pods",
					Namespace: mesh.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "control-pods",
			},
		}
		ctrl.SetControllerReference(mesh, binding, r.Scheme)
		r.Log.Info("Creating clusterrolebinding", "Name", "control-pods", "Namespace", mesh.Namespace)
		if err := r.Create(ctx, binding); err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create rbacv1.ClusterRoleBinding for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get rbacv1.ClusterRoleBinding for %s:control", mesh.Namespace))
	}

	// Check if the deployment exists; if not, create a new one
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: "control", Namespace: mesh.Namespace}, deployment); err != nil && errors.IsNotFound(err) {
		deployment = r.mkControlDeployment(mesh, gmi)
		r.Log.Info("Creating deployment", "Name", "control", "Namespace", mesh.Namespace)
		if err := r.Create(ctx, deployment); err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create appsv1.Deployment for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get appsv1.Deployment for %s:control", mesh.Namespace))
	}

	// Check if the service exists; if not, create a new one
	service := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: "control", Namespace: mesh.Namespace}, service); err != nil && errors.IsNotFound(err) {
		service = r.mkControlService(mesh)
		r.Log.Info("Creating service", "Name", "control", "Namespace", mesh.Namespace)
		if err := r.Create(ctx, service); err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to create service for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get service for %s:control", mesh.Namespace))
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
