package controllers

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sso "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	ssov1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
)

// AkWorkerReconciler reconciles a AkWorker object
type AkWorkerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akworkers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akworkers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=akworkers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AkWorker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AkWorkerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	ns := os.Getenv("AUTHENTIK_MANAGER_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	// GET CRD
	crd := &sso.AkWorker{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("AkWorker resource changed but disappeared. Ignoring since object must have been deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get AkWorker resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found AkWorker resource `%v` in `%v`.", crd.Name, crd.Namespace))

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ssov1alpha1.AkWorker{}).
		Complete(r)
}
