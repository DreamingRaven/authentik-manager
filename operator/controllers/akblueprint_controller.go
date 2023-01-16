package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AkBlueprint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
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
	l.Info("AkBlueprint found")

	// check blueprint validity

	// link blueprint if unlinked

	// update links

	// fmt.Printf("%v", crd)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkBlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ssov1alpha1.AkBlueprint{}).
		Complete(r)
}
