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
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// "helm.sh/helm/v3/pkg/chart"
	// chartLoader "helm.sh/helm/v3/pkg/chart/loader"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
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
	l := log.FromContext(ctx)

	// GET CRD
	crd := &akmv1a1.Ak{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("Ak resource changed but disappeared. Ignoring since object must have been deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get Ak resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found Ak resource `%v` in `%v`.", crd.Name, crd.Namespace))

	// GET SOURCE HELM CHART
	// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
	u, err := url.Parse("file://somefile.tar.gz")
	if err != nil {
		return ctrl.Result{}, err
	}
	fmt.Println(u)

	// c, err := r.GetHelmChart(u)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }
	// fmt.Println(c)

	// Authenticate to Kubernetes
	// https://stackoverflow.com/questions/66730436/how-to-connect-to-kubernetes-cluster-using-serviceaccount-token

	return ctrl.Result{}, nil
}

// func (r *AkReconciler) GetHelmChart(u *url.URL) (*chart.Chart, error) {
// 	fmt.Println("Scheme:", u.Scheme)
// 	fmt.Println("Opaque:", u.Opaque)
// 	fmt.Println("User:", u.User)
// 	fmt.Println("Host:", u.Host)
// 	fmt.Println("Path:", u.Path)
// 	fmt.Println("RawPath:", u.RawPath)
// 	fmt.Println("ForceQuery:", u.ForceQuery)
// 	fmt.Println("RawQuery:", u.RawQuery)
// 	fmt.Println("Fragment:", u.Fragment)
// 	fmt.Println("RawFragment:", u.RawFragment)
// 	// GET HELM CHART
// 	c, err := chartLoader.Load(u.Path)
// 	if err != nil {
// 		return c, err
// 	}
// 	return c, err
// }

// SetupWithManager sets up the controller with the Manager.
func (r *AkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1a1.Ak{}).
		Complete(r)
}
