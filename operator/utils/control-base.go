// Package utils implements various utilities  for general use in our controllers.
package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	chartLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

// ControlBase struct centralises common controller functions into an embedded base struct
// to make the functions available with as little repetition as possible.
// https://stackoverflow.com/a/31505875
type ControlBase struct {
	client.Client
	Scheme *runtime.Scheme
}

// Control composes additional functionality we would like available to our controllers.
// This functionality is key to ensuring we KISS, and implements common routines
// like searching namespaces for resources or lists, along with common transformations.
// This does not include functions that do not require client or scheme context
// since those are better as standalone implementations rather than bundled routines.
type Control interface {}

// KUBERNETES routines

// ListInNamespace lists resources of given group, version, kind in the given namespace.
func (c *ControlBase) ListInNamespace() {}

// HELM routines

// GetReleaseValues gets the ACTUAL values used in a helm release by merging user values with manifests in the same way that we do when deploying the helm chart.
func (c *ControlBase) GetReleaseValues() {}

// UpgradeOrInstallChart upgrades a chart in cluster or installs it new if it does not already exist
// ulr format is [scheme:][//[userinfo@]host][/]path[?query][#fragment] e.g file://workspace/helm-charts/ak-0.1.0.tgz"
func (c *ControlBase) UpgradeOrInstallChart(nn types.NamespacedName, u *url.URL, a *action.Configuration, o map[string]interface{}) (*release.Release, error) {
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

	ch, err := c.LoadHelmChart(u)
	if err != nil {
		return nil, err
	}

	// fmt.Println(o)

	var rel *release.Release
	if toUpgrade {
		// Helm Upgrade
		updateAction := action.NewUpgrade(a)
		rel, err = updateAction.Run(nn.Name, ch, o)
		if err != nil {
			return nil, err
		}

	} else {
		// Helm Install
		installAction := action.NewInstall(a)
		installAction.Namespace = nn.Namespace
		installAction.ReleaseName = nn.Name
		rel, err = installAction.Run(ch, o)
		if err != nil {
			return nil, err
		}
	}
	return rel, nil
}

func (c *ControlBase) UninstallChart(nn types.NamespacedName, a *action.Configuration) (*release.UninstallReleaseResponse, error) {
	uninstallAction := action.NewUninstall(a)
	releaseResponse, err := uninstallAction.Run(nn.Name)
	if err != nil {
		return nil, err
	}
	return releaseResponse, nil
}

// GetActionConfig Get the Helm action config from in cluster service account
func (c *ControlBase) GetActionConfig(namespace string, l logr.Logger) (*action.Configuration, error) {
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
func (c *ControlBase) GetKubeClient() (*kubernetes.Clientset, error) {
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
func (c *ControlBase) LoadHelmChart(u *url.URL) (*chart.Chart, error) {
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
	//_, err = utils.Exists(path)
	_, err = Exists(path)
	if err != nil {
		return nil, err
	}
	return chartLoader.Load(path)
}

// NewSQLConfig best effort to generate a connection config based on env variables and system
func (c *ControlBase) NewSQLConfig() *SQLConfig {
	// TODO populate with real values from go-arg
	return &SQLConfig{
		Host:     "postgres",
		Port:     5432,
		User:     "postgres",
		Password: "MIwHsckSqhCli0KCEmq5RZDld744vP", // this is the password from example secret in docs docs
		DBName:   "authentik",
		SSLMode:  "disable",
	}
}
