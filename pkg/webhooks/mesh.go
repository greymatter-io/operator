package webhooks

import (
	"context"
	"net/http"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/installer"

	admissionv1 "k8s.io/api/admission/v1"
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
	*admission.Decoder
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
		return admission.ValidationResponse(true, "allowed")
	}

	// TODO: Ensure only one mesh exists in a namespace
	// TODO: Ensure namespace doesn't belong to another Mesh (as a WatchNamespace)
	// TODO: Ensure Mesh watch namespaces are unique

	mesh := &v1alpha1.Mesh{}
	if err := mv.Decode(req, mesh); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if req.Operation == admissionv1.Create {
		go mv.ApplyMesh(nil, mesh)
	} else {
		prev := &v1alpha1.Mesh{}
		if err := mv.DecodeRaw(req.OldObject, prev); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		go mv.ApplyMesh(prev, mesh)
	}

	return admission.ValidationResponse(true, "allowed")
}
