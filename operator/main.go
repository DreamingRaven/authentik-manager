/*
Copyright 2023.

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
	"encoding/json"
	"fmt"
	"os"

	arg "github.com/alexflint/go-arg"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	akmv1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(akmv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// Opts options struct for the operator to autopopulate help templates, autogenerate options, and ensure consistency between env and cli.
type Opts struct {
	MetricsAddr          string `arg:"--metrics-bind-address" default:":8080" json:"metricsAddr,omitempty" help:"The address the metric endpoint binds to."`
	LeaderElectionID     string `arg:"--leader-election-id" default:"d460f2c2.goauthentik.io" json:"leaderElectionID,omitempty" help:"Lease name to use for leader election."`
	WatchesPath          string `arg:"--watches-file" default:"watches.yaml" json:"watchesPath,omitempty" help:"Path to watches file."`
	ProbeAddr            string `arg:"--health-probe-bind-address" default:":8081" json:"probeAddr,omitempty" help:"The address the probe endpoint binds to."`
	EnableLeaderElection bool   `arg:"--leader-elect" json:"enableLeaderElection,omitempty" help:"To elect a leader to be active else all active."`
	OperatorNamespace    string `arg:"--operator-namespace" default:"auth" json:"operatorNamespace,omitempty" help:"The operators namespace for leader election."`
	WatchedNamespace     string `arg:"--watched-namespace" default:"" json:"watchedNamespace,omitempty" help:"The operators watched namespace. Defaults to empty (which watches all)."`
	Debug                bool   `arg:"-d,--debug" json:"debug,omitempty" help:"We should run in debug mode."`
	Port                 int    `arg:"-p,--port" default:"9443" json:"port,omitempty" help:"What port should the controller bind to."`
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func main() {
	o := Opts{}
	arg.MustParse(&o)

	if o.Debug {
		fmt.Println(prettyPrint(o))
	}

	opts := zap.Options{
		Development: o.Debug,
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     o.MetricsAddr,
		Port:                   o.Port,
		HealthProbeBindAddress: o.ProbeAddr,
		LeaderElection:         o.EnableLeaderElection,
		LeaderElectionID:       o.LeaderElectionID,
		// Specified the namespace the leader "lease" resource belongs
		// this will also affect clusterwide searches by operator which we dont want
		// so we specify so that these roles do not need to be granted
		LeaderElectionNamespace: o.OperatorNamespace,
		Namespace:               o.WatchedNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.AkReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ak")
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
