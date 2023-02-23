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
	"flag"
	"os"
	"runtime"
	"time"

	arg "github.com/alexflint/go-arg"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	"github.com/operator-framework/helm-operator-plugins/pkg/annotation"
	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler"
	"github.com/operator-framework/helm-operator-plugins/pkg/watches"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrlruntime "k8s.io/apimachinery/pkg/runtime"
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
	scheme                         = ctrlruntime.NewScheme()
	setupLog                       = ctrl.Log.WithName("setup")
	defaultMaxConcurrentReconciles = runtime.NumCPU()
	defaultReconcilePeriod         = time.Minute
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(akmv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// // Opts option struct for this operator.
// // Optins are populated by a seperate function, that may or may not use the struct tags to reflect and auto-populate fields.
// type Opts struct {
// 	MetricsAddr          string `json:"metricsAddr,omitempty" opts:"metrics-bind-address,:8080,Address the metrics endpoint binds to."`
// 	LeaderElectionID     string `json:"leaderElectionID,omitempty" opts:"leader-election-id,d460f2c2.goauthentik.io,Leader election lease name."`
// 	WatchesPath          string `json:"watchesPath,omitempty" opts:"watches-file,watches.yaml,Path to watches file."`
// 	ProbeAddr            string `json:"probeAddr,omitempty" opts:"health-probe-bind-address,:8081,Address the probe endpoint binds to."`
// 	EnableLeaderElection bool   `json:"enableLeaderElection,omitempty" opts:"leader-elect,false,Should we try to elect a leader."`
// 	OperatorNamespace    string `json:"operatorNamespace,omitempty" opts:"operator-namespace,auth,The operators namespace for leader election."`
// 	WatchedNamespace     string `json:"watchedNamespace,omitempty" opts:"watched-namespace,,The operators watched namespace. Defaults to empty (which watches all)."`
// }
//
// func NewOpts() {
// 	o := Opts{}
// 	fields := reflect.VisibleFields(reflect.TypeOf(o))
// 	for _, field := range fields {
// 		fmt.Println(fmt.Sprintf("Field: `%v`, Type: `%v`", field.Name, field.Type))
//
// 	}
//
// }

type Opts struct {
	MetricsAddr          string `arg:"--metrics-bind-address" default:":8080" json:"metricsAddr,omitempty" help:"The address the metric endpoint binds to."`
	LeaderElectionID     string `arg:"--leader-election-id" default:"d460f2c2.goauthentik.io" json:"leaderElectionID,omitempty" help:"Lease name to use for leader election."`
	WatchesPath          string `arg:"--watches-file" default:"watches.yaml" json:"watchesPath,omitempty" help:"Path to watches file."`
	ProbeAddr            string `arg:"--health-probe-bind-address" default:":8081" json:"probeAddr,omitempty" help:"The address the probe endpoint binds to."`
	EnableLeaderElection bool   `arg:"--leader-elect" default:"false" json:"enableLeaderElection,omitempty" help:"To elect a leader to be active else all active."`
	OperatorNamespace    string `arg:"--operator-namespace" default:"auth" json:"operatorNamespace,omitempty" help:"The operators namespace for leader election."`
	WatchedNamespace     string `arg:"--watched-namespace" default:"" json:"watchedNamespace,omitempty" help:"The operators watched namespace. Defaults to empty (which watches all)."`
}

func main() {
	o := Opts{}
	arg.MustParse(&o)

	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     o.MetricsAddr,
		Port:                   9443,
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

	if err = (&controllers.AkBlueprintReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AkBlueprint")
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

	ws, err := watches.Load(o.WatchesPath)
	if err != nil {
		setupLog.Error(err, "Failed to create new manager factories")
		os.Exit(1)
	}

	for _, w := range ws {
		// Register controller with the factory
		reconcilePeriod := defaultReconcilePeriod
		if w.ReconcilePeriod != nil {
			reconcilePeriod = w.ReconcilePeriod.Duration
		}

		maxConcurrentReconciles := defaultMaxConcurrentReconciles
		if w.MaxConcurrentReconciles != nil {
			maxConcurrentReconciles = *w.MaxConcurrentReconciles
		}

		r, err := reconciler.New(
			reconciler.WithChart(*w.Chart),
			reconciler.WithGroupVersionKind(w.GroupVersionKind),
			reconciler.WithOverrideValues(w.OverrideValues),
			reconciler.SkipDependentWatches(w.WatchDependentResources != nil && !*w.WatchDependentResources),
			reconciler.WithMaxConcurrentReconciles(maxConcurrentReconciles),
			reconciler.WithReconcilePeriod(reconcilePeriod),
			reconciler.WithInstallAnnotations(annotation.DefaultInstallAnnotations...),
			reconciler.WithUpgradeAnnotations(annotation.DefaultUpgradeAnnotations...),
			reconciler.WithUninstallAnnotations(annotation.DefaultUninstallAnnotations...),
		)
		if err != nil {
			setupLog.Error(err, "unable to create helm reconciler", "controller", "Helm")
			os.Exit(1)
		}
		if err := r.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Helm")
			os.Exit(1)
		}
		setupLog.Info("configured watch", "gvk", w.GroupVersionKind, "chartPath", w.ChartPath, "maxConcurrentReconciles", maxConcurrentReconciles, "reconcilePeriod", reconcilePeriod)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
