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
	name string
	cue  cue.Value
	opts []cue.Value
}

type Opt func(o []cue.Value)

func WithIngressSubDomain(domain string) func([]cue.Value) {
	return func(opts []cue.Value) {
		//lint:ignore SA4006 slices are pointers so this actually does get used
		opts = append(opts, cueutils.FromStrings(fmt.Sprintf(`domain: %s`, domain)))
	}
}

// New creates a *Version instance which contains all known internal container versions.
// It accepts a Mesh CR which overrides defaults if supplied.
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

	// Add all are options into the original base components.cue evaluated template.
	v.cue = v.cue.Unify(m)

	// TODO (alec): These will generally be very small unifications but we need to be careful about
	// how expensive this can get.
	v.Unify(v.opts...)

	return v, nil
}

// Copy deep copies a Version's cue.Value into a new Version.
func (v Version) Copy() Version {
	return Version{v.name, v.cue, v.opts}
}

// Unify gets the lower bound cue.Value of Version.cue and all argument values.
func (v *Version) Unify(ws ...cue.Value) {
	for _, w := range ws {
		v.cue = v.cue.Unify(w)
	}
}
