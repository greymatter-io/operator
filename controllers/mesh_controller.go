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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
	"github.com/bcmendoza/gm-operator/controllers/gmcore"
	"github.com/bcmendoza/gm-operator/controllers/meshobjects"
	"github.com/bcmendoza/gm-operator/controllers/reconcilers"
)

// MeshReconciler reconciles a Mesh object
type MeshReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	ObjectCache meshobjects.Cache
}

type reconcileId string

//+kubebuilder:rbac:groups=install.greymatter.io,resources=meshes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=install.greymatter.io,resources=meshes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=install.greymatter.io,resources=meshes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts;secrets;pods,verbs=get;list;watch;create
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mesh object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MeshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.WithValue(ctx, reconcileId("id"), uuid.New().String())
	log := r.Log.WithValues("mesh", req.NamespacedName)

	// Fetch the Mesh object
	mesh := &installv1.Mesh{}
	if err := r.Get(ctx, req.NamespacedName, mesh); err != nil {
		if errors.IsNotFound(err) {
			// Mesh object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Mesh resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Mesh")
		return ctrl.Result{}, err
	}

	// For now add defaults here.
	// Later this can be added to a mutating webhook.
	if mesh.Spec.ImagePullSecret == "" {
		secret := "docker.secret"
		mesh.Spec.ImagePullSecret = secret
	}

	// Control API
	key := types.NamespacedName{Name: string(gmcore.ControlApi), Namespace: mesh.Namespace}
	err := r.reconcile(ctx, mesh, reconcilers.Deployment{GmService: gmcore.ControlApi, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcile(ctx, mesh, reconcilers.Service{GmService: gmcore.ControlApi, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Control
	name := "control-pods"
	err = r.reconcile(ctx, mesh, reconcilers.ClusterRole{Name: name})
	if err != nil {
		return ctrl.Result{}, err
	}
	sarKey := types.NamespacedName{Name: name, Namespace: mesh.Namespace}
	err = r.reconcile(ctx, mesh, reconcilers.ServiceAccount{ObjectKey: sarKey})
	if err != nil {
		return ctrl.Result{}, err
	}
	// TODO: The ClusterRoleBinding should be updated with added subjects per namespace.
	// If another mesh is deployed into another namespace, this will break.
	err = r.reconcile(ctx, mesh, reconcilers.ClusterRoleBinding{Name: name})
	if err != nil {
		return ctrl.Result{}, err
	}
	key = types.NamespacedName{Name: string(gmcore.Control), Namespace: mesh.Namespace}
	err = r.reconcile(ctx, mesh, reconcilers.Deployment{GmService: gmcore.Control, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcile(ctx, mesh, reconcilers.Service{GmService: gmcore.Control, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Catalog
	key = types.NamespacedName{Name: string(gmcore.Catalog), Namespace: mesh.Namespace}
	err = r.reconcile(ctx, mesh, reconcilers.Deployment{GmService: gmcore.Catalog, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcile(ctx, mesh, reconcilers.Service{GmService: gmcore.Catalog, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Edge
	key = types.NamespacedName{Name: "edge", Namespace: mesh.Namespace}
	err = r.reconcile(ctx, mesh, reconcilers.Deployment{GmService: gmcore.Proxy, ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcile(ctx, mesh, reconcilers.Service{
		GmService:   gmcore.Proxy,
		ObjectKey:   key,
		ServiceKind: corev1.ServiceTypeLoadBalancer,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Ingress
	key = types.NamespacedName{Name: "ingress", Namespace: mesh.Namespace}
	err = r.reconcile(ctx, mesh, reconcilers.Ingress{ObjectKey: key})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Mesh object configuration
	// TODO: Add a ping; if non-responsive, start over
	// TODO: Track the status of each object LOCALLY and store in mesh CR
	if !mesh.Status.Deployed {
		addr := fmt.Sprintf("http://control-api.%s.svc:5555", mesh.Namespace)
		client := meshobjects.NewClient(addr)
		if err := client.MkMeshObjects(
			"zone-default-zone",
			[]string{"control-api:5555", "catalog:9080"},
		); err != nil {
			r.Log.Error(err, "failed to configure mesh")
			return ctrl.Result{}, err
		}
		mesh.Status.Deployed = true
		if err := r.Status().Update(ctx, mesh); err != nil {
			log.Error(err, "Failed to set mesh status to deployed")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MeshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&installv1.Mesh{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&extensionsv1beta1.Ingress{}).
		Complete(r)
}
