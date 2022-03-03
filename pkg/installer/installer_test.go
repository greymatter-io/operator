package installer

import (
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"testing"
	//"github.com/greymatter-io/operator/pkg/cueutils"
	//"github.com/greymatter-io/operator/pkg/cuemodule"
	//_ appsv1 "k8s.io/api/apps/v1"
	//_ corev1 "k8s.io/api/core/v1"
	//_ rbacv1 "k8s.io/api/rbac/v1"
)

func TestLoading(t *testing.T) {

	// K8S MANIFESTS LOADING TEST

	//instances := load.Instances([]string{
	//	"./new_structure/k8s/outputs",
	//}, &load.Config{
	//	Dir: "/Users/danielpcox/projects/decipher/operator/pkg/cuemodule",
	//})
	//if len(instances) < 1 {
	//	t.Fatal("No go")
	//}
	//
	//value := cuecontext.New().BuildInstance(instances[0])
	////t.Logf("\nVALUE: %#v \nError: %#v", value, value.Err())
	//
	//var extracted struct {
	//	K8sManifests []json.RawMessage `json:"k8s_manifests"`
	//}
	//err := cuemodule.Extract(value, &extracted)
	//if err != nil {
	//	t.Fatalf("Extraction Error: %v", err)
	//}
	//
	//var extracted2 struct {
	//	Kind string `json:"kind"`
	//}
	//
	//var manifestObjects []k8sClient.Object
	//
	//for _, manifest := range extracted.K8sManifests {
	//	_ = json.Unmarshal(manifest, &extracted2)
	//	//t.Log(extracted2.Kind)
	//	switch extracted2.Kind {
	//	case "Namespace":
	//		var obj corev1.Namespace
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "Secret":
	//		var obj corev1.Secret
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "Service":
	//		var obj corev1.Service
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "Deployment":
	//		var obj appsv1.Deployment
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "StatefulSet":
	//		var obj appsv1.StatefulSet
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "DaemonSet":
	//		var obj appsv1.DaemonSet
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "Role":
	//		var obj rbacv1.Role
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "RoleBinding":
	//		var obj rbacv1.RoleBinding
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "ServiceAccount":
	//		var obj corev1.ServiceAccount
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "ClusterRole":
	//		var obj rbacv1.ClusterRole
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "ClusterRoleBinding":
	//		var obj rbacv1.ClusterRoleBinding
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	case "ConfigMap":
	//		var obj corev1.ConfigMap
	//		_ = json.Unmarshal(manifest, &obj)
	//		manifestObjects = append(manifestObjects, &obj)
	//	default:
	//		t.Logf("Got an unrecognized object: %#v \n", extracted2.Kind)
	//	}
	//}
	//
	//t.Logf("Extracted Objects: %#v", manifestObjects)
	//
	////v := //some k8s resource from the CUE
	//// figure out what kind it is, and make an s of that kind
	//
	////extracted := cueutils.Extract(v, s)
	//// k8sapi.Apply

	// GREY MATTER CONFIG LOADING TEST

	//instances := load.Instances([]string{
	//	"./new_structure/gm/outputs",
	//}, &load.Config{
	//	Dir: "/Users/danielpcox/projects/decipher/operator/pkg/cuemodule",
	//})
	//if len(instances) < 1 {
	//	t.Fatal("No go")
	//}
	//
	//value := cuecontext.New().BuildInstance(instances[0])
	//
	//var extracted struct {
	//	MeshConfigs []json.RawMessage `json:"mesh_configs"`
	//}
	//// TODO handle extraction error
	//_ = cuemodule.Extract(value, &extracted)
	//
	//// Just used for pulling the keys so we can determine the object type
	//type justKeys struct {
	//	ProxyKey    string `json:"proxy_key"`
	//	ClusterKey  string `json:"cluster_key"`
	//	RouteKey    string `json:"route_key"`
	//	DomainKey   string `json:"domain_key"`
	//	ListenerKey string `json:"listener_key"`
	//}
	//
	//var extracted2 justKeys
	//
	//for _, configObject := range extracted.MeshConfigs {
	//	extracted2 = justKeys{}
	//	_ = json.Unmarshal(configObject, &extracted2)
	//	if extracted2.ProxyKey != "" {
	//		t.Logf("Got a Proxy object: %#v \n", string(configObject))
	//	} else if extracted2.ClusterKey != "" {
	//		t.Logf("Got a Cluster object: %#v \n", string(configObject))
	//	} else if extracted2.RouteKey != "" {
	//		t.Logf("Got a Route object: %#v \n", string(configObject))
	//	} else if extracted2.DomainKey != "" {
	//		t.Logf("Got a Domain object: %#v \n", string(configObject))
	//	} else if extracted2.ListenerKey != "" {
	//		t.Logf("Got a Listener object: %#v \n", string(configObject))
	//	}
	//}

	// ODDS AND ENDS LOADING TEST

	//instances := load.Instances([]string{
	//	"./new_structure",
	//}, &load.Config{
	//	Dir: "/Users/danielpcox/projects/decipher/operator/pkg/cuemodule",
	//})
	//if len(instances) < 1 {
	//	t.Fatal("No go")
	//}
	//
	//value := cuecontext.New().BuildInstance(instances[0])
	////t.Logf("%#v", value)
	//
	//type Input struct {
	//	InstallNamespace string `json:"install_namespace"`
	//	WatchNamespaces  string `json:"watch_namespaces"` // comma-separated string
	//}
	//
	//var extracted struct {
	//	Input Input `json:"inputs"`
	//}
	//// TODO handle extraction error
	//_ = cuemodule.Extract(value, &extracted)
	//t.Logf("%#v", string(extracted.Input.InstallNamespace))

	// MESH LOADING TEST

	instances := load.Instances([]string{
		"./new_structure",
	}, &load.Config{
		Dir: "/Users/danielpcox/projects/decipher/operator/pkg/cuemodule",
	})
	if len(instances) < 1 {
		t.Fatal("No go")
	}

	value := cuecontext.New().BuildInstance(instances[0])
	//t.Logf("%#v", value)

	var extracted struct {
		Mesh v1alpha1.Mesh `json:"mesh"`
	}

	// TODO handle extraction error
	_ = cuemodule.Extract(value, &extracted)
	mesh := extracted.Mesh
	t.Logf("watch_namespaces: %#v UID: %#v", mesh.Spec.WatchNamespaces, mesh.UID)

	// then try tweaking the mesh and unifying, and seeing what we extract the second time
	t.Logf("Original name: %s", mesh.Name)
	mesh.Name = "different name"
	newMeshValue, _ := cuemodule.FromStruct("mesh", mesh)
	newValue := value.Unify(newMeshValue)
	_ = cuemodule.Extract(newValue, &extracted)

	t.Logf("New name: %s", extracted.Mesh.Name)
	t.Logf("New value: %#v", newValue)

	// UNMARSHALL TEST
	//data := []byte("{\"mesh_id\":\"Experiment\"}")
	//var extracted struct {
	//	MeshID string `json:"mesh_id"`
	//}
	//_ = json.Unmarshal(data, &extracted)
	//t.Logf("meshid:%v", extracted.MeshID)

}
