package controllers

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sso "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
)

// AkServerReconciler reconciles a AkServer object
type AkServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AkServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	ns := os.Getenv("AUTHENTIK_MANAGER_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	// GET CRD
	crd := &sso.AkServer{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("AkServer resource changed but disappeared. Ignoring since object must have been deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get AkServer resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found AkServer resource `%v` in `%v`.", crd.Name, crd.Namespace))

	dep := &appsv1.Deployment{}
	depWant := r.genDeploy(crd)
	depWant.Namespace = crd.Namespace
	depWant.Name = crd.Name
	ctrl.SetControllerReference(crd, depWant, r.Scheme)
	depSearch := types.NamespacedName{
		Namespace: depWant.Namespace,
		Name:      depWant.Name,
	}
	err = r.Get(ctx, depSearch, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, depWant)
			if err != nil {
				l.Error(err, fmt.Sprintf("Failed to create AkServer %v in %v", depWant.Name, depWant.Namespace))
				return ctrl.Result{}, err
			}
		} else {
			l.Error(err, "Failed to get AkServer. Likely fetch error. Retrying.")
			return ctrl.Result{}, err
		}
	} else {
		dep.Spec = depWant.Spec
		err = r.Update(ctx, dep)
		if err != nil {
			l.Error(err, fmt.Sprintf("Failed to update AkServer %v in %v", depWant.Name, depWant.Namespace))
			return ctrl.Result{}, err
		}
	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// genDeploy generate an authentik server deployment from the CRD we are given.
func (r *AkServerReconciler) genDeploy(crd *sso.AkServer) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crd.Name,
			Namespace: crd.Namespace,
		},
		Spec: appsv1.DeploymentSpec{},
	}
	return deploy
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sso.AkServer{}).
		Complete(r)
}
