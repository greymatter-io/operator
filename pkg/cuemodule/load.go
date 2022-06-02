package cuemodule

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"errors"
	"github.com/greymatter-io/operator/api/v1alpha1"
	opnshftsec "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("cuemodule")
)

// OperatorCUE holds the two top-level cue.Values that configure the operator,
// according to the major split between K8s and GM configuration
type OperatorCUE struct {
	// cue.Value for all of k8s/outputs containing k8s manifest objects
	K8s cue.Value

	// cue.Value for all of gm/outputs containing Grey Matter config objects
	GM cue.Value
}

// LoadAll loads the provided CUE for configuring the operator into an OperatorCUE and a Mesh
func LoadAll(cuemoduleRoot string) (*OperatorCUE, *v1alpha1.Mesh) {
	//cwd, _ := os.Getwd()
	allCUEInstances := load.Instances([]string{
		"./k8s/outputs",
		"./gm/outputs",
	}, &load.Config{
		Dir: cuemoduleRoot, // "If Dir is empty, the tool is run in the current directory"
	})
	operatorCUE := &OperatorCUE{}
	operatorCUE.K8s = cuecontext.New().BuildInstance(allCUEInstances[0])
	operatorCUE.GM = cuecontext.New().BuildInstance(allCUEInstances[1])
	if err := operatorCUE.K8s.Err(); err != nil {
		panic(err)
	}
	if err := operatorCUE.GM.Err(); err != nil {
		panic(err)
	}

	// load default mesh and store it in mesh_install. Later, one operator, one mesh.
	var extracted struct {
		Mesh v1alpha1.Mesh `json:"mesh"`
	}

	err := Extract(operatorCUE.K8s, &extracted)
	if err != nil {
		panic(err)
	}
	return operatorCUE, &extracted.Mesh
}

// Config represents the `config` struct from the operator CUE in inputs.cue
type Config struct {
	// Flags
	Spire                bool `json:"spire"`
	AutoApplyMesh        bool `json:"auto_apply_mesh"`
	GenerateWebhookCerts bool `json:"generate_webhook_certs"`

	// Values
	ClusterIngressName string `json:"cluster_ingress_name"`
}

type Defaults struct {
	SidecarList []string `json:"sidecar_list"`
}

// ExtractConfig pulls the values from the CUE into the Config struct in Go
// for use in various places in the operator
func (operatorCUE *OperatorCUE) ExtractConfig() (Config, Defaults) {
	var extracted struct {
		Config   Config   `json:"config"`
		Defaults Defaults `json:"defaults"`
	}

	err := Extract(operatorCUE.K8s, &extracted)
	if err != nil {
		panic(err)
	}

	return extracted.Config, extracted.Defaults
}

// TODO who should be responsible for logging errors - these, or the calling functions? I've been inconsistent about it

// UnifyWithMesh unifies the operatorCUE with a Mesh CR to fill in values
func (operatorCUE *OperatorCUE) UnifyWithMesh(mesh *v1alpha1.Mesh) error {
	meshValue, err := FromStruct("mesh", mesh)
	if err != nil {
		return err
	}
	k8sManifestsValue := operatorCUE.K8s.Unify(meshValue)
	if err := k8sManifestsValue.Err(); err != nil {
		logger.Error(err,
			"Error while attempting to unify provided Mesh resource with Kubernetes mesh configs CUE",
			"GM CUE", operatorCUE.K8s,
			"Mesh Value", meshValue,
			"Unification Result", k8sManifestsValue)
		return err
	}
	// We're also going to do unification with the GM CUE and cache it, as an optimization
	meshConfigsValue := operatorCUE.GM.Unify(meshValue)
	if err := meshConfigsValue.Err(); err != nil {
		logger.Error(err,
			"Error while attempting to unify provided Mesh resource with Grey Matter mesh configs CUE",
			"GM CUE", operatorCUE.GM,
			"Mesh Value", meshValue,
			"Unification Result", meshConfigsValue)
		return err
	}
	operatorCUE.K8s = k8sManifestsValue
	operatorCUE.GM = meshConfigsValue
	return nil
}

func (operatorCUE *OperatorCUE) TempGMValueUnifiedWithDefaults(defaults Defaults) (OperatorCUE, error) {
	defaultsValue, err := FromStruct("defaults", defaults)
	if err != nil {
		return OperatorCUE{}, err
	}
	meshConfigsValue := operatorCUE.GM.Unify(defaultsValue)
	if err := meshConfigsValue.Err(); err != nil {
		logger.Error(err,
			"Error while attempting to unify provided Defaults resource with Grey Matter mesh configs CUE",
			"GM CUE", operatorCUE.GM,
			"Defaults Value", defaultsValue,
			"Unification Result", meshConfigsValue)
		return OperatorCUE{}, err
	}
	return OperatorCUE{GM: meshConfigsValue, K8s: operatorCUE.K8s}, nil
}

// K8s Manifests

// ExtractCoreK8sManifests extracts the K8s manifests for a mesh from the top-level array in the k8s/outputs/EXTRACTME.cue
func (operatorCUE *OperatorCUE) ExtractCoreK8sManifests() (manifestObjects []client.Object, err error) {

	// Extract correct K8s config for options - for now there's only one
	var extracted struct {
		K8sManifests []json.RawMessage `json:"k8s_manifests"`
	}
	// TODO handle extraction error by exploding loudly
	err = Extract(operatorCUE.K8s, &extracted)
	if err != nil {
		return nil, err
	}

	manifestObjects = ExtractAndTypeK8sManifestObjects(extracted.K8sManifests)
	return manifestObjects, nil
}

// Mesh Configs

// ExtractCoreMeshConfigs extracts the GM config objects for a mesh from the top-level array in the gm/outputs/EXTRACTME.cue
func (operatorCUE *OperatorCUE) ExtractCoreMeshConfigs() (meshConfigs []json.RawMessage, kinds []string, err error) {
	var extracted struct {
		MeshConfigs []json.RawMessage `json:"mesh_configs"`
	}
	// TODO handle extraction error
	err = Extract(operatorCUE.GM, &extracted)
	if err != nil {
		return nil, nil, err // TODO error context?
	}
	kinds = IdentifyGMConfigObjects(extracted.MeshConfigs)
	return extracted.MeshConfigs, kinds, nil
}

// Deployment assist sidecar K8s and GM

// UnifyAndExtractSidecar unifies the cluster meant for a deployment with the CUE for a to-be-injected sidecar,
// and extracts the K8s manifest components to be injected
func (operatorCUE *OperatorCUE) UnifyAndExtractSidecar(clusterLabel string) (container corev1.Container, volumes []corev1.Volume, err error) {
	// By this point, we can assume GM has *already* been unified with THE mesh that this operator manages,
	// when the mesh was created.

	// Unify with name
	injectName := struct {
		Name string `json:"name"`
	}{Name: clusterLabel}
	withSidecarName, _ := FromStruct("sidecar_container", injectName)
	unifiedValue := operatorCUE.K8s.Unify(withSidecarName) // bit overkill, but it shouldn't matter
	if err := unifiedValue.Err(); err != nil {
		logger.Error(err,
			"Error while attempting to unify provided name with Grey Matter K8s CUE",
			"K8s CUE", operatorCUE.K8s,
			"Struct with sidecar name", withSidecarName,
			"Unification Result", unifiedValue)
		return container, volumes, err
	}

	type sidecarContainer struct {
		Container corev1.Container `json:"container"`
		Volumes   []corev1.Volume  `json:"volumes"`
	}
	var extracted struct {
		SidecarContainer sidecarContainer `json:"sidecar_container"`
	}
	// TODO handle extraction error by exploding loudly
	err = Extract(unifiedValue, &extracted)

	return extracted.SidecarContainer.Container, extracted.SidecarContainer.Volumes, err
}

// UnifyAndExtractSidecarConfig unifies a name and port with the Grey Matter sidecar configuration CUE for injected
// sidecars, and returns those configuration objects, along with their kinds (e.g., listener, cluster, etc.)
// It also extracts the special redis_listener object.
// NB: This method expects that the embedded Mesh in the CUE has already been updated with a status.sidecar_list
// for that redis_listener
func (operatorCUE *OperatorCUE) UnifyAndExtractSidecarConfig(name string, port int) (configObjects []json.RawMessage, kinds []string, err error) {

	// Unify with Name and Port
	injectNameAndPort := struct {
		Name string `json:"Name"`
		Port int    `json:"Port"`
	}{Name: name, Port: port}
	withNameAndPort, _ := FromStruct("sidecar_config", injectNameAndPort)
	unifiedValue := operatorCUE.GM.Unify(withNameAndPort) // bit overkill, but it shouldn't matter

	type sidecarConfig struct {
		LocalName         string            `json:"LocalName"`
		EgressToRedisName string            `json:"EgressToRedisName"`
		ConfigObjects     []json.RawMessage `json:"objects"`
	}

	var extracted struct {
		SidecarConfig sidecarConfig `json:"sidecar_config"`
	}
	err = Extract(unifiedValue, &extracted)
	// Extract sidecar container and (spire) volume
	if err != nil {
		return nil, nil, fmt.Errorf("extraction from CUE failed after workload value unification: %w", err)
	}

	kinds = IdentifyGMConfigObjects(extracted.SidecarConfig.ConfigObjects)

	return extracted.SidecarConfig.ConfigObjects, kinds, nil
}

// ExtractRedisListener returns the listener object for the redis listener with spire subjects set.
// Assumes unification has already happened to insert the correct sidecarList
func (operatorCUE *OperatorCUE) ExtractRedisListener() (configObject json.RawMessage, err error) {
	var extracted struct {
		RedisListener json.RawMessage `json:"redis_listener"`
	}
	err = Extract(operatorCUE.GM, &extracted)
	// Extract sidecar container and (spire) volume
	if err != nil {
		return nil, fmt.Errorf("redis listener extraction from CUE failed after workload value unification: %w", err)
	}
	return extracted.RedisListener, nil
}

type justKeys struct {
	ProxyKey    string `json:"proxy_key"`
	ClusterKey  string `json:"cluster_key"`
	RouteKey    string `json:"route_key"`
	DomainKey   string `json:"domain_key"`
	ListenerKey string `json:"listener_key"`
	ServiceID   string `json:"service_id"` // CatalogService
}

// IdentifyGMConfigObjects takes a list of raw objects and identifies them as particular GreyMatter config object types
func IdentifyGMConfigObjects(rawObjects []json.RawMessage) (kinds []string) {
	var extracted2 justKeys

	for _, configObject := range rawObjects {
		extracted2 = justKeys{}
		kind := ""
		_ = json.Unmarshal(configObject, &extracted2)
		if extracted2.ProxyKey != "" {
			kind = "proxy"
		} else if extracted2.ClusterKey != "" {
			kind = "cluster"
		} else if extracted2.RouteKey != "" { // route_key check must come before domain b/c routes have a domain_key
			kind = "route"
		} else if extracted2.DomainKey != "" {
			kind = "domain"
		} else if extracted2.ListenerKey != "" {
			kind = "listener"
		} else if extracted2.ServiceID != "" {
			kind = "catalogservice"
		}
		kinds = append(kinds, kind)
	}
	return kinds
}

// ExtractAndTypeK8sManifestObjects takes a list of raw k8s manifest objects, determines their types, and unmarshals
// each one into an object of the correct type.
func ExtractAndTypeK8sManifestObjects(manifests []json.RawMessage) (manifestObjects []client.Object) {

	var ke struct {
		Kind string `json:"kind"`
	}

	// TODO It'll be important to explode on an unmarshal error in case
	// customers have provided custom CUE for the operator to install
	for _, manifest := range manifests {
		_ = json.Unmarshal(manifest, &ke)
		//t.Log(ke.Kind)
		switch ke.Kind {
		case "Namespace":
			var obj corev1.Namespace
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "Secret":
			var obj corev1.Secret
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "Service":
			var obj corev1.Service
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "Deployment":
			var obj appsv1.Deployment
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "StatefulSet":
			var obj appsv1.StatefulSet
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "DaemonSet":
			var obj appsv1.DaemonSet
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "Role":
			var obj rbacv1.Role
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "RoleBinding":
			var obj rbacv1.RoleBinding
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "ServiceAccount":
			var obj corev1.ServiceAccount
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "ClusterRole":
			var obj rbacv1.ClusterRole
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "ClusterRoleBinding":
			var obj rbacv1.ClusterRoleBinding
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "ConfigMap":
			var obj corev1.ConfigMap
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
		case "SecurityContextConstraints":
			var obj opnshftsec.SecurityContextConstraints
			_ = json.Unmarshal(manifest, &obj)
			manifestObjects = append(manifestObjects, &obj)
			logger.Info("SecurityContextConstraints object loaded", "object", obj) // DEBUG
		default:
			logger.Error(errors.New("got unrecognized K8s manifest object from CUE"), "ignoring", "Kind", ke.Kind, "Object", manifest)
		}
	}
	return manifestObjects
}

//// TODO remove old stuff once we've adjusted the tests vvvvv
//
//// Loader loads a package from our CUE module.
//type Loader func(string) (cue.Value, error)
//
//// LoadPackage loads a package from our Cue module.
//// Packages are added to subdirectories and declared with the same name as the subdirectory.
//func LoadPackage(pkgName string) (cue.Value, error) {
//	dirPath, err := os.Getwd()
//	if err != nil {
//		return cue.Value{}, err
//	}
//
//	return loadPackage(pkgName, dirPath)
//}
//
//// LoadPackageForTest loads a package from our Cue module within a test context.
//func LoadPackageForTest(pkgName string) (cue.Value, error) {
//	_, filename, _, _ := runtime.Caller(0)
//	dirPath := path.Dir(filename)
//
//	return loadPackage(pkgName, dirPath)
//}
//
//func loadPackage(pkgName, dirPath string) (cue.Value, error) {
//	instances := load.Instances([]string{"greymatter.io/operator/" + pkgName}, &load.Config{
//		ModuleRoot: dirPath,
//	})
//
//	if len(instances) != 1 {
//		return cue.Value{}, fmt.Errorf("did not load expected package %s", pkgName)
//	}
//
//	value := cuecontext.New().BuildInstance(instances[0])
//	if err := value.Err(); err != nil {
//		return cue.Value{}, err
//	}
//
//	return value, nil
//}
