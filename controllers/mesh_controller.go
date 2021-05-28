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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	installv1 "github.com/bcmendoza/gm-operator/api/v1"
)

// MeshReconciler reconciles a Mesh object
type MeshReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=install.greymatter.io,resources=meshes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=install.greymatter.io,resources=meshes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=install.greymatter.io,resources=meshes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;create;update;patch;delete

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
		return ctrl.Result{Requeue: true}, err
	}

	if mesh.Spec.ImagePullSecret == nil {
		secret := "docker.secret"
		mesh.Spec.ImagePullSecret = &secret
	}

	var gmi gmImages
	if mesh.Spec.Version != nil {
		switch *mesh.Spec.Version {
		case "1.2":
			gmi = gmVersionMap["1.2"]
		default:
			gmi = gmVersionMap["1.3"]
		}
	}

	// Control API
	if err := r.mkControlAPI(ctx, mesh, gmi); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Control
	if err := r.mkControl(ctx, mesh, gmi); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Edge
	if err := r.mkEdge(ctx, mesh, gmi); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Ingress
	if err := r.mkIngress(ctx, mesh); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MeshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&installv1.Mesh{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&extensionsv1beta1.Ingress{}).
		Complete(r)
}
