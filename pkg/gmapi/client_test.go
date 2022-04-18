package gmapi

import (
	"testing"
)

//func TestNewClient(t *testing.T) {
//	t.Skip() // only test in dev, not in ci
//
//	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
//
//	ctx, cancel := context.WithCancel(context.Background())
//
//	c, err := New(ctx, cuemodule.LoadPackageForTest, true)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	mesh := &v1alpha1.Mesh{
//		ObjectMeta: metav1.ObjectMeta{Name: "mesh"},
//		Spec: v1alpha1.MeshSpec{
//			Zone:           "zone",
//			ReleaseVersion: "1.7",
//		},
//	}
//
//	c.configureMeshClient(
//		mesh,
//		"--base64-config", mkCLIConfig(
//			"http://localhost:5555",
//			"http://localhost:8181",
//			"mesh",
//		),
//	)
//
//	deployment := &appsv1.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "example",
//			Namespace: "myns",
//		},
//		Spec: appsv1.DeploymentSpec{
//			Template: corev1.PodTemplateSpec{
//				Spec: corev1.PodSpec{
//					Containers: []corev1.Container{
//						{
//							Name: "api",
//							Ports: []corev1.ContainerPort{
//								{ContainerPort: 5555},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	c.ConfigureSidecar("mesh", "mock", deployment)
//	c.UnconfigureSidecar("mesh", "mock", deployment)
//
//	c.RemoveMeshClient("mesh")
//
//	time.Sleep(time.Second * 5)
//	cancel()
//}

func TestCLIVersion(t *testing.T) {
	t.Skip() // only test in dev, not in ci

	v, err := cliversion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
