package version

import (
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/cueutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	mesh = &v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "default-mesh"},
		Spec: v1alpha1.MeshSpec{
			ReleaseVersion:   "latest",
			InstallNamespace: "myns",
			WatchNamespaces:  []string{"ns2", "ns3"},
			Images: v1alpha1.Images{
				Catalog: "docker.greymatter.io/development/gm-catalog:3.0.0",
			},
			Zone: "default-zone",
		},
	}
)

func TestNew(t *testing.T) {
	base, err := cuemodule.LoadPackageForTest("base")
	if err != nil {
		cueutils.LogError(logger, err)
		t.FailNow()
	}

	versions, err := New(base, mesh)
	if err != nil {
		t.FailNow()
	}

	t.Logf("%+v", versions.cue)

	m := versions.Manifests()
	for _, group := range m {
		if group.Deployment != nil {
			c := group.Deployment.Spec.Template.Spec.Containers
			for _, container := range c {
				// We expect catalog to be overriden here
				if container.Name == "catalog" {
					if container.Image != mesh.Spec.Images.Catalog {
						t.Fatalf("received the incorrect catalog image, want: %s, got: %s", mesh.Spec.Images.Catalog, container.Image)
					}
				}
			}
		}
	}
}
