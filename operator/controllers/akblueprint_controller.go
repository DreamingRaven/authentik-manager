package controllers

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
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

	ns := os.Getenv("AUTHENTIK_MANAGER_NAMESPACE")
	if ns == "" {
		ns = "default"
	}
	wn := os.Getenv("AUTHENTIK_WORKER_NAME")
	if wn == "" {
		wn = "authentik-worker"
	}

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

	// configmap name is a composite of the namespace the blueprint is from and the blueprint name itself with a bp suffix
	name := fmt.Sprintf("bp-%v-%v", crd.Namespace, crd.Name)
	// fetch from kubeapi the current state of the configmap
	cm := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, cm)
	// create an object that is what we would like the config map to be
	cmWant := r.configForBlueprint(crd, name, ns)
	// check that the configmap is what we expected and no errors
	if err != nil && errors.IsNotFound(err) {
		// configmap was not found rety and notify the user
		l.Info(fmt.Sprintf("AkBlueprint's configmap `%v` not found in namespace `%v` but desired, reconciling", cmWant.Name, cmWant.Namespace))
		r.Create(ctx, cmWant)
		l.Info(fmt.Sprintf("AkBlueprint's configmap `%v` successfully created  in `%v`", cmWant.Name, cmWant.Namespace))
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// something went wrong with fetching the config map could be fatal
		l.Error(err, "Failed to get ConfigMap", name, "in", crd.Namespace)
		return ctrl.Result{}, err
	}
	//TODO: check configmap matches what we want it to be

	// get authentik worker deployment by name
	dep := &appsv1.Deployment{}
	depWant := types.NamespacedName{
		Namespace: ns,
		Name:      wn,
	}
	err = r.Get(ctx, depWant, dep)

	if err != nil && errors.IsNotFound(err) {
		// if deployment cannot be found
		l.Error(err, fmt.Sprintf("Authentik worker deployment `%v` not found in namespace `%v` but required, retrying", depWant.Name, depWant.Namespace))
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// if there was some failure in searching for deployment
		l.Error(err, "Failed to get Authentik worker deployment", name, "in", crd.Namespace)
		return ctrl.Result{}, err
	}
	// first check if volume is already present
	// https://github.com/kubernetes/api/blob/master/core/v1/types.go#L36
	volWant := &corev1.Volume{
		Name: name,
	}
	// volWant := &corev1.ConfigMapVolumeSource{
	// 	Name: name,
	// 	// Items: []corev1.KeyToPath,
	// }
	fmt.Println(volWant)
	for i, vol := range dep.Spec.Template.Spec.Volumes {
		l.Info(fmt.Sprintf("volume: %v: %T", i, vol))
		if vol.Name == name {
			l.Info(fmt.Sprintf("existing blueprint volume: %v: %T found validating", vol, vol))
			// return ctrl.Result{}, nil
		}
	}
	// volume was not found so create it and requeue
	// TODO: ensure deployment matches what we want with volume + mount of prior configmap

	// fmt.Print(dep.Spec.Template.Spec.Volumes)

	return ctrl.Result{}, nil
}

func (r *AkBlueprintReconciler) configForBlueprint(crd *ssov1alpha1.AkBlueprint, name string, namespace string) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		// Metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
