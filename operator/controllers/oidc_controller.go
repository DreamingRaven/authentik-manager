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
	"fmt"
	"regexp"
	"strings"

	"github.com/alexflint/go-arg"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OIDC object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *OIDCReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := klog.FromContext(ctx)

	// Parsing options to make them available TODO: pass them in rather than read continuously
	o := utils.Opts{}
	arg.MustParse(&o)

	actionConfig, err := r.GetActionConfig(req.NamespacedName.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	// GET CRD
	crd := &akmv1a1.OIDC{}
	err = r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("OIDC resource reconciliation triggered but disappeared. Uninstalling OIDC integration.")
			_, err := r.UninstallChart(req.NamespacedName, actionConfig)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get OIDC resource. Likely fetch error. Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found OIDC resource `%v` in `%v` for domains %v.", crd.Name, crd.Namespace, crd.Spec.Domains))

	// GENERATE BLUEPRINT
	bp := r.BlueprintFromOIDC(crd)
	fmt.Printf("blueprint %v", utils.PrettyPrint(bp))

	// GENERATE SECRET

	// GENERATE INGRESS WELL-KNOWN

	return ctrl.Result{}, nil
}

func (r *OIDCReconciler) BlueprintFromOIDC(crd *akmv1a1.OIDC) *akmv1a1.AkBlueprint {
	name := strings.ToLower(fmt.Sprintf("%v-%v-%v", crd.Namespace, crd.Kind, crd.Name))
	name = regexp.MustCompile(`[^a-zA-Z0-9\-\_]+`).ReplaceAllString(name, "")
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
			},
		},
	}
	// set that we are controlling this resource
	ctrl.SetControllerReference(crd, bp, r.Scheme)
	return bp
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