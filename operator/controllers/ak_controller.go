/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	chartLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils"
)

// AkReconciler reconciles a Ak object
type AkReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=aks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=aks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=aks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := klog.FromContext(ctx)

	// Parsing options to make them available TODO: pass them in rather than read continuously
	o := utils.Opts{}
	arg.MustParse(&o)

	actionConfig, err := r.GetActionConfig(req.NamespacedName.Namespace, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	// GET CRD
	crd := &akmv1a1.Ak{}
	err = r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("Ak resource reconciliation triggered but disappeared. Checking for residual chart for uninstall then ignoring since object must have been deleted.")
			_, err := r.UninstallChart(req.NamespacedName, actionConfig)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get Ak resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found Ak resource `%v` in `%v`.", crd.Name, crd.Namespace))

	// Helm Chart Identification
	u, err := url.Parse(fmt.Sprintf("file://workspace/helm-charts/ak-%v.tgz", o.SrcVersion))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Helm Install or Upgrade Chart
	var vals map[string]interface{}
	err = json.Unmarshal(crd.Spec.Values, &vals)
	if err != nil {
		return ctrl.Result{}, err
	}
	_, err = r.UpgradeOrInstallChart(req.NamespacedName, u, actionConfig, vals)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// UpgradeOrInstallChart upgrades a chart in cluster or installs it new if it does not already exist
// ulr format is [scheme:][//[userinfo@]host][/]path[?query][#fragment] e.g file://workspace/helm-charts/ak-0.1.0.tgz"
func (r *AkReconciler) UpgradeOrInstallChart(nn types.NamespacedName, u *url.URL, a *action.Configuration, o map[string]interface{}) (*release.Release, error) {
	// Helm List Action
	listAction := action.NewList(a)
	releases, err := listAction.Run()
	if err != nil {
		return nil, err
	}

	toUpgrade := false
	for _, release := range releases {
		// fmt.Println("Release: " + release.Name + " Status: " + release.Info.Status.String())
		if release.Name == nn.Name {
			toUpgrade = true
		}
	}

	c, err := r.LoadHelmChart(u)
	if err != nil {
		return nil, err
	}

	fmt.Println(o)

	var rel *release.Release
	if toUpgrade {
		// Helm Upgrade
		updateAction := action.NewUpgrade(a)
		rel, err = updateAction.Run(nn.Name, c, o)
		if err != nil {
			return nil, err
		}

	} else {
		// Helm Install
		installAction := action.NewInstall(a)
		installAction.Namespace = nn.Namespace
		installAction.ReleaseName = nn.Name
		rel, err = installAction.Run(c, o)
		if err != nil {
			return nil, err
		}
	}
	return rel, nil
}

func (r *AkReconciler) UninstallChart(nn types.NamespacedName, a *action.Configuration) (*release.UninstallReleaseResponse, error) {
	uninstallAction := action.NewUninstall(a)
	releaseResponse, err := uninstallAction.Run(nn.Name)
	if err != nil {
		return nil, err
	}
	return releaseResponse, nil
}

// GetActionConfig Get the Helm action config from in cluster service account
func (r *AkReconciler) GetActionConfig(namespace string, l logr.Logger) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	var kubeConfig *genericclioptions.ConfigFlags
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Set properties manually from official rest config
	kubeConfig = genericclioptions.NewConfigFlags(false)
	kubeConfig.APIServer = &config.Host
	kubeConfig.BearerToken = &config.BearerToken
	kubeConfig.CAFile = &config.CAFile
	kubeConfig.Namespace = &namespace
	if err := actionConfig.Init(kubeConfig, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {

	}
	return actionConfig, nil
}

// Get Connection Client to Kubernetes
func (r *AkReconciler) GetKubeClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// GetHelmChart loads a helm chart from a given file as URL
func (r *AkReconciler) LoadHelmChart(u *url.URL) (*chart.Chart, error) {
	// fmt.Println("Scheme:", u.Scheme)
	// fmt.Println("Opaque:", u.Opaque)
	// fmt.Println("User:", u.User)
	// fmt.Println("Host:", u.Host)
	// fmt.Println("Path:", u.Path)
	// fmt.Println("RawPath:", u.RawPath)
	// fmt.Println("ForceQuery:", u.ForceQuery)
	// fmt.Println("RawQuery:", u.RawQuery)
	// fmt.Println("Fragment:", u.Fragment)
	// fmt.Println("RawFragment:", u.RawFragment)

	// GET HELM CHART
	if u.Scheme != "file" {
		err := errors.NewInvalid(
			schema.GroupKind{
				Group: "akm.goauthentik.io",
				Kind:  "Ak",
			},
			fmt.Sprintf("Url scheme `%v` != `file`, unsupported scheme.", u.Scheme),
			field.ErrorList{})
		return nil, err
	}
	// load chart from filepath (which is part of host in url)
	path, err := filepath.Abs(u.Host + u.Path)
	if err != nil {
		return nil, err
	}
	_, err = utils.Exists(path)
	if err != nil {
		return nil, err
	}
	return chartLoader.Load(path)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1a1.Ak{}).
		Complete(r)
}
