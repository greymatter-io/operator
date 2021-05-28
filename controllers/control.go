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

func (r *MeshReconciler) mkControl(ctx context.Context, mesh *installv1.Mesh) error {
	// Make deployment
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "control", Namespace: mesh.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment = r.mkDeploymentForControl(mesh)
		r.Log.Info("Creating deployment", "Name", "control", "Namespace", mesh.Namespace)
		err = r.Create(ctx, deployment)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to create appsv1.Deployment for %s:control", mesh.Namespace))
			return err
		}
	} else if err != nil {
		r.Log.Error(err, fmt.Sprintf("failed to get appsv1.Deployment for %s: control", mesh.Namespace))
	}

	// Make service
	// service := &corev1.ServiceAccountList{}
	// err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-service", mesh.Name), Namespace: mesh.Namespace}, service)

	return nil
}

//mkdeploymentForControl returns a control Deployment object
func (r *MeshReconciler) mkDeploymentForControl(m *installv1.Mesh) *appsv1.Deployment {
	ls := labelsForControl("control")
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control",
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "docker.greymatter.io/release/gm-control:1.5.3",
						Name:            "control",
						ImagePullPolicy: "Always",
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

// func (r *MeshReconciler) mkServiceForControl(m *installv1.Mesh) *corev1.Service {
// 	ls := labelsForControl(m.Name)
// 	svc := &corev1.Service{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      fmt.Sprintf("%s-svc", m.Name),
// 			Namespace: m.Namespace,
// 			Labels:    ls,
// 		},
// 		Spec: corev1.ServiceSpec{
// 			Ports: []corev1.ServicePort{
// 				{Port: 50000, Protocol: "TCP", TargetPort: "gprc"},
// 			},
// 			Selector: ls,
// 		},
// 	}
// 	ctrl.SetControllerReference(m, svc, r.Scheme)
// 	return svc
// }

//labelsForControl returns the labels for selecting the resources
// belongs to the given control CR name.
func labelsForControl(name string) map[string]string {
	return map[string]string{"greymatter.io": name}
}
