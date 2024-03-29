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
	"time"

	"github.com/alexflint/go-arg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"   // Required for watching
	"sigs.k8s.io/controller-runtime/pkg/client"    // Required for watching
	"sigs.k8s.io/controller-runtime/pkg/handler"   // Required for watching
	klog "sigs.k8s.io/controller-runtime/pkg/log"  // Required for watching
	"sigs.k8s.io/controller-runtime/pkg/predicate" // Required for watching
	"sigs.k8s.io/controller-runtime/pkg/reconcile" // Required for watching
	"sigs.k8s.io/yaml"

	// Required for watching
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

	actionConfig, err := r.GetActionConfig(req.NamespacedName.Namespace)
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

	// GET FILE-BASED BLUEPRINTS LIST
	l.Info(fmt.Sprintf("Searching for blueprint configs in `%v`.", o.OperatorNamespace))
	configs := &corev1.ConfigMapList{}
	err = r.List(ctx, configs,
		client.InNamespace(o.OperatorNamespace),
		client.MatchingLabels{"akm.goauthentik.io/type": "blueprint"})
	if err != nil {
		return ctrl.Result{}, err
	}

	// HELM OVERRIDES LOAD
	var vals map[string]interface{}
	err = yaml.Unmarshal(crd.Spec.Values, &vals)
	if err != nil {
		return ctrl.Result{}, err
	}

	// OVERRIDE FILE-BASED BLUEPRINTS TO HELM
	// blueprints in the helm chart should be a list under .Values.authentik.blueprints
	// - name: some-default-blueprint
	//   dest: /blueprints/default/some-default-blueprint.yaml
	//   configMap:
	//     name: example-custom-blueprint-configmap
	//     key: my-default-blueprint
	var configBps []map[string]interface{}
	for i, config := range configs.Items {
		count := 0
		for j, data := range config.Data {
			bp := &akmv1a1.BP{}
			err = yaml.Unmarshal([]byte(data), bp)
			if err != nil {
				return ctrl.Result{}, err
			}
			l.Info(fmt.Sprintf("Capturing bpConfig: `%v`(%v), `%v` at `%v`)", config.Name, i, bp.Metadata.Name, j))
			// TODO: add checks to ensure things like annotation path actually exists
			configBps = append(configBps, map[string]interface{}{
				"name": fmt.Sprintf("%v-%v", config.Name, count), // I would have liked to use `j` instead of `count` but there is a char limit,
				"dest": fmt.Sprintf("%v/%v", config.Annotations["akm.goauthentik.io/path"], j),
				"configMap": map[string]interface{}{
					"name": config.Name,
					"key":  j,
				},
			})
			count = count + 1
		}
	}
	configBpsAsValues := map[string]interface{}{
		"authentik": map[string]interface{}{
			"blueprints": configBps,
		},
	}
	vals = utils.MergeDicts(vals, configBpsAsValues)
	// Final Adjustments and Overrides
	//TODO inherit value from operator not deployed CRD
	vals["instanceOverride"] = crd.Labels["app.kubernetes.io/instance"]
	fmt.Println(utils.PrettyPrint(vals))

	// HELM INSTALL OR UPGRADE
	_, err = r.UpgradeOrInstallChart(req.NamespacedName, u, actionConfig, vals)
	if err != nil {
		t, _ := time.ParseDuration("10s")
		return ctrl.Result{Requeue: true, RequeueAfter: t}, err
	}

	return ctrl.Result{}, nil
}

// findAkForConfigMap finds the specific Ak resource context that needs to be passed to the reconciler
// when the reconciliation is triggered by a configmap change rather than Ak resource directly.
func (r *AkReconciler) findAkForConfigMap(ctx context.Context, configMap client.Object) []reconcile.Request {
	// Parsing options to make them available TODO: pass them in rather than read continuously
	o := utils.Opts{}
	arg.MustParse(&o)

	aks := &akmv1a1.AkList{}
	opts := &client.ListOptions{
		Namespace: o.OperatorNamespace,
	}
	err := r.List(ctx, aks, opts)
	if err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(aks.Items))
	for i, item := range aks.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// PARSE OPTIONS
	// TODO: pass them in rather than read continuously
	o := utils.Opts{}
	arg.MustParse(&o)
	labelPredicate, err := predicate.LabelSelectorPredicate(
		*metav1.AddLabelToSelector(
			&metav1.LabelSelector{}, "akm.goauthentik.io/type", "blueprint",
		),
	)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1a1.Ak{}).
		// dont let the docs lie to you about what Watches supports
		//WatchesRawSource(
		Watches(
			// watch for configmaps
			//source.Kind(mgr.GetCache(), &corev1.ConfigMap{}),
			&corev1.ConfigMap{},
			// set request to be the relevant Ak resource for that configmap
			handler.EnqueueRequestsFromMapFunc(r.findAkForConfigMap),
			// when the resource has passes these predicates:
			builder.WithPredicates(
				// - resource version has changed
				//predicate.ResourceVersionChangedPredicate{},
				// - resource is in the namespace of the operator
				utils.NamespacePredicate{Namespace: o.OperatorNamespace},
				// - resource has the correct label to mark it as blueprint config
				labelPredicate,
			),
		).
		Complete(r)
}

//Watches(source.Source, handler.EventHandler, ...)
//Watches(source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueuRequestForObject{}).
//https://yash-kukreja-98.medium.com/develop-on-kubernetes-series-demystifying-the-for-vs-owns-vs-watches-controller-builders-in-c11ab32a046e
//https://nakamasato.medium.com/kubernetes-operator-series-4-controller-runtime-component-builder-c649c0ad2dc0
// For(&corev1alpha1.MyCustomResource{}) == Watches(&source.Kind{Type: &corev1alpha1.MyCustomResource{}}, &handler.EnqueueRequestForObject{})
// Owns(&corev1.Pod{}) == Watches(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &corev1alpha1.MyCustomResource{}})
