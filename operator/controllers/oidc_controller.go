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

	"github.com/alexflint/go-arg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	akmv1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils"
)

// OIDCReconciler reconciles a OIDC object
type OIDCReconciler struct {
	utils.ControlBase
}

//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=oidcs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=oidcs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=oidcs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *OIDCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := klog.FromContext(ctx)

	// Parsing options to make them available TODO: pass them in rather than read continuously
	o := utils.Opts{}
	arg.MustParse(&o)

	// GET CRD
	crd := &akmv1a1.OIDC{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("OIDC resource reconciliation triggered but disappeared. Uninstalling OIDC integration.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get OIDC resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found OIDC resource `%v` in `%v`.", crd.Name, crd.Namespace))

	// PROVIDERS - generate secret and blueprint for each provider
	// secret contains clientID and clientSecret
	for i := range crd.Spec.Providers {
		provider := &crd.Spec.Providers[i]
		fmt.Printf("provider: %v\n", provider.Name)
		secret, err := spawnAndFetchOIDCSecret()
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("secret: %v\n", secret)
		provider_blueprint, err := reconcileProviderBlueprint()
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("blueprint: %v\n", provider_blueprint)

	}

	// APPLICATIONS - generate configmap and blueprint for each application
	// config contains urls for login, profile, logout, well-known, etc
	for i := range crd.Spec.Applications {
		application := &crd.Spec.Applications[i]
		fmt.Printf("application: %v\n", application.Name)
	}

	//// FETCH SECRET OR CREATE BUT NOT UPDATE
	//oidcSecret := &corev1.Secret{}
	//err = r.Get(ctx, types.NamespacedName{
	//	Name:      crd.Name,
	//	Namespace: crd.Namespace,
	//}, oidcSecret)
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		// Create secret as not found
	//		l.Info(fmt.Sprintf("Secret not found generating `%v` in `%v`.", crd.Name, crd.Namespace))
	//		charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//		if crd.Spec.ClientID == "" {
	//			crd.Spec.ClientID = utils.GenerateRandomString(40, charset)
	//		}
	//		if crd.Spec.ClientSecret == "" {
	//			crd.Spec.ClientSecret = utils.GenerateRandomString(128, charset)
	//		}
	//		oidcDesiredSecret := r.SecretFromOIDC(crd)
	//		err = r.Create(ctx, oidcDesiredSecret)
	//		if err != nil {
	//			return ctrl.Result{}, err
	//		}
	//		oidcSecret = oidcDesiredSecret
	//	} else {
	//		// Some error other than "not found" when trying to fetch secret
	//		return ctrl.Result{}, err
	//	}
	//} else {
	//	l.Info(fmt.Sprintf("Secret update ignored to prevent downtime for `%v` in `%v`.", oidcSecret.Name, oidcSecret.Namespace))
	//	crd.Spec.ClientID = string(oidcSecret.Data["clientID"])
	//	crd.Spec.ClientSecret = string(oidcSecret.Data["clientSecret"])
	//}

	//// GENERATE OR UPDATE BLUEPRINT
	//bps, err := r.BlueprintFromOIDC(crd)
	//if err != nil {
	//	return ctrl.Result{}, err
	//}
	//for _, bp := range bps {
	//	bp.Namespace = o.OperatorNamespace
	//	l.Info(fmt.Sprintf("Updating blueprint `%v` in `%v`", bp.Name, bp.Namespace))
	//	err = r.Update(ctx, bp)
	//	if err != nil {
	//		if errors.IsNotFound(err) {
	//			l.Info(fmt.Sprintf("Blueprint not found creating `%v` in `%v`", bp.Name, bp.Namespace))
	//			m, err := utils.PrettyPrint(bp)
	//			if err != nil {
	//				return ctrl.Result{}, err
	//			}
	//			fmt.Printf("bp: %v", m)
	//			err = r.Create(ctx, bp)
	//			if err != nil {
	//				return ctrl.Result{}, err
	//			}
	//		} else {
	//			l.Error(err, "Failed to update blueprint. Retrying.")
	//			return ctrl.Result{}, err
	//		}
	//	}
	//}

	// Generate secret and configmap for OIDC
	// Secret contains OIDC clientID and clientSecret
	// Configmap contains well-known configuration endpoints for the application to use

	//TODO: add live testing of OIDC by operator and locking system to prevent binding to non-functioning OIDC
	// GENERATE OR UPDATE INGRESS WELL-KNOWN
	//in := r.IngressFromOIDC(crd)
	//in.Namespace = o.OperatorNamespace
	//l.Info(fmt.Sprintf("Updating ingress `%v` in `%v`", in.Name, in.Namespace))
	//err = r.Update(ctx, in)
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		l.Info(fmt.Sprintf("Ingress not found creating `%v` in `%v`", bp.Name, bp.Namespace))
	//		errc := r.Create(ctx, in)
	//		if errc != nil {
	//			return ctrl.Result{}, errc
	//		}
	//	} else {
	//		l.Error(err, "Failed to update ingress. Retrying.")
	//		return ctrl.Result{}, err
	//	}
	//}

	return ctrl.Result{}, nil
}

// spawnAndFetchOIDCSecret creates a secret for a client application to use to register and identify itself using the client_id and client_secret within.
func spawnAndFetchOIDCSecret() (*corev1.Secret, error) {
	return &corev1.Secret{}, nil
}

// spawnAndFetchOIDCConfigmap creates a configmap to let the client know the relevant endpoints to use for OIDC
func spawnAndFetchOIDCConfigmap() (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{}, nil
}

func reconcileProviderBlueprint() (*akmv1a1.AkBlueprint, error) {
	return &akmv1a1.AkBlueprint{}, nil
}

func reconcileApplicationBlueprint() (*akmv1a1.AkBlueprint, error) {
	return &akmv1a1.AkBlueprint{}, nil
}

//// SecretFromOIDC creates a secret for a client application to use to register and identify itself.
//func (r *OIDCReconciler) SecretFromOIDC(crd *akmv1a1.OIDC) *corev1.Secret {
//	var dataMap = make(map[string][]byte)
//	dataMap["clientSecret"] = []byte(crd.Spec.ClientSecret)
//	dataMap["clientID"] = []byte(crd.Spec.ClientID)
//	oidcSecret := &corev1.Secret{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:        crd.Name,
//			Namespace:   crd.Namespace,
//			Annotations: crd.Annotations,
//		},
//		Data: dataMap,
//	}
//	ctrl.SetControllerReference(crd, oidcSecret, r.Scheme)
//	return oidcSecret
//}
//
//// BlueprintFromOIDC creates the necessary blueprint to enable OIDC for an application.
//func (r *OIDCReconciler) BlueprintFromOIDC(crd *akmv1a1.OIDC) ([]*akmv1a1.AkBlueprint, error) {
//	name := strings.ToLower(fmt.Sprintf("%v-%v-%v", crd.Namespace, crd.Kind, crd.Name))
//	name = regexp.MustCompile(`[^a-zA-Z0-9\-\_]+`).ReplaceAllString(name, "")
//	appName := fmt.Sprintf("%v-application", name)
//	provName := fmt.Sprintf("%v-provider", name)
//
//	var entries = make([]akmv1a1.BPModel, 2)
//
//	appIdentifier := make(map[string]interface{})
//	appIdentifier["slug"] = appName
//	appIdentifierBytes, err := json.Marshal(appIdentifier)
//	if err != nil {
//		return nil, err
//	}
//	// unmarshal app identifier into raw.Raw
//	aIRaw := raw.Raw{}
//	err = json.Unmarshal(appIdentifierBytes, &aIRaw)
//	if err != nil {
//		return nil, err
//	}
//
//	//TODO: instead of instantiating like so lets turn this into a propper struct
//	// to allow for consistency and re-use, especially when any changes are necessary
//	// it will notify us of where we need to change in code references a lot sooner.
//	appAttrs := make(map[string]interface{})
//	appAttrs["name"] = crd.Namespace
//	appAttrs["group"] = crd.Namespace
//	appAttrs["policy_engine_mode"] = "any"
//	// provider must point to the pk of the provider model
//	appAttrs["provider"] = fmt.Sprintf("!KeyOf %v", provName)
//	appAttrs["slug"] = crd.Namespace
//	appAttrsBytes, err := json.Marshal(appAttrs)
//	// unmarshall app attrs into raw.Raw
//	aARaw := raw.Raw{}
//	err = json.Unmarshal(appAttrsBytes, &aARaw)
//	if err != nil {
//		return nil, err
//	}
//
//	// authentik "application" model
//	entries[0] = akmv1a1.BPModel{
//		Model:       "authentik_core.application",
//		State:       "present",
//		Id:          appName,
//		Identifiers: &aIRaw,
//		Attrs:       &aARaw,
//	}
//
//	// provider meta
//	provIdentifier := make(map[string]interface{})
//	provIdentifier["slug"] = provName
//	provIdentifierBytes, err := json.Marshal(provIdentifier)
//	if err != nil {
//		return nil, err
//	}
//	// unmarshal provider identifier into raw.Raw
//	pIRaw := raw.Raw{}
//	err = json.Unmarshal(provIdentifierBytes, &pIRaw)
//	if err != nil {
//		return nil, err
//	}
//
//	// provider attribs
//	type ProviderAttribs struct {
//	}
//	provAttrs := map[string]interface{}{
//		"access_code_validity":  crd.Spec.AccessCodeValidity,
//		"access_token_validity": crd.Spec.AccessTokenValidity,
//		//accesstoken:,
//		//application:,
//		"authentication_flow": fmt.Sprintf("!KeyOf %v", crd.Spec.AuthenticationFlow),
//		//authentication_flow_id:,
//		"authorization_flow": fmt.Sprintf("!KeyOf %v", crd.Spec.AuthorizationFlow),
//		//authorization_flow_id:,
//		//authorizationcode:,
//		//backchannel_application:,
//		//backchannel_application_id:,
//		"client_id":     crd.Spec.ClientID,
//		"client_secret": crd.Spec.ClientSecret,
//		"client_type":   crd.Spec.ClientType,
//		//devicetoken:,
//		//id:,
//		"include_claims_in_id_token": true,
//		//is_backchannel:,
//		"issuer_mode": crd.Spec.IssuerMode,
//		//jwks_sources:,
//		"name": crd.Namespace,
//		//outpost:,
//		//property_mappings:,
//		//provider_ptr:,
//		//provider_ptr_id:,
//		//proxyprovider:,
//		//redirect_uris:,
//		"refresh_token_validity": crd.Spec.RefreshTokenValidity,
//		//refreshtoken:,
//		"signing_key": crd.Spec.SigningKey,
//		//signing_key_id:,
//		"sub_mode": crd.Spec.SubMode,
//	}
//	provAttrsBytes, err := json.Marshal(provAttrs)
//	if err != nil {
//		return nil, err
//	}
//	// unmarshall provider attrs into raw.Raw
//	pARaw := raw.Raw{}
//	err = json.Unmarshal(provAttrsBytes, &pARaw)
//	if err != nil {
//		return nil, err
//	}
//
//	// authentik "provider" model
//	entries[1] = akmv1a1.BPModel{
//		Model:       "authentik_providers_oauth2.oauth2provider",
//		State:       "present",
//		Id:          provName,
//		Identifiers: &pIRaw,
//		Attrs:       &pARaw,
//	}
//
//	var blueprints = make([]*akmv1a1.AkBlueprint, len(entries))
//	for ix, el := range entries {
//
//		bp := akmv1a1.BP{
//			Version: 1,
//			Metadata: akmv1a1.BPMeta{
//				Name: el.Id,
//			},
//			Entries: []akmv1a1.BPModel{el},
//		}
//
//		bpBytes, err := yaml_v3.Marshal(&bp)
//		if err != nil {
//			return nil, err
//		}
//
//		akbp := &akmv1a1.AkBlueprint{
//			// Metadata
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      el.Id,
//				Namespace: "default",
//				// TODO copy some annotations as if we copy last-applied-configuration we get:
//				// https://github.com/argoproj/argo-cd/issues/3657
//				//Annotations: crd.Annotations,
//			},
//			// Specification
//			Spec: akmv1a1.AkBlueprintSpec{
//				// Setting to near-arbitrary but unique path assuming name is properly sanitised
//				File:      fmt.Sprintf("/blueprints/operator/%v", el.Id),
//				Blueprint: string(bpBytes),
//			},
//		}
//		blueprints[ix] = akbp
//	}
//
//	// set that we are controlling this resource
//	// Annoyingly setting this causes the other reconciler to take it over
//	// when this is taken over it then gets deleted almost instantly so its kind of
//	// a nuisance. I need to either find a good way to prevent its sudden deletion or
//	// clean it up manually when the inital CRD is removed or changed to not need it
//	// any more.
//	//ctrl.SetControllerReference(crd, bp, r.Scheme)
//	return blueprints, nil
//}
//
//// IngressFromOIDC creates the necessary well-known configuration for a set of domains
//// to a given authentik server.
//func (r *OIDCReconciler) IngressFromOIDC(crd *akmv1a1.OIDC) *netv1.Ingress {
//	oidcIngress := &netv1.Ingress{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:        crd.Name,
//			Namespace:   crd.Namespace,
//			Annotations: crd.Annotations,
//		},
//	}
//	ctrl.SetControllerReference(crd, oidcIngress, r.Scheme)
//	return oidcIngress
//}
//
//// TestOIDCLiveness checks OIDC is working and is producing expected results based on the following procedure:
//func (r *OIDCReconciler) TestOIDCLiveness(ctx context.Context, url, id, secret, redirect string) error {
//
//	provider, err := oidc.NewProvider(ctx, url)
//	if err != nil {
//		return err
//	}
//	_ = oauth2.Config{
//		ClientID:     id,
//		ClientSecret: secret,
//		Endpoint:     provider.Endpoint(),
//		RedirectURL:  "http://127.0.0.1:5556/auth/google/callback",
//		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
//	}
//	return nil
//
//}

// SetupWithManager sets up the controller with the Manager.
func (r *OIDCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1alpha1.OIDC{}).
		Complete(r)
}
