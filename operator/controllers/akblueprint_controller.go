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
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// akmv1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	akmv1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
)

// AkBlueprintReconciler reconciles a AkBlueprint object
type AkBlueprintReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=akblueprints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=akblueprints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=akblueprints/finalizers,verbs=update

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

	// GET CRD
	crd := &akmv1alpha1.AkBlueprint{}
	err := r.Get(ctx, req.NamespacedName, crd)

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

	name := fmt.Sprintf("bp-%v-%v", crd.Namespace, crd.Name)
	cmWant, err := r.configForBlueprint(crd, name, ns)
	if err != nil {
		return ctrl.Result{}, err
	}

	// GET CONFIGMAP
	cm := &corev1.ConfigMap{}
	l.Info(fmt.Sprintf("Searching for configmap %v in %v", name, ns))
	err = r.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, cm)

	if err != nil && errors.IsNotFound(err) {
		// configmap was not found rety and notify the user
		l.Info(fmt.Sprintf("Not found. Creating configmap `%v` in `%v`", name, ns))
		err = r.Create(ctx, cmWant)
		if err != nil {
			l.Error(err, fmt.Sprintf("Failed to create configmap `%v` in `%v`", name, ns))
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// something went wrong with fetching the config map could be fatal
		l.Error(err, fmt.Sprintf("Failed to get configmap `%v` in `%v`", name, ns))
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found configmap %v in %v", name, ns))

	//check configmap matches what we want it to be by updating it
	r.Update(ctx, cmWant)
	if err != nil {
		// something went wrong with updating the deployment
		l.Error(err, fmt.Sprintf("Failed to update configmap %v in %v", name, ns))
		return ctrl.Result{}, err
	}

	// GET DEPLOYMENT
	dep := &appsv1.Deployment{}
	// instantiating minimal namespacedname to use in searching the kubeapi for the deployment
	depSearch := types.NamespacedName{
		Namespace: ns,
		Name:      wn,
	}
	l.Info(fmt.Sprintf("Searching for deployment %v in %v", depSearch.Name, depSearch.Namespace))
	err = r.Get(ctx, depSearch, dep)

	if err != nil && errors.IsNotFound(err) {
		// if deployment cannot be found
		l.Error(err, fmt.Sprintf("Authentik worker deployment `%v` not found in namespace `%v` but required, retrying", depSearch.Name, depSearch.Namespace))
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// if there was some failure in searching for deployment
		l.Error(err, "Failed to get Authentik worker deployment", depSearch.Name, "in", depSearch.Namespace)
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found deployment %v in %v", depSearch.Name, depSearch.Namespace))

	// create a copy of the spec for us to modify with what we want it to be
	depWant := *dep
	depWant.Namespace = ns
	depWant.Name = wn

	// GET VOLUME
	// first check if volume is already present
	// https://github.com/kubernetes/api/blob/master/core/v1/types.go#L36
	volWant := r.volumeForConfig(cm, filepath.Base(filepath.Clean(crd.Spec.File)))
	volIsFound := false
	for i, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.Name == cmWant.Name {
			l.Info(fmt.Sprintf("Existing blueprint volume: %v: %T found reconciling", vol, vol))
			depWant.Spec.Template.Spec.Volumes[i] = *volWant
			volIsFound = true
		}
	}
	if volIsFound == false {
		l.Info(fmt.Sprintf("Volume for configmap not found appending"))
		depWant.Spec.Template.Spec.Volumes = append(depWant.Spec.Template.Spec.Volumes, *volWant)
	}

	// GET MOUNT
	mountWant := r.mountForVolume(volWant, filepath.Dir(filepath.Clean(crd.Spec.File)))
	mountIsFound := false
	for i, cont := range dep.Spec.Template.Spec.Containers {
		for j, mount := range cont.VolumeMounts {
			if mount.Name == mountWant.Name {
				l.Info(fmt.Sprintf("VolumeMount found container: %v (%v), volMount: %v %+v", i, cont.Name, j, mount))
				depWant.Spec.Template.Spec.Containers[i].VolumeMounts[j] = *mountWant
				mountIsFound = true
			}
		}
		if mountIsFound == false {
			l.Info(fmt.Sprintf("VolumeMount for volume not found creating, cont:%v (%v)", i, cont.Name))
			depWant.Spec.Template.Spec.Containers[i].VolumeMounts = append(depWant.Spec.Template.Spec.Containers[i].VolumeMounts, *mountWant)
		}
	}

	// APPLY
	// ctrl.SetControllerReference(crd, &depWant, r.Scheme)
	err = r.Update(ctx, &depWant)
	if err != nil {
		// something went wrong with updating the deployment
		l.Error(err, fmt.Sprintf("Failed to update deployment %v in %v", dep.Name, dep.Namespace))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// mountForVolume creates a volumeMount inside a pod from a Volume specification which was generated from a configmap which was generated froma blueprint. This takes the volume and gives back a volume mount for this volume.
func (r *AkBlueprintReconciler) mountForVolume(crd *corev1.Volume, basePath string) *corev1.VolumeMount {
	volMount := corev1.VolumeMount{
		Name:      crd.Name,
		MountPath: filepath.Join(basePath, crd.VolumeSource.ConfigMap.Items[0].Path),
		SubPath:   crd.VolumeSource.ConfigMap.Items[0].Path,
	}
	return &volMount
}

// volumeForConfig creates a Volume to be added to add to the array of volums that holds a desired configmap which itself is generated from a blueprint. This takes the blueprint-configmap and the key to use inside the configmap to create the volume.
func (r *AkBlueprintReconciler) volumeForConfig(crd *corev1.ConfigMap, key string) *corev1.Volume {
	k2p := make([]corev1.KeyToPath, 1)
	k2p[0] = corev1.KeyToPath{
		Key:  key,
		Path: key,
	}
	volSpec := &corev1.ConfigMapVolumeSource{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: crd.Name,
		},
		Items: k2p,
	}
	vol := corev1.Volume{
		Name: crd.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: volSpec,
		},
	}
	return &vol
}

// configForBlueprint generates a configmap spec from a given blueprint that contains the blueprint data as a kube-native configmap to mount into our deployment later.
func (r *AkBlueprintReconciler) configForBlueprint(crd *akmv1alpha1.AkBlueprint, name string, namespace string) (*corev1.ConfigMap, error) {
	// create the map of key values for the data in configmap from blueprint contents
	cleanFP := filepath.Clean(crd.Spec.File)
	var dataMap = make(map[string]string)
	// set the key to be the filename and extension from path
	// set data to be the blueprint string
	dataMap[filepath.Base(cleanFP)] = crd.Spec.Blueprint

	// create annotation for destination path
	var annMap = make(map[string]string)
	annMap["akm.goauthentik/v1alpha1"] = filepath.Dir(cleanFP)

	cm := corev1.ConfigMap{
		// Metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annMap,
		},
		Data: dataMap,
	}
	// set that we are controlling this resource
	ctrl.SetControllerReference(crd, &cm, r.Scheme)
	return &cm, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkBlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1alpha1.AkBlueprint{}).
		Complete(r)
}
