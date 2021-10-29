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
	"os"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/bootstrap"
	"github.com/greymatter-io/operator/pkg/cli"
	"github.com/greymatter-io/operator/pkg/installer"
	"github.com/greymatter-io/operator/pkg/webhooks"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
	logger = ctrl.Log.WithName("setup")

	// Global config flags
	configFile           string
	metricsAddr          string
	probeAddr            string
	enableLeaderElection bool
	development          bool
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// Add tags here for generating RBAC rules for the role that will be used by the Operator.
//+kubebuilder:rbac:groups=greymatter.io,resources=meshes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=greymatter.io,resources=meshes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;configmaps;serviceaccounts;secrets;pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func main() {
	flag.StringVar(&configFile, "config", "", "The operator will load its initial configuration from this file if defined.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true, "Enable leader election, ensuring only one active controller manager.")
	flag.BoolVar(&development, "development", false, "Run in development mode.")

	// Bind flags for Zap logger options, which I assume allows args to be passed in by OLM (?)
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// If the development flag is set, override previous settings as this is likely not running in-cluster.
	if development {
		opts.Development = development
		enableLeaderElection = false
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Initialize manager options with set values.
	// These values will not be replaced by any values set in a read configFile.
	var err error
	options := ctrl.Options{
		Scheme:                  scheme,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "715805a0.greymatter.io",
		LeaderElectionNamespace: "gm-operator",
	}

	// Attempt to read a configFile if one has been configured.
	cfg := bootstrap.BootstrapConfig{}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&cfg))
		if err != nil {
			logger.Error(err, "Unable to load bootstrap config", "path", configFile)
			os.Exit(1)
		} else {
			logger.Info("Loaded bootstrap config", "Path", configFile)
		}
	}

	// If the configFile does not define these values, use defaults.
	if options.Port == 0 {
		options.Port = 9443
	}
	if options.MetricsBindAddress == "" {
		options.MetricsBindAddress = metricsAddr
	}
	if options.HealthProbeBindAddress == "" {
		options.HealthProbeBindAddress = probeAddr
	}

	// Create context for goroutine cleanup
	ctx := ctrl.SetupSignalHandler()

	// Initialize interface with greymatter CLI
	gmcli, err := cli.New(ctx)
	if err != nil {
		os.Exit(1)
	}

	// Create a rest.Config that has settings for communicating with the K8s cluster.
	restConfig := ctrl.GetConfigOrDie()

	// Create a write+read client for making requests to the API server.
	c, err := ctrlclient.New(restConfig, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		logger.Error(err, "Unable to create initial client")
	}

	// Set default image pull secret name in bootstrap config.
	if cfg.ImagePullSecretName == "" {
		cfg.ImagePullSecretName = "gm-docker-secret"
	}

	// Initialize installer
	inst, err := installer.New(c, gmcli, cfg.ImagePullSecretName)
	if err != nil {
		os.Exit(1)
	}

	// Initialize manager with configured options
	mgr, err := ctrl.NewManager(restConfig, options)
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Register the webhook handlers with our server to receive requests
	webhooks.Register(mgr, inst, gmcli, c)

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := mgr.Start(ctx); err != nil {
		logger.Error(err, "problem running manager")
		os.Exit(1)
	}
}
