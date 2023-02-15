/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sso "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
)

// AkReconciler reconciles a Ak object
type AkReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=aks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=aks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sso.goauthentik.io,resources=aks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ak object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	ns := os.Getenv("AUTHENTIK_MANAGER_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	// GET CRD
	crd := &sso.Ak{}
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

	// check to ensure that the namespace is the same as the expected  namespace that all other resources will end up using.
	// otherwise we could have a scenario where this creates everything fine, but then nothing can use it as the other resources,
	// are trying to look for it in the wrong namespace. So we stop it here as an early notification that something is wrong with
	// the namespace environment variable
	if crd.Namespace != ns {
		l.Error(err, fmt.Sprintf("Ak resource `%v` in `%v` is not in the expected namespace `%v` (AUTHENTIK_MANAGER_NAMESPACE) ignoring.", crd.Name, crd.Namespace, ns))
		return ctrl.Result{}, err
	}

	// TODO: at the moment we assume tyranny implement more harmonious republic

	// Generate, search, and update server resource from generic ak resource
	server := &sso.AkServer{}
	serverWant := &sso.AkServer{
		Spec: sso.AkServerSpec{},
	}
	serverWant.Namespace = crd.Namespace
	serverWant.Name = fmt.Sprintf("%v-%v", crd.Spec.Naming.Base, crd.Spec.Naming.Server)
	serverSearch := types.NamespacedName{
		Namespace: serverWant.Namespace,
		Name:      serverWant.Name,
	}
	err = r.Get(ctx, serverSearch, server)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, serverWant)
			if err != nil {
				l.Error(err, fmt.Sprintf("Failed to create AkServer %v in %v", serverWant.Name, serverWant.Namespace))
				return ctrl.Result{}, err
			}
		} else {
			l.Error(err, "Failed to get AkServer. Likely fetch error. Retrying.")
			return ctrl.Result{}, err
		}
	} else {
		server.Spec = serverWant.Spec
		err = r.Update(ctx, server)
		if err != nil {
			l.Error(err, fmt.Sprintf("Failed to update AkServer %v in %v", serverWant.Name, serverWant.Namespace))
			return ctrl.Result{}, err
		}
	}

	// Generate worker resource from generic ak resource
	worker := &sso.AkWorker{}
	workerWant := &sso.AkWorker{
		Spec: sso.AkWorkerSpec{},
	}
	workerWant.Namespace = crd.Namespace
	workerWant.Name = fmt.Sprintf("%v-%v", crd.Spec.Naming.Base, crd.Spec.Naming.Worker)
	workerSearch := types.NamespacedName{
		Namespace: workerWant.Namespace,
		Name:      workerWant.Name,
	}
	err = r.Get(ctx, workerSearch, worker)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, workerWant)
			if err != nil {
				l.Error(err, fmt.Sprintf("Failed to create AkWorker %v in %v", workerWant.Name, workerWant.Namespace))
				return ctrl.Result{}, err
			}
		} else {
			l.Error(err, "Failed to get AkWorker. Likely fetch error. Retrying.")
			return ctrl.Result{}, err
		}
	} else {
		worker.Spec = workerWant.Spec
		err = r.Update(ctx, worker)
		if err != nil {
			l.Error(err, fmt.Sprintf("Failed to update AkWorker %v in %v", workerWant.Name, workerWant.Namespace))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AkReconciler) genServer() *sso.AkServer {
	return &sso.AkServer{}
}

func (r *AkReconciler) genWorker() *sso.AkWorker {
	return &sso.AkWorker{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sso.Ak{}).
		Complete(r)
}
