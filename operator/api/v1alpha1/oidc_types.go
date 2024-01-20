/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OIDCSpec defines abstract and safe interfaces to provision an authentik OIDC authentication stack
// this is meant to be deployed with applications and secrets so that OIDC can be provisioned for them.
type OIDCSpec struct {
	// Provider which defines how and where OIDC takes place
	Providers []OIDCProvider `json:"providers,omitempty"`
	// Applications define what the provider authenticates for
	Applications []OIDCApplication `json:"applications,omitempty"`
}

type OIDCApplication struct {
	// Name is the name of the application to display
	Name string `json:"name,omitempty"`
	// Slug is the unique name of the application used internally
	Slug string `json:"slug,omitempty"`
	//+kubebuilder:validation:Optional
	// Group is a string that is used to group applications with the idential group
	Group string `json:"group,omitempty"`
	// Provider is the primary applications provider to use
	Provider string `json:"provider,omitempty"`
	//+kubebuilder:validation:Optional
	// BackChannelProviders is a list of providers that should be used to augment the main providers functionality
	BackChannelProviders []string `json:"backChannelProviders,omitempty"`

	//+kubebuilder:validation:Enum="any";"all"
	// PolicyEngineMode determines if all or any policy engine should match to grant access
	PolicyEngineMode string `json:"policyEngineMode,omitempty"`

	//+kubebuilder:validation:Required
	// OIDCApplicationUISettings defines the behaviour of the application displayed or clicked
	OIDCApplicationUISettings OIDCApplicationUISettings `json:"oidcApplicationUISettings,omitempty"`
}

type OIDCApplicationUISettings struct {
	//+kubebuilder:validation:Optional
	// Try to detect URL based on provider or set explicitly here
	LaunchURL string `json:"launchURL,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=false
	// When user clicks on "launch url" open a new browser tab or window for it
	OpenInNewTab bool `json:"openInNewTab,omitempty"`
	//+kubebuilder:validation:Optional
	// Icon is the full URL, a relative path, or 'fa://fa-test' to use FontAwesome to display for the applications badge
	Icon string `json:"icon,omitempty"`
	//+kubebuilder:validation:Optional
	// Publisher is the name of the publisher to display to user
	Publisher string `json:"publisher,omitempty"`
	//+kubebuilder:validation:Optional
	// Description is the description of the application to display to user
	Description string `json:"description,omitempty"`
}

type OIDCProvider struct {
	//+kubebuilder:validation:Required
	// Name is the name of the provider
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	// Secret is an object reference in this namespace to generate or lookup the clientID and clientSecret
	Secret corev1.LocalObjectReference `json:"secret,omitempty"`

	//+kubebuilder:validation:Optional
	// AuthenticationFlow is the name of the authentication flow to authenticate users with
	AuthenticationFlow string `json:"authenticationFlow,omitempty"`
	//+kubebuilder:validation:Required
	// AuthorizationFlow is the name of the authorization flow to authorize this provider
	AuthorizationFlow string `json:"authorizationFlow"`
	//+kubebuilder:validation:Required
	// ProtocolSettings is the settings for the OIDC protocol to use
	ProtocolSettings OIDCProviderProtocolSettings `json:"protocolSettings"`
	//+kubebuilder:validation:Required
	// MachineToMachineSettings is the settings for machine to machine OIDC
	MachineToMachineSettings OIDCProviderMachineToMachineSettings `json:"machineToMachineSettings"`
}

type OIDCProviderProtocolSettings struct {
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Enum="confidential";"public"
	// ClientType is the type of client to use confidential or public
	ClientType string `json:"clientType"`
	//+kubebuilder:validation:Optional
	// ClientID (optional) identifies the application to the OIDC server
	// If a secret is provided we will automatically use that instead of this
	ClientID string `json:"clientID,omitempty"`
	//+kubebuilder:validation:Optional
	// ClientSecret (optional) identifies the application to the OIDC server
	// If a secret is provided we will automatically use that instead of this
	ClientSecret string `json:"clientSecret,omitempty"`
	//+kubebuilder:validation:Optional
	// Specifies valid redirect URIs to accept when returning user back from authentication / authorization flow. This also sets the origins for explicit flows.
	RedirectURIs []string `json:"redirectURIs,omitempty"`
	//+kubebuilder:validation:Optional
	// The name of the signing key to use for signing authenticated users tokens
	SigningKey string `json:"signingKey,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="minutes=1"
	// The length of time that an access code is valid in format hours=1;minutes=2;seconds=3 can be used with weeks,days,hours,minutes,seconds,milliseconds
	AccessCodeValidity string `json:"accessCodeValidity,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="minutes=5"
	// The length of time that an access token is valid in format hours=1;minutes=2;seconds=3 can be used with weeks,days,hours,minutes,seconds,milliseconds
	AccessTokenValidity string `json:"accessTokenValidity,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="days=30"
	// The length of time that a refresh token is valid in format hours=1;minutes=2;seconds=3 can be used with weeks,days,hours,minutes,seconds,milliseconds
	RefreshTokenValidity string `json:"refreshTokenValidity,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default={"email","openid","profile"}
	Scopes []string `json:"scopes,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="hashedId"
	//+kubebuilder:validation:Enum="hashedId";"id";"uuid";"username";"email";"upn"
	// SubjectMode how you know the same user between applications of this provider
	SubjectMode string `json:"subjectMode,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=true
	// IncludeClaimsInIDToken whether to include claims in token for apps that arent checking userinfo endpoint
	IncludeClaimsInIDToken bool `json:"includeClaimsInIDToken,omitempty"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default="different"
	//+kubebuilder:validation:Enum="same";"different"
	// IssuerMode whether to use the same or different issuer for each application slug
	IssuerMode string `json:"issuerMode,omitempty"`
}

type OIDCProviderMachineToMachineSettings struct {
	//+kubebuilder:validation:Optional
	TrustedOIDCSources []string `json:"trustedOIDCSources,omitempty"`
}

// ConfigmapSettings defines various information on the configmap to use or generate for well-known OIDC configuration or public information
type ConfigmapSettings struct {
	//+kubebuilder:validation:Optional
	Name string `json:"name"`
}

// SecretSettings defines various information on the secret to use or generate with clientID and clientSecret for confidential clients
type SecretSettings struct {
	//+kubebuilder:validation:Optional
	Name string `json:"name"`
}

// OIDCStatus defines the observed state of OIDC
type OIDCStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OIDC is the Schema for the oidcs API
type OIDC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OIDCSpec   `json:"spec,omitempty"`
	Status OIDCStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OIDCList contains a list of OIDC
type OIDCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OIDC `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OIDC{}, &OIDCList{})
}
