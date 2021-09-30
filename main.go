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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1alpha1 "github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/bootstrap"
	"github.com/greymatter-io/operator/pkg/clients"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// Add tags here for generating RBAC rules for the role that will be used by the Operator.
//+kubebuilder:rbac:groups=greymatter.io,resources=meshes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=greymatter.io,resources=meshes/status,verbs=get;update;patch

func main() {
	var configFile string
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	var development bool

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
		Scheme:         scheme,
		LeaderElection: enableLeaderElection,
		// TODO: Generate hash for ID; should be unique with multiple operator replicas
		LeaderElectionID: "715805a0.greymatter.io",
	}

	// Attempt to read a configFile if one has been configured.
	cfg := bootstrap.BootstrapConfig{}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&cfg))
		if err != nil {
			setupLog.Error(err, "unable to load config file", "path", configFile)
			os.Exit(1)
		} else {
			setupLog.Info("Loaded bootstrap config", "Path", configFile)
		}
	}

	// Set defaults for BootstrapConfig values
	if cfg.ImagePullSecretName == "" {
		cfg.ImagePullSecretName = "gm-docker-secret"
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

	// TODO: Move this further down; it's here for now because ctrl.NewManager requires talking to a K8s cluster.
	_, err = clients.New()
	if err != nil {
		setupLog.Error(err, "unable to initialize clients")
		os.Exit(1)
	}

	// Initialize manager with configured options
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
