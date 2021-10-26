package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/installer"

	admissionv1 "k8s.io/api/admission/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type meshDefaulter struct {
	*installer.Installer
	*admission.Decoder
}

// InjectDecoder implements admission.DecoderInjector.
// A decoder will be automatically injected for decoding meshes.
func (md *meshDefaulter) InjectDecoder(d *admission.Decoder) error {
	md.Decoder = d
	return nil
}

// Handle implements admission.Handler.
// It will be invoked for defaulting values prior to creating or updating a Mesh.
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
	ctrlclient.Client
}

// InjectDecoder implements admission.DecoderInjector.
// A decoder will be automatically injected for decoding meshes.
func (mv *meshValidator) InjectDecoder(d *admission.Decoder) error {
	mv.Decoder = d
	return nil
}

// Handle implements admission.Handler.
// It will be invoked for validating values prior to creating or updating a Mesh.
func (mv *meshValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		prev := &v1alpha1.Mesh{}
		if err := mv.DecodeRaw(req.OldObject, prev); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		go mv.RemoveMesh(prev)
		return admission.ValidationResponse(true, "allowed")
	}

	mesh := &v1alpha1.Mesh{}
	if err := mv.DecodeRaw(req.Object, mesh); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	installNS := mesh.Spec.InstallNamespace
	watchNS := strings.Join(mesh.Spec.WatchNamespaces, ",")
	if installNS == "gm-operator" {
		return admission.ValidationResponse(false, "blocked attempt to install Mesh in 'gm-operator' namespace")
	}

	meshList := &v1alpha1.MeshList{}
	if err := mv.List(context.TODO(), meshList); err != nil {
		logger.Error(err, "failed to list all meshes to validate namespaces", "Mesh", mesh.Name)
		return admission.ValidationResponse(false, "Internal server error; check logs with valid cluster permissions")
	}
	for _, m := range meshList.Items {
		// Ensure install namespace isn't occupied by another Mesh
		if m.Name != mesh.Name && m.Spec.InstallNamespace == installNS {
			return admission.ValidationResponse(false, fmt.Sprintf("blocked attempt to install second Mesh in namespace %s (occupied by Mesh %s)", installNS, m.Name))
		}
		// Ensure watch namespaces don't include another Mesh's install namespace
		if strings.Contains(watchNS, m.Spec.InstallNamespace) {
			return admission.ValidationResponse(false, fmt.Sprintf("blocked attempt to include watch namespace %s in Mesh (install namespace for Mesh %s)", installNS, m.Name))
		}
		for _, watched := range m.Spec.WatchNamespaces {
			// Ensure install namespace isn't watched by another Mesh
			if watched == installNS {
				return admission.ValidationResponse(false, fmt.Sprintf("blocked attempt to install Mesh in watched namespace %s (watched by Mesh %s)", installNS, m.Name))
			}
			// Ensure watch namespaces don't include a namespace watched by another Mesh
			if strings.Contains(watchNS, watched) {
				return admission.ValidationResponse(false, fmt.Sprintf("blocked attempt to include watch namespace %s in Mesh (already watched by Mesh %s)", watched, m.Name))
			}
		}
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
