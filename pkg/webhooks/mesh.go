package webhooks

import (
	"context"
	"net/http"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"

	admissionv1 "k8s.io/api/admission/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type meshDefaulter struct {
	*installer.Installer
	*admission.Decoder
}

// Implements admission.DecoderInjector.
// A decoder will be automatically injected.
func (md *meshDefaulter) InjectDecoder(d *admission.Decoder) error {
	md.Decoder = d
	return nil
}

func (md *meshDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {

	return admission.ValidationResponse(true, "allowed")
	// mesh := &v1alpha1.Mesh{}
	// md.decoder.Decode(req, mesh)
	// update, err := json.Marshal(req)
	// if err != nil {
	// 	return admission.Errored(http.StatusInternalServerError, err)
	// }

	// return admission.PatchResponseFromRaw(req.Object.Raw, update)
}

type meshValidator struct {
	*installer.Installer
	*cli.CLI
	*admission.Decoder
	ctrlclient.Client
}

// Implements admission.DecoderInjector.
// A decoder will be automatically injected.
func (mv *meshValidator) InjectDecoder(d *admission.Decoder) error {
	mv.Decoder = d
	return nil
}

func (mv *meshValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		prev := &v1alpha1.Mesh{}
		if err := mv.DecodeRaw(req.OldObject, prev); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		go mv.RemoveMesh(prev)
		go mv.RemoveMeshClient(req.Name)
		return admission.ValidationResponse(true, "allowed")
	}

	if req.Namespace == "gm-operator" {
		return admission.ValidationResponse(false, "attempted to create Mesh in 'gm-operator' namespace")
	}

	// Ensure that each mesh spec has a unique install-namespace
	meshList := &v1alpha1.MeshList{}
	if err := mv.List(context.TODO(), meshList); err != nil {
		return admission.ValidationResponse(false, "Unable to get list of existing mesh resources")
	}
	// parse through meshes  and see if the request object Namespace is already in a mesh spec
	obj := &v1alpha1.Mesh{}
	if err := mv.DecodeRaw(req.Object, obj); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	for _, mesh := range meshList.Items {
		if mesh.Spec.InstallNamespace == obj.Spec.InstallNamespace {
			return admission.ValidationResponse(false, "A mesh exists with the specified install_namespace already.")
		}
	}

	// TODO: Ensure only one mesh exists in a namespace
	// TODO: Ensure namespace doesn't belong to another Mesh (as a WatchNamespace)
	// TODO: Ensure Mesh watch namespaces are unique

	go mv.ConfigureMeshClient(obj)

	if req.Operation == admissionv1.Create {
		go mv.ApplyMesh(nil, obj)
	} else {
		prev := &v1alpha1.Mesh{}
		if err := mv.DecodeRaw(req.OldObject, prev); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		go mv.ApplyMesh(prev, obj)
	}

	return admission.ValidationResponse(true, "allowed")
}
