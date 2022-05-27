// Package gmapi executes greymatter CLI commands to configure mesh behavior
// in Control and Catalog APIs in each install namespace for each mesh.
// It enables Mesh CR specifications to define how a mesh should be configured.
package gmapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/greymatter-io/operator/pkg/wellknown"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"sync"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("gmapi")
)

// CLI exposes methods for configuring clients that execute greymatter CLI commands.
type CLI struct {
	*sync.RWMutex
	Client      *Client
	operatorCUE *cuemodule.OperatorCUE
}

// New returns a new *CLI instance.
// It receives a context for cleaning up goroutines started by the *CLI.
func New(ctx context.Context, operatorCUE *cuemodule.OperatorCUE) (*CLI, error) {
	v, err := cliversion()
	if err != nil {
		logger.Error(err, "Failed to initialize greymatter CLI")
		return nil, err
	}

	logger.Info("Using greymatter CLI", "Version", v)

	gmcli := &CLI{
		RWMutex:     &sync.RWMutex{},
		Client:      nil,
		operatorCUE: operatorCUE,
	}

	// Cancel all Client goroutines if package context is done.
	go func(c *CLI) {
		<-ctx.Done()
		c.RLock()
		defer c.RUnlock()
		if c.Client != nil {
			c.Client.Cancel()
		}
	}(gmcli)

	return gmcli, nil
}

// ConfigureMeshClient initializes or updates a Client with flags specifying connection options
// for reaching Control and Catalog for the given Mesh CR.
func (c *CLI) ConfigureMeshClient(mesh *v1alpha1.Mesh) {
	conf := mkCLIConfig( // TODO this should come from config
		// control
		fmt.Sprintf("http://controlensemble.%s.svc.cluster.local:5555", mesh.Spec.InstallNamespace),
		// catalog
		fmt.Sprintf("http://catalog.%s.svc.cluster.local:8080", mesh.Spec.InstallNamespace),
		mesh.Name,
	)
	flags := []string{"--base64-config", conf}

	if err := c.configureMeshClient(mesh, flags...); err != nil {
		logger.Error(err, "failed to configure Client", "Mesh", mesh.Name)
	}
}

func mkCLIConfig(apiHost, catalogHost, catalogMesh string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`
	[api]
	host = "%s"
	[catalog]
	host = "%s"
	mesh = "%s"
	`, apiHost, catalogHost, catalogMesh)))
}

func (c *CLI) configureMeshClient(mesh *v1alpha1.Mesh, flags ...string) error {
	c.Lock()
	defer c.Unlock()

	// Close an existing cmds channel if updating
	if c.Client != nil {
		logger.Info("Updating mesh Client", "Mesh", mesh.Name)
		c.Client.Cancel()
	} else {
		logger.Info("Initializing mesh Client", "Mesh", mesh.Name)
	}

	cl, err := newClient(c.operatorCUE, mesh, flags...)
	if err != nil {
		return err
	}

	c.Client = cl

	return nil
}

// RemoveMeshClient cleans up a Client's goroutines before removing it from the *CLI.
func (c *CLI) RemoveMeshClient() {
	if c.Client != nil {
		c.Client.Cancel()
	}
}

// ConfigureSidecar applies fabric objects that add a workload to the mesh specified
// given the workload's annotations and a list of its corev1.Containers.
func (c *CLI) ConfigureSidecar(operatorCUE *cuemodule.OperatorCUE, name string, metadata metav1.ObjectMeta) {
	annotations := metadata.Annotations
	injectedSidecarPortString, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
	var injectedSidecarPort int
	if injectSidecar {
		parsedPort, err := strconv.Atoi(injectedSidecarPortString)
		if err != nil {
			logger.Error(err, "provided port for sidecar upstream could not be parsed as int", wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT, injectedSidecarPortString)
			return
		}
		injectedSidecarPort = parsedPort
	} else { // if we're not injecting a sidecar, skip configuration
		return
	}

	// we skip configuration if we're explicitly told to
	configureSidecar := annotations[wellknown.ANNOTATION_CONFIGURE_SIDECAR]
	if configureSidecar == "false" {
		return
	}

	configObjects, kinds, err := operatorCUE.UnifyAndExtractSidecarConfig(name, injectedSidecarPort)
	if err != nil {
		logger.Error(err, "Failed to unify or extract CUE", "name", name, "injectedSidecarPort", injectedSidecarPort)
	}

	ApplyAll(c.Client, configObjects, kinds)
}

// UnconfigureSidecar removes fabric objects, disconnecting the workload from the mesh specified
func (c *CLI) UnconfigureSidecar(operatorCUE *cuemodule.OperatorCUE, name string, metadata metav1.ObjectMeta) {
	annotations := metadata.Annotations
	logger.Info("Unconfiguring sidecar with values", "name", name, "annotations", annotations)
	injectedSidecarPortString, injectSidecar := annotations[wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT]
	var injectedSidecarPort int
	if injectSidecar {
		parsedPort, err := strconv.Atoi(injectedSidecarPortString)
		if err != nil {
			logger.Error(err, "provided port for sidecar upstream could not be parsed as int", wellknown.ANNOTATION_INJECT_SIDECAR_TO_PORT, injectedSidecarPortString)
			return
		}
		injectedSidecarPort = parsedPort
	} else { // if we're not injecting a sidecar, skip configuration
		return
	}

	// we also skip configuration if we're explicitly told to
	configureSidecar := annotations[wellknown.ANNOTATION_CONFIGURE_SIDECAR]
	if configureSidecar == "false" {
		return
	}

	configObjects, kinds, err := operatorCUE.UnifyAndExtractSidecarConfig(name, injectedSidecarPort)
	if err != nil {
		logger.Error(err, "Failed to unify or extract CUE", "name", name, "injectedSidecarPort", injectedSidecarPort)
	}

	// HACK - This is kind of a silly way to get the redis listener out of here, but it lets me reuse UnifyAndExtractSidecarConfig as-is
	redisListener := configObjects[len(configObjects)-1]
	configObjects = configObjects[:len(configObjects)-1]
	kinds = kinds[:len(kinds)-1]
	c.Client.ControlCmds <- MkApply("listener", redisListener)

	UnApplyAll(c.Client, configObjects, kinds)
}
