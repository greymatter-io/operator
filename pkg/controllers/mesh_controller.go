/*
Copyright Decipher Technology Studios 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/greymatter.io/operator/pkg/api/v1"
	"github.com/greymatter.io/operator/pkg/gmcore"
	"github.com/greymatter.io/operator/pkg/reconcilers"
)

// MeshController reconciles a Mesh object
type MeshController struct {
	client.Client
	Scheme *runtime.Scheme
	Logger logr.Logger
}

func NewMeshController(client client.Client, scheme *runtime.Scheme) *MeshController {
	return &MeshController{
		Client: client,
		Scheme: scheme,
		Logger: ctrl.Log.WithName("controllers").WithName("Mesh"),
	}
}

/*
	Specify RBAC cluster role rules to generate when running `make manifests`.
	This updates /manifests/rbac/role.yaml
*/

//+kubebuilder:rbac:groups=greymatter.io,resources=meshes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=greymatter.io,resources=meshes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=greymatter.io,resources=meshes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts;secrets;pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It compares the state specified by the Mesh object against the actual state
// of the namespace and creates/updates all deployments, services, roles, ingresses,
// mesh objects, etc. to the desired Mesh object configuration.
//
// For more details, check Reconcile and its result:
// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (controller *MeshController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := controller.Logger.WithName(req.NamespacedName.Name)

	// Fetch the Mesh object
	mesh := &v1.Mesh{}
	if err := controller.Get(ctx, req.NamespacedName, mesh); err != nil {
		if errors.IsNotFound(err) {
			// Mesh object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Mesh")
		return ctrl.Result{}, err
	}

	// For now add defaults here.
	// Later this can be added to a mutating webhook for v1.Mesh.
	if mesh.Spec.ImagePullSecret == "" {
		mesh.Spec.ImagePullSecret = "docker.secret"
	}
	if mesh.Spec.Version == "" {
		mesh.Spec.Version = "latest"
	}

	configs := gmcore.GetConfigs(mesh.Spec.Version)

	// Get the secret within this gm-operator namespace and re-create it in the mesh namesapce
	key := types.NamespacedName{Name: mesh.Spec.ImagePullSecret, Namespace: "gm-operator"}
	operatorSecret := &corev1.Secret{}
	if err := controller.Get(ctx, key, operatorSecret); err != nil && errors.IsNotFound(err) {
		// If the secret does not exist, return and don't requeue.
		// No resources will be created without a valid ImagePullSecret.
		logger.Error(err, fmt.Sprintf("Failed to get secret '%s' in gm-operator namespace", mesh.Spec.ImagePullSecret))
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Secret{
		ObjectKey: types.NamespacedName{Name: mesh.Spec.ImagePullSecret, Namespace: mesh.Namespace},
		ObjectLiteral: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mesh.Spec.ImagePullSecret,
				Namespace: mesh.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name":       mesh.Spec.ImagePullSecret,
					"app.kubernetes.io/part-of":    "greymatter",
					"app.kubernetes.io/managed-by": "gm-operator",
					"app.kubernetes.io/created-by": "gm-operator",
				},
			},
			Type: operatorSecret.Type,
			Data: operatorSecret.Data,
		},
	}); err != nil {
		return ctrl.Result{}, err
	}

	// Control API
	key = types.NamespacedName{Name: string(gmcore.ControlApi), Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.ControlApi, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.ControlApi, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	go reconcileMesh(controller, mesh, logger)

	// Catalog
	key = types.NamespacedName{Name: string(gmcore.Catalog), Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.Catalog, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.Catalog, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	go reconcileCatalog(controller, mesh, logger)

	// Dashboard
	key = types.NamespacedName{Name: string(gmcore.Dashboard), Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.Dashboard, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.Dashboard, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}

	// JWT Security
	if len(mesh.Spec.Users) > 0 {
		users, err := json.Marshal(mesh.Spec.Users)
		if err != nil {
			return ctrl.Result{}, err
		}
		smKey := types.NamespacedName{Name: "jwt-users", Namespace: mesh.Namespace}
		if err := apply(ctx, controller, mesh, configs, reconcilers.ConfigMap{
			ObjectKey: smKey,
			Data:      map[string]string{"users.json": string(users)},
		}); err != nil {
			return ctrl.Result{}, err
		}
	}
	key = types.NamespacedName{Name: string(gmcore.JwtSecurity), Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.JwtSecurity, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.JwtSecurity, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}

	if mesh.Spec.Version == "1.3" {
		// SLO
		key = types.NamespacedName{Name: string(gmcore.Slo), Namespace: mesh.Namespace}
		if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.Slo, ObjectKey: key}); err != nil {
			return ctrl.Result{}, err
		}
		if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.Slo, ObjectKey: key}); err != nil {
			return ctrl.Result{}, err
		}

		// Postgres- SLO
		key = types.NamespacedName{Name: string(gmcore.Postgres), Namespace: mesh.Namespace}
		if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.Postgres, ObjectKey: key}); err != nil {
			return ctrl.Result{}, err
		}
		if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.Postgres, ObjectKey: key}); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Edge
	key = types.NamespacedName{Name: "edge", Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.Proxy, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Service{
		GmService:   gmcore.Proxy,
		ObjectKey:   key,
		ServiceKind: corev1.ServiceTypeLoadBalancer,
	}); err != nil {
		return ctrl.Result{}, err
	}

	// Ingress
	key = types.NamespacedName{Name: "ingress", Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Ingress{ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}

	// Control
	roleName := "control-pods"
	if err := apply(ctx, controller, mesh, configs, reconcilers.ClusterRole{
		Name: roleName,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"list"},
			},
		},
	}); err != nil {
		return ctrl.Result{}, err
	}
	saKey := types.NamespacedName{Name: roleName, Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.ServiceAccount{ObjectKey: saKey}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.ClusterRoleBinding{Name: roleName}); err != nil {
		return ctrl.Result{}, err
	}
	key = types.NamespacedName{Name: string(gmcore.Control), Namespace: mesh.Namespace}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Deployment{GmService: gmcore.Control, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}
	if err := apply(ctx, controller, mesh, configs, reconcilers.Service{GmService: gmcore.Control, ObjectKey: key}); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (controller *MeshController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Mesh{}).
		Complete(controller)
}
