package k8sapi

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	logger = ctrl.Log.WithName("k8sapi")
)

// Apply is a functional interface for interacting with the K8s apiserver in a consistent way.
// Each client.Object argument must implement the necessary interfaces for Reader/Writer interfaces implemented by client.Client.
func Apply(c client.Client, obj, owner client.Object, action func(client.Client, client.Object) (string, error)) {
	scheme := c.Scheme()

	var kind string
	if gvk, err := apiutil.GVKForObject(obj.(runtime.Object), scheme); err != nil {
		kind = "Object"
	} else {
		kind = gvk.Kind
	}

	// Set an owner reference on the manifest for garbage collection if the owner is deleted.
	var ownerName string
	if owner != nil {
		ownerName = client.ObjectKeyFromObject(owner).Name
		if err := controllerutil.SetOwnerReference(owner, obj, scheme); err != nil {
			logger.Error(err, "SetOwnerReference", "Owner", ownerName, kind, client.ObjectKeyFromObject(obj))
			return
		}
	}

	act, err := action(c, obj)
	if err != nil {
		if ownerName != "" {
			logger.Error(err, act, "Owner", ownerName, kind, client.ObjectKeyFromObject(obj))
		} else {
			logger.Error(err, act, kind, client.ObjectKeyFromObject(obj))
		}
		return
	}

	if ownerName != "" {
		logger.Info(act, "Owner", ownerName, kind, client.ObjectKeyFromObject(obj))
	} else {
		logger.Info(act, kind, client.ObjectKeyFromObject(obj))
	}
}

// CreateOrUpdate applies a resource in the K8s apiserver.
func CreateOrUpdate(c client.Client, obj client.Object) (string, error) {
	key := client.ObjectKeyFromObject(obj)

	// Make a pointer copy of the object so that our actual object is not modified by client.Get.
	// This way, the object passed into client.Update still has our desired state.
	existing := obj.DeepCopyObject()
	if err := c.Get(context.TODO(), key, existing.(client.Object)); err != nil {
		if !errors.IsNotFound(err) {
			return "create/update", err
		}
		if err := c.Create(context.TODO(), obj); err != nil {
			return "create", err
		}
		return "create", nil
	}

	if err := c.Update(context.TODO(), obj); err != nil {
		return "update", err
	}

	return "update", nil
}

// GetOrCreate ensures a resource exists in the K8s apiserver.
func GetOrCreate(c client.Client, obj client.Object) (string, error) {
	key := client.ObjectKeyFromObject(obj)

	if err := c.Get(context.TODO(), key, obj); err != nil {
		if err := c.Create(context.TODO(), obj); err != nil {
			return "create", err
		}
		return "create", nil
	}

	return "get", nil
}

// MkPatchAction returns a function that applies the patch specified when called.
func MkPatchAction(patch func(client.Object) client.Object) func(client.Client, client.Object) (string, error) {
	return func(c client.Client, obj client.Object) (string, error) {
		key := client.ObjectKeyFromObject(obj)
		if err := c.Get(context.TODO(), key, obj); err != nil {
			return "get", err
		}

		mp := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		obj = patch(obj)
		if err := c.Patch(context.TODO(), obj, mp); err != nil {
			return "patch", err
		}

		return "patch", nil
	}
}
