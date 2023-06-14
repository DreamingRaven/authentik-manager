/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

/*

OpenID Connect (OIDC) is an authentication protocol built on top of the OAuth 2.0 framework. It provides a standardized way for users to authenticate and obtain identity information from an identity provider (IdP). Here's a simplified overview of how OIDC works:

1. *Client Registration*: The client application (relying party) registers itself with the OIDC provider (authorization server). During registration, the client obtains a client ID and client secret, which are used to identify and authenticate the client.

2. *Authentication Request*: When a user wants to authenticate with the client application, they are redirected to the OIDC provider's authentication endpoint. The client initiates the authentication request by sending the user to this endpoint, along with the client ID, requested scopes, and a redirect URL.

3. *User Authentication*: The user is presented with a login page provided by the OIDC provider. The authentication process can involve various mechanisms such as username/password, multifactor authentication, or social logins. The user submits their credentials to the OIDC provider for verification.

4. *Authorization Grant*: After successful authentication, the OIDC provider asks the user to authorize the client application to access their protected resources. This step ensures that the user explicitly grants consent to the client.

5. *Issuance of Authorization Code*: If the user grants authorization, the OIDC provider issues an authorization code and redirects the user back to the client application's redirect URL. The authorization code serves as a temporary, one-time-use credential.

6. *Token Request*: The client application receives the authorization code and sends a token request to the OIDC provider's token endpoint. Along with the code, the client also provides its client ID and client secret. This request is typically made using the OAuth 2.0 "authorization code" grant type.

7. *Token Response*: The OIDC provider verifies the client credentials and the authorization code. If valid, the provider responds with an access token, an ID token, and optionally a refresh token. The access token is used to authenticate subsequent API requests, while the ID token contains identity information about the authenticated user.

8. *User Information*: If needed, the client application can use the access token to request additional user information from the OIDC provider's user info endpoint. This endpoint provides user attributes, such as name, email, or profile picture.

9. *Token Validation*: The client application validates the received tokens' integrity, authenticity, and expiration time using cryptographic measures. It verifies the digital signature of the tokens using the OIDC provider's public key.

            +------------------+
            |    Client App    |
            +--------+---------+
                     |
            1. Initiate Authentication
                     |
            +--------v---------+
            | OIDC Provider    |
            | (Authorization   |
            |    Server)       |
            +--------+---------+
                     |
            2. Redirect User to
               Authentication Endpoint
                     |
            +--------v---------+
            | User's Web Browser|
            +--------+---------+
                     |
            3. User Authenticates
                     |
            +--------v---------+
            | OIDC Provider    |
            | (Authorization   |
            |    Server)       |
            +--------+---------+
                     |
            4. Authorization Request
                     |
            +--------v---------+
            | User's Web Browser|
            +--------+---------+
                     |
            5. Grant Authorization
                     |
            +--------v---------+
            | OIDC Provider    |
            | (Authorization   |
            |    Server)       |
            +--------+---------+
                     |
            6. Issue Authorization Code
                     |
            +--------v---------+
            |    Client App    |
            +--------+---------+
                     |
            7. Token Request
                     |
            +--------v---------+
            | OIDC Provider    |
            | (Token Endpoint) |
            +--------+---------+
                     |
            8. Token Response
                     |
            +--------v---------+
            |    Client App    |
            +--------+---------+
                     |
            9. Access Resources
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/alexflint/go-arg"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/coreos/go-oidc/v3/oidc"
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
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("OIDC resource reconciliation triggered but disappeared. Uninstalling OIDC integration.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get OIDC resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found OIDC resource `%v` in `%v` for domains %v.", crd.Name, crd.Namespace, crd.Spec.Domains))

	// FETCH SECRET OR CREATE BUT NOT UPDATE
	oidcSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      crd.Name,
		Namespace: crd.Namespace,
	}, oidcSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create secret as not found
			l.Info(fmt.Sprintf("Secret not found generating `%v` in `%v`.", crd.Name, crd.Namespace))
			charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			if crd.Spec.ClientID == "" {
				crd.Spec.ClientID = utils.GenerateRandomString(40, charset)
			}
			if crd.Spec.ClientSecret == "" {
				crd.Spec.ClientSecret = utils.GenerateRandomString(128, charset)
			}
			oidcDesiredSecret := r.SecretFromOIDC(crd)
			err = r.Create(ctx, oidcDesiredSecret)
			if err != nil {
				return ctrl.Result{}, err
			}
			oidcSecret = oidcDesiredSecret
		} else {
			// Some error other than "not found" when trying to fetch secret
			return ctrl.Result{}, err
		}
	} else {
		l.Info(fmt.Sprintf("Secret update ignored to prevent downtime for `%v` in `%v`.", oidcSecret.Name, oidcSecret.Namespace))
		crd.Spec.ClientID = string(oidcSecret.Data["clientID"])
		crd.Spec.ClientSecret = string(oidcSecret.Data["clientSecret"])
	}

	// GENERATE OR UPDATE BLUEPRINT
	bp, err := r.BlueprintFromOIDC(crd)
	if err != nil {
		return ctrl.Result{}, err
	}
	bp.Namespace = o.OperatorNamespace
	l.Info(fmt.Sprintf("Updating blueprint `%v` in `%v`", bp.Name, bp.Namespace))
	err = r.Update(ctx, bp)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info(fmt.Sprintf("Blueprint not found creating `%v` in `%v`", bp.Name, bp.Namespace))
			m, err := utils.ManifestString(bp)
			if err != nil {
				return ctrl.Result{}, err
			}
			fmt.Printf("bp: %v", m)
			err = r.Create(ctx, bp)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			l.Error(err, "Failed to update blueprint. Retrying.")
			return ctrl.Result{}, err
		}
	}

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

	//TODO: add live testing of OIDC by operator and locking system to prevent binding to non-functioning OIDC

	return ctrl.Result{}, nil
}

// SecretFromOIDC creates a secret for a client application to use to register and identify itself.
func (r *OIDCReconciler) SecretFromOIDC(crd *akmv1a1.OIDC) *corev1.Secret {
	var dataMap = make(map[string][]byte)
	dataMap["clientSecret"] = []byte(crd.Spec.ClientSecret)
	dataMap["clientID"] = []byte(crd.Spec.ClientID)
	oidcSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        crd.Name,
			Namespace:   crd.Namespace,
			Annotations: crd.Annotations,
		},
		Data: dataMap,
	}
	ctrl.SetControllerReference(crd, oidcSecret, r.Scheme)
	return oidcSecret
}

// BlueprintFromOIDC creates the necessary blueprint to enable OIDC for an application.
func (r *OIDCReconciler) BlueprintFromOIDC(crd *akmv1a1.OIDC) (*akmv1a1.AkBlueprint, error) {
	name := strings.ToLower(fmt.Sprintf("%v-%v-%v", crd.Namespace, crd.Kind, crd.Name))
	name = regexp.MustCompile(`[^a-zA-Z0-9\-\_]+`).ReplaceAllString(name, "")

	type slug struct {
		Slug string `json:"state,omitempty"`
	}
	appSlug := slug{
		Slug: fmt.Sprintf("%v-application", name),
	}
	//provSlug := slug{
	//	Slug: fmt.Sprintf("%v-provider", name),
	//}
	appMar, err := json.Marshal(appSlug)
	if err != nil {
		return nil, err
	}
	//provMar, err := json.Marshal(provSlug)
	//if err != nil {
	//	return nil, err
	//}

	var entries = make([]akmv1a1.BPModel, 1)
	// authentik "application" model
	entries[0] = akmv1a1.BPModel{
		Model:       "authentik_core.application",
		State:       "present",
		Id:          name,
		Identifiers: json.RawMessage(appMar),
	}
	// authentik "provider" model
	//entries[1] = akmv1a1.BPModel{}

	bp := &akmv1a1.AkBlueprint{
		// Metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: crd.Annotations,
		},
		// Specification
		Spec: akmv1a1.AkBlueprintSpec{
			// Setting to near-arbitrary but unique path assuming name is properly sanitised
			File: fmt.Sprintf("/blueprints/operator/%v", name),
			Blueprint: akmv1a1.BP{
				Version: 1,
				Metadata: akmv1a1.BPMeta{
					Name: name,
				},
				Entries: entries,
			},
		},
	}
	// set that we are controlling this resource
	ctrl.SetControllerReference(crd, bp, r.Scheme)
	return bp, nil
}

// IngressFromOIDC creates the necessary well-known configuration for a set of domains
// to a given authentik server.
func (r *OIDCReconciler) IngressFromOIDC(crd *akmv1a1.OIDC) *netv1.Ingress {
	oidcIngress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        crd.Name,
			Namespace:   crd.Namespace,
			Annotations: crd.Annotations,
		},
	}
	ctrl.SetControllerReference(crd, oidcIngress, r.Scheme)
	return oidcIngress
}

// TestOIDCLiveness checks OIDC is working and is producing expected results based on the following procedure:
func (r *OIDCReconciler) TestOIDCLiveness(ctx context.Context, url, id, secret, redirect string) error {

	provider, err := oidc.NewProvider(ctx, url)
	if err != nil {
		return err
	}
	_ = oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://127.0.0.1:5556/auth/google/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *OIDCReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1alpha1.OIDC{}).
		Complete(r)
}
