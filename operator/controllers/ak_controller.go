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
	"net/url"
	"time"

	"github.com/alexflint/go-arg"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils"
)

// AkReconciler reconciles a Ak object
type AkReconciler struct {
	utils.ControlBase
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
		t, _ := time.ParseDuration("10s")
		return ctrl.Result{Requeue: true, RequeueAfter: t}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1a1.Ak{}).
		Complete(r)
}
