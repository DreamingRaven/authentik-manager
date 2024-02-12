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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	akmv1alpha1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils"
	uhelm "gitlab.com/GeorgeRaven/authentik-manager/operator/utils/helm"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils/raw"
	yaml_v3 "gopkg.in/yaml.v3"
)

// Statically bundled templaes to ensure they are available in binaries

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

	// AUTHENTIK INSTANCE
	// check operator and authentik instance specified are the same namespace
	if crd.Spec.Instance.Namespace != o.OperatorNamespace {
		l.Info(fmt.Sprintf("OIDC resource reconciliation triggered but CRD specifies a different namespace to operator (operator namespace: %v, crd namespace: %v), Ignoring.", o.OperatorNamespace, crd.Spec.Instance.Namespace))
		return ctrl.Result{}, nil
	}
	aks, err := r.ListAk(o.OperatorNamespace)
	if err != nil {
		l.Error(err, "Failed to get Authentik instance. Retrying.")
	}
	if len(aks) > 1 {
		l.Info(fmt.Sprintf("OIDC resource reconciliation triggered but more than one Authentik instance found in namespace `%v`. Ignoring.", o.OperatorNamespace))
		return ctrl.Result{}, fmt.Errorf("more than one Authentik instance found in namespace `%v`", o.OperatorNamespace)
	} else if len(aks) == 0 {
		l.Info(fmt.Sprintf("OIDC resource reconciliation triggered but no Authentik instance found in namespace `%v`. Retrying.", o.OperatorNamespace))
		return ctrl.Result{}, fmt.Errorf("no Authentik instance found in namespace `%v`", o.OperatorNamespace)
	}
	ak := aks[0]

	// PROVIDERS - generate secret and blueprint for each provider
	// secret contains clientID and clientSecret
	for i := range crd.Spec.Providers {
		provider := &crd.Spec.Providers[i]
		fmt.Printf("provider: %v\n", provider.Name)
		secret, err := r.spawnAndFetchOIDCSecret(ctx, crd, provider)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("secret: %v\n", secret)
		// ensure clientID and clientSecret are back into provider
		provider.ProtocolSettings.ClientID = string(secret.Data["clientID"])
		provider.ProtocolSettings.ClientSecret = string(secret.Data["clientSecret"])
		provider_blueprint, err := r.reconcileProviderBlueprint(ak, ctx, crd, provider)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("privider blueprint: %v\n", provider_blueprint)
	}

	// APPLICATIONS - generate configmap and blueprint for each application
	// config contains urls for login, profile, logout, well-known, etc
	for i := range crd.Spec.Applications {
		application := &crd.Spec.Applications[i]
		fmt.Printf("application: %v\n", application.Name)
		configmap, err := r.reconcileConfigmap(ak, ctx, crd, application)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("configmap: %v\n", configmap)
		application_blueprint, err := r.reconcileApplicationBlueprint(ak, ctx, crd, application)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("application blueprint: %v\n", application_blueprint)
	}
	//TODO: add live testing of OIDC status by operator and locking system to prevent binding to non-functioning OIDC

	return ctrl.Result{}, nil
}

// spawnAndFetchOIDCSecret creates a secret for a client application to use to register and identify itself using the client_id and client_secret within.
func (r *OIDCReconciler) spawnAndFetchOIDCSecret(ctx context.Context, crd *akmv1a1.OIDC, provider *akmv1a1.OIDCProvider) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: provider.Secret.Name, Namespace: crd.Namespace}, secret)
	if err != nil {
		// if the secret does not exist, create it
		if errors.IsNotFound(err) {
			secret = r.SecretFromOIDCProvider(crd, provider)
			err = r.Create(ctx, secret)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	// return the secret we just created or that we found
	return secret, nil
}

// SecretFromOIDCProvider creates a secret specification containing the clientID and clientSecret with controller references
func (r *OIDCReconciler) SecretFromOIDCProvider(crd *akmv1a1.OIDC, provider *akmv1a1.OIDCProvider) *corev1.Secret {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	clientSecret := provider.ProtocolSettings.ClientSecret
	if clientSecret == "" {
		clientSecret = utils.GenerateRandomString(128, charset)
	}
	clientID := provider.ProtocolSettings.ClientID
	if clientID == "" {
		clientID = utils.GenerateRandomString(64, charset)
	}
	var dataMap = make(map[string][]byte)
	dataMap["clientSecret"] = []byte(clientSecret)
	dataMap["clientID"] = []byte(clientID)
	oidcSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        provider.Secret.Name,
			Namespace:   crd.Namespace,
			Annotations: crd.Annotations,
		},
		Data: dataMap,
	}
	ctrl.SetControllerReference(crd, oidcSecret, r.Scheme)
	return oidcSecret
}

// reconcileConfigmap creates a configmap to let the client know the relevant endpoints to use for OIDC
func (r *OIDCReconciler) reconcileConfigmap(ak *akmv1a1.Ak, ctx context.Context, crd *akmv1a1.OIDC, application *akmv1a1.OIDCApplication) (*corev1.ConfigMap, error) {
	akfqdn, err := uhelm.GetAkFQDN(ak)
	if err != nil {
		return nil, err
	}
	// generate desired configmap with all URLs based on existing AK crd
	configmap := r.ConfigmapFromOIDC(akfqdn, crd, application)
	// Always try to update and create configmap as we want to keep it in sync
	err = r.Update(ctx, configmap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, configmap)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	// apply generated configmap in this namespace
	return configmap, nil
}

func (r *OIDCReconciler) ConfigmapFromOIDC(akfqdn string, crd *akmv1a1.OIDC, application *akmv1a1.OIDCApplication) *corev1.ConfigMap {
	var slug string = application.Slug
	var dataMap = make(map[string]string)
	dataMap["wellKnownURL"] = fmt.Sprintf(
		"https://%v/application/o/%v/.well-known/openid-configuration", akfqdn, slug)
	dataMap["issuerURL"] = fmt.Sprintf(
		"https://%v/application/o/%v/", akfqdn, slug)
	dataMap["authorizationURL"] = fmt.Sprintf(
		"https://%v/application/o/authorize/", akfqdn)
	dataMap["tokenURL"] = fmt.Sprintf(
		"https://%v/application/o/token/", akfqdn)
	dataMap["userInfoURL"] = fmt.Sprintf(
		"https://%v/application/o/userinfo/", akfqdn)
	dataMap["logoutURL"] = fmt.Sprintf(
		"https://%v/application/o/%v/end-session/", akfqdn, slug)
	dataMap["jwksURL"] = fmt.Sprintf(
		"https://%v/application/o/%v/jwks/", akfqdn, slug)
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        application.ConfigMap.Name,
			Namespace:   crd.Namespace,
			Annotations: crd.Annotations,
		},
		Data: dataMap,
	}
	ctrl.SetControllerReference(crd, configmap, r.Scheme)
	return configmap
}

func (r *OIDCReconciler) reconcileApplicationBlueprint(ak *akmv1a1.Ak, ctx context.Context, crd *akmv1a1.OIDC, application *akmv1a1.OIDCApplication) (*akmv1a1.AkBlueprint, error) {
	// TODO: Once this works skip intermediary stage and go straight to blueprint
	// as this is far too complex, but it does add sanity checking
	// Really complicated way to take a map[string]string and convert it to a raw.Raw
	mapId := map[string]interface{}{
		//"id": nil,
		"slug": application.Slug,
	}
	rawId, err := yaml_v3.Marshal(mapId)
	if err != nil {
		return nil, err
	}
	id := raw.Raw{}
	err = yaml_v3.Unmarshal(rawId, &id)
	if err != nil {
		return nil, err
	}

	// Another complicated way to create the attrs map
	mapAttrs := map[string]string{
		"name":               application.Name,
		"group":              application.Group,
		"policy_engine_mode": application.PolicyEngineMode,
		"provider":           fmt.Sprintf("!Find [authentik_providers_oauth2.oauth2provider, [name, %v]]", application.Provider),
		"slug":               application.Slug,
	}
	rawAttrs, err := yaml_v3.Marshal(mapAttrs)
	if err != nil {
		return nil, err
	}
	attrs := raw.Raw{}
	err = yaml_v3.Unmarshal(rawAttrs, &attrs)
	if err != nil {
		return nil, err
	}

	bpContent := &akmv1a1.BP{
		Version: 1,
		Metadata: akmv1a1.BPMeta{
			Name: fmt.Sprintf("%v-app-%v", crd.Namespace, application.Slug),
		},
		Entries: []akmv1a1.BPModel{
			akmv1a1.BPModel{
				Model:       "authentik_core.application",
				State:       "present",
				Identifiers: &id,
				Attrs:       &attrs,
				Conditions:  []string{},
			},
		},
	}
	fmt.Printf("APPLICATION BLUEPRINT: %+v\n", bpContent)
	bpContentStr, err := yaml_v3.Marshal(bpContent)
	fmt.Printf("APPLICATION BLUEPRINT: %+v\n", bpContentStr)
	if err != nil {
		return nil, err
	}
	bp := &akmv1a1.AkBlueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-app-%v", crd.Namespace, application.Slug),
			Namespace: ak.Namespace,
		},
		Spec: akmv1a1.AkBlueprintSpec{
			StorageType: "file",
			File:        fmt.Sprintf("/blueprints/operator/%v-app-%v.yaml", crd.Namespace, application.Slug),
			Blueprint:   string(bpContentStr),
		},
	}
	ctrl.SetControllerReference(crd, bp, r.Scheme)
	fmt.Printf("APPLICATION BLUEPRINT: %+v\n", bp)

	err = r.Update(ctx, bp)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, bp)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return bp, nil
}

// reconcileProviderBlueprint ensure the provider blueprint exists and matches the desired state in the auth namespace
func (r *OIDCReconciler) reconcileProviderBlueprint(ak *akmv1a1.Ak, ctx context.Context, crd *akmv1a1.OIDC, provider *akmv1a1.OIDCProvider) (*akmv1a1.AkBlueprint, error) {
	// TODO: Once this works skip intermediary stage and go straight to blueprint
	// as this is far too complex, but it does add sanity checking
	// Really complicated way to take a map[string]string and convert it to a raw.Raw

	//mapId := map[string]interface{}{
	//	"id": nil,
	//}
	//rawId, err := yaml_v3.Marshal(mapId)
	//if err != nil {
	//	return nil, err
	//}
	//id := raw.Raw{}
	//err = yaml_v3.Unmarshal(rawId, &id)
	//if err != nil {
	//	return nil, err
	//}

	bpPlainContent := map[string]interface{}{
		"version": 1,
		"metadata": map[string]interface{}{
			"name": fmt.Sprintf("%v-provider-%v", crd.Namespace, provider.Name),
		},
		"entries": []map[string]interface{}{
			{
				"model": "authentik_providers_oauth2.oauth2provider",
				"state": "created",
				"identifiers": map[string]interface{}{
					"id": nil,
				},
				"attrs": map[string]interface{}{
					"name":                       provider.Name,
					"access_code_validity":       provider.ProtocolSettings.AccessCodeValidity,
					"access_token_validity":      provider.ProtocolSettings.AccessTokenValidity,
					"refresh_token_validity":     provider.ProtocolSettings.RefreshTokenValidity,
					"authentication_flow":        fmt.Sprintf("!Find [authentik_flows.flow, [slug, %v]]", provider.AuthenticationFlow),
					"authorization_flow":         fmt.Sprintf("!Find [authentik_flows.flow, [slug, %v]]", provider.AuthorizationFlow),
					"signing_key":                fmt.Sprintf("!Find [authentik_crypto.certificatekeypair, [name, %v]]", provider.ProtocolSettings.SigningKey),
					"client_id":                  provider.ProtocolSettings.ClientID,
					"client_secret":              provider.ProtocolSettings.ClientSecret,
					"client_type":                provider.ProtocolSettings.ClientType,
					"include_claims_in_id_token": bool(provider.ProtocolSettings.IncludeClaimsInIDToken),
					"issuer_mode":                provider.ProtocolSettings.IssuerMode,
					"redirect_uris":              provider.ProtocolSettings.RedirectURIs,
					"subject_claims":             provider.ProtocolSettings.SubjectMode,
				},
				"conditions": []string{},
			},
		},
	}

	bpPlainContentStr, err := yaml_v3.Marshal(bpPlainContent)
	if err != nil {
		return nil, err
	}

	//bpPlain := map[string]interface{}{
	//b	"metadata": metav1.ObjectMeta{
	//b		Name:      fmt.Sprintf("%v-provider-%v", crd.Namespace, provider.Name),
	//b		Namespace: ak.Namespace,
	//b	},
	//b	"spec": map[string]interface{}{
	//b		"storageType": "file",
	//b		"file":        fmt.Sprintf("/blueprints/operator/%v-provider-%v.yaml", crd.Namespace, provider.Name),
	//b		"blueprint":   string(bpPlainContentStr),
	//b	},
	//b}

	//b// Another complicated way to create the attrs map
	//bmapAttrs := map[string]interface{}{
	//b	"name":                       provider.Name,
	//b	"access_code_validity":       provider.ProtocolSettings.AccessCodeValidity,
	//b	"access_token_validity":      provider.ProtocolSettings.AccessTokenValidity,
	//b	"refresh_token_validity":     provider.ProtocolSettings.RefreshTokenValidity,
	//b	"authentication_flow":        fmt.Sprintf("!Find [authentik_flows.flow, [slug, %v]]", provider.AuthenticationFlow),
	//b	"authorization_flow":         fmt.Sprintf("!Find [authentik_flows.flow, [slug, %v]]", provider.AuthorizationFlow),
	//b	"signing_key":                fmt.Sprintf("!Find [authentik_crypto.certificatekeypair, [name, %v]]", provider.ProtocolSettings.SigningKey),
	//b	"client_id":                  provider.ProtocolSettings.ClientID,
	//b	"client_secret":              provider.ProtocolSettings.ClientSecret,
	//b	"client_type":                provider.ProtocolSettings.ClientType,
	//b	"include_claims_in_id_token": bool(provider.ProtocolSettings.IncludeClaimsInIDToken),
	//b	"issuer_mode":                provider.ProtocolSettings.IssuerMode,
	//b	"redirect_uris":              provider.ProtocolSettings.RedirectURIs,
	//b	"subject_claims":             provider.ProtocolSettings.SubjectMode,
	//b}
	//bfmt.Printf("PROVIDER BLUEPRINT ATTRS STRUCT: %+v\n", mapAttrs)
	//brawAttrs, err := yaml_v3.Marshal(mapAttrs)
	//bfmt.Printf("PROVIDER BLUEPRINT ATTRS SERIAL: %+v\n", string(rawAttrs))
	//bif err != nil {
	//b	return nil, err
	//b}
	//battrs := raw.Raw{}
	//berr = yaml_v3.Unmarshal(rawAttrs, &attrs)
	//bif err != nil {
	//b	return nil, err
	//b}

	//bbpContent := &akmv1a1.BP{
	//b	Version: 1,
	//b	Metadata: akmv1a1.BPMeta{
	//b		Name: fmt.Sprintf("%v-provider-%v", crd.Namespace, provider.Name),
	//b	},
	//b	Entries: []akmv1a1.BPModel{
	//b		akmv1a1.BPModel{
	//b			Model:       "authentik_providers_oauth2.oauth2provider",
	//b			State:       "present",
	//b			Identifiers: &id,
	//b			Attrs:       &attrs,
	//b			Conditions:  []string{},
	//b		},
	//b	},
	//b}
	//bfmt.Printf("PROVIDER BLUEPRINT CONTENT: %+v\n", bpContent)
	//bbpContentStr, err := yaml_v3.Marshal(bpContent)
	//bfmt.Printf("PROVIDER BLUEPRINT CONTENT STRING: %+v\n", string(bpContentStr))
	//bif err != nil {
	//b	return nil, err
	//b}
	bp := &akmv1a1.AkBlueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-provider-%v", crd.Namespace, provider.Name),
			Namespace: ak.Namespace,
		},
		Spec: akmv1a1.AkBlueprintSpec{
			StorageType: "file",
			File:        fmt.Sprintf("/blueprints/operator/%v-provider-%v.yaml", crd.Namespace, provider.Name),
			Blueprint:   string(bpPlainContentStr),
		},
	}
	ctrl.SetControllerReference(crd, bp, r.Scheme)
	fmt.Printf("PROVIDER BLUEPRINT: %+v\n", bp)

	err = r.Update(ctx, bp)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, bp)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return bp, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OIDCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1alpha1.OIDC{}).
		Complete(r)
}
