package webhooks

import (
	"context"
	"net/http"

	"github.com/greymatter-io/operator/api/v1alpha1"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type meshDefaulter struct {
	inst
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
	inst
	decoder *admission.Decoder
}

// Implements admission.DecoderInjector.
// A decoder will be automatically injected.
func (mv *meshValidator) InjectDecoder(d *admission.Decoder) error {
	mv.decoder = d
	return nil
}

func (mv *meshValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		go mv.RemoveMesh(req.Name)
		return admission.ValidationResponse(true, "allowed")
	}

	// TODO: Ensure only one mesh exists in a namespace
	// TODO: Ensure namespace doesn't belong to another Mesh (as a WatchNamespace)
	// TODO: Ensure Mesh watch namespaces are unique

	mesh := &v1alpha1.Mesh{}
	if err := mv.decoder.Decode(req, mesh); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if req.Operation == admissionv1.Create {
		go mv.ApplyMesh(mesh, true)
	} else {
		go mv.ApplyMesh(mesh, false)
	}

	return admission.ValidationResponse(true, "allowed")
}
