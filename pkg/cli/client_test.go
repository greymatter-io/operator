package cli

import (
	"context"
	"testing"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestNewClient(t *testing.T) {
	t.Skip() // only test in dev, not in ci

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	ctx, cancel := context.WithCancel(context.Background())

	c, err := New(ctx, cuemodule.LoadPackageForTest, true)
	if err != nil {
		t.Fatal(err)
	}

	mesh := &v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
		Spec: v1alpha1.MeshSpec{
			Zone:           "zone",
			ReleaseVersion: "1.7",
		},
	}

	c.configureMeshClient(
		mesh,
		"--base64-config", mkCLIConfig(
			"http://localhost:5555",
			"http://localhost:8181",
			"mesh",
		),
	)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "myns",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "api",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5555},
							},
						},
					},
				},
			},
		},
	}

	c.ConfigureService("mesh", "mock", deployment)
	c.RemoveService("mesh", "mock", deployment)

	c.RemoveMeshClient("mesh")

	time.Sleep(time.Second * 5)
	cancel()
}

func TestCLIVersion(t *testing.T) {
	t.Skip() // only test in dev, not in ci

	v, err := cliversion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
