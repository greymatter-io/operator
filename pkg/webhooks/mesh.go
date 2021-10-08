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
	decoder *admission.Decoder
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

// Implements admission.DecoderInjector.
// A decoder will be automatically injected.
func (md *meshDefaulter) InjectDecoder(d *admission.Decoder) error {
	md.decoder = d
	return nil
}

type meshValidator struct {
	inst
	decoder *admission.Decoder
}

func (mv *meshValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	mesh := &v1alpha1.Mesh{}
	if err := mv.decoder.Decode(req, mesh); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	switch req.Operation {
	case admissionv1.Create:
		go mv.ApplyMesh(mesh, true)
	case admissionv1.Update:
		go mv.ApplyMesh(mesh, false)
	case admissionv1.Delete:
		go mv.RemoveMesh(mesh.Name)
	}

	return admission.ValidationResponse(true, "allowed")
}

// Implements admission.DecoderInjector.
// A decoder will be automatically injected.
func (mv *meshValidator) InjectDecoder(d *admission.Decoder) error {
	mv.decoder = d
	return nil
}
