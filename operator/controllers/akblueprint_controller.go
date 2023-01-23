package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ssov1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
)

// AkBlueprintReconciler reconciles a AkBlueprint object
type AkBlueprintReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akblueprints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akblueprints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akblueprints/finalizers,verbs=update

func (r *AkBlueprintReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	// blank crd struct to populate
	crd := &ssov1alpha1.AkBlueprint{}
	// populating blank crd struct
	err := r.Get(ctx, req.NamespacedName, crd)

	// check if crd has been fetched correctly and early exit if not, or reque if there was some sort of request failure
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("AkBlueprint resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get AkBlueprint")
		return ctrl.Result{}, err
	}

	name := fmt.Sprintf("%v-%v", crd.Namespace, crd.Name)
	found := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: name, Namespace: crd.Namespace}, found)
	desire := r.configForBlueprint(crd, name)

	if err != nil && errors.IsNotFound(err) {
		l.Info(fmt.Sprintf("AkBlueprint's configmap `%v` not found in namespace `%v` but desired, reconciling", name, crd.Namespace))
		r.Create(ctx, desire)
		l.Info(fmt.Sprintf("AkBlueprint's configmap `%v` successfully created  in `%v`", name, crd.Namespace))
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		l.Error(err, "Failed to get ConfigMap", name, "in", crd.Namespace)
		return ctrl.Result{}, err
	}

	// check configmap still matches our crd

	// attatch configmap to deployment if not already by name

	return ctrl.Result{}, nil
}

func (r *AkBlueprintReconciler) configForBlueprint(crd *ssov1alpha1.AkBlueprint, name string) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		// Metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: crd.Namespace,
		},
	}
	// set that we are controlling this resource
	ctrl.SetControllerReference(crd, &cm, r.Scheme)
	return &cm
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkBlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ssov1alpha1.AkBlueprint{}).
		Complete(r)
}
