package version

import (
	"fmt"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cueutils"

	"cuelang.org/go/cue"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("version")
)

// Version contains a cue.Value that holds all installation templates for a
// version of Grey Matter, plus options applied from a Mesh custom resource.
type Version struct {
	cue  cue.Value
	opts []cue.Value
}

type Opt func(o []cue.Value)

func WithIngressSubDomain(domain string) func([]cue.Value) {
	return func(opts []cue.Value) {
		//lint:ignore SA4010,SA4006 slices are pointers so this actually does get used
		opts = append(opts, cueutils.FromStrings(fmt.Sprintf(`IngressSubDomain: %s`, domain)))
	}
}

// New creates a *Version instance which contains all manifests needed to install a version of Grey Matter.
// It accepts a base install templates and a Mesh CR which overrides defaults when supplied.
func New(tmpl cue.Value, mesh *v1alpha1.Mesh, opts ...Opt) (*Version, error) {
	v := &Version{
		cue:  tmpl,
		opts: make([]cue.Value, 0),
	}
	for _, o := range opts {
		o(v.opts)
	}

	m, err := cueutils.FromStruct("mesh", mesh)
	if err != nil {
		return nil, err
	}

	// Unify our base install template with the mesh CR, plus any supplied options.
	// TODO (alec): These will generally be very small unifications but we need to be careful about
	// how expensive this can get.
	v.Unify(append(v.opts, m)...)
	if err := v.cue.Err(); err != nil {
		cueutils.LogError(logger, err)
		return nil, err
	}

	return v, nil
}

// Unify combines multiple cue.Values into a Version's cue.Value.
func (v *Version) Unify(ws ...cue.Value) {
	for _, w := range ws {
		v.cue = v.cue.Unify(w)
	}
}
