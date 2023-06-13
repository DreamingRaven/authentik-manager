/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OIDCSpec defines the desired state of OIDC
type OIDCSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Domains is a list of domain names the OIDC controller should capture the /well-known paths from.
  // Each domain will be enforced to be unique between all namespaces.
	Domains []string `json:"domains,omitempty"`
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
