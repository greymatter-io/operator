/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/bootstrap"
	"github.com/greymatter-io/operator/pkg/cfsslsrv"
	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"github.com/greymatter-io/operator/pkg/installer"
	"github.com/greymatter-io/operator/pkg/webhooks"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	configv1 "github.com/openshift/api/config/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	basecfg "k8s.io/component-base/config/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
	logger = ctrl.Log.WithName("init")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(extv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// Global config flags
var (
	configPath string
	zapDevMode bool
  pprofAddr  string
)

func main() {
	if err := run(); err != nil {
		logger.Error(err, "Failed to run operator")
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	flag.StringVar(&pprofAddr, "pprofAddr", ":1234", "Address for pprof server; has no effect on release builds")
	flag.StringVar(&configPath, "configPath", "", "The operator will load its initial configuration from this file if defined.")
	flag.BoolVar(&zapDevMode, "zapDevMode", false, "Configure zap logger in development mode.")

	// Bind flags for Zap logger options.
	opts := zap.Options{Development: zapDevMode}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Initialize operator options with set values.
	// These values will not be replaced by any values set in a read configPath.
	options := ctrl.Options{
		Scheme:                  scheme,
		LeaderElection:          true,
		LeaderElectionID:        "715805a0.greymatter.io",
		LeaderElectionNamespace: "gm-operator",
		Port:                    9443,
		MetricsBindAddress:      ":8080",
		HealthProbeBindAddress:  ":8081",
	}

	// Default bootstrap config values
	defaultBootstrapConfig := bootstrap.BootstrapConfig{
		// LeaderElection is required as an empty config since it cannot be nil.
		ControllerManagerConfigurationSpec: cfg.ControllerManagerConfigurationSpec{
			LeaderElection: &basecfg.LeaderElectionConfiguration{},
		},
		ClusterIngressName: "cluster",
	}

	// Attempt to read a configPath if one has been configured.
	cfg := defaultBootstrapConfig
	var err error
	if configPath != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configPath).OfKind(&cfg))
		if err != nil {
			return fmt.Errorf("failed to load bootstrap config at path %s: %w", configPath, err)
		}
		logger.Info("Loaded bootstrap config", "Path", configPath)
	}

	// Start up our CFSSL server for issuing two certs:
	// 1) Webhook server certs (unless disabled in the bootstrap config)
	// 2) SPIRE's intermediate CA for issuing identities to workloads
	cs, err := cfsslsrv.New(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to configure CFSSL server: %w", err)
	}
	if err := cs.Start(); err != nil {
		return fmt.Errorf("failed to start CFSSL server: %w", err)
	}

	// Create context for goroutine cleanup
	ctx := ctrl.SetupSignalHandler()

	// Initialize interface with greymatter CLI
	// For now, mTLSEnabled is always true since we install SPIRE by default.
	gmcli, err := cli.New(ctx, cuemodule.LoadPackage, true)
	if err != nil {
		return err
	}

	// Create a rest.Config that has settings for communicating with the K8s cluster.
	restConfig := ctrl.GetConfigOrDie()

	// Create a write+read client for making requests to the API server.
	c, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create initial client: %w", err)
	}

	// Initialize controller-runtime manager with configured options
	mgr, err := ctrl.NewManager(restConfig, options)
	if err != nil {
		return fmt.Errorf("failed to initialize controller-manager: %w", err)
	}

	// Initialize manifests installer.
	inst, err := installer.New(c, cuemodule.LoadPackage, gmcli, cs, cfg.ClusterIngressName)
	if err != nil {
		return fmt.Errorf("failed to initialize manifest installer: %w", err)
	}

	// Initialize the webhooks loader.
	wl, err := webhooks.New(c, inst, gmcli, cs, cfg.DisableWebhookCertGeneration, mgr.GetWebhookServer)
	if err != nil {
		return err
	}

	// Register our webhooks loader and manifests installer into the controller manager's start process queue.
	mgr.Add(wl)
	mgr.Add(inst)

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("failed to set up healthz endpoint: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("failed to set up readyz endpoint: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start controller-manager: %w", err)
	}

	return nil
}
