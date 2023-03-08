/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:validation:Optional

// AkSpec defines the desired state of Ak
type AkSpec struct {

	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless

	// Values is the helm chart values map to override chart defaults. This is often further adapted by the controller
	// to add additional resources like declarative blueprints into the deployments. Values is a loose, and unstructured
	// datatype. It will not complain if the values do not override anything, or do anything at all.
	Values json.RawMessage `json:"values,omitempty"`
	// Values map[string]interface{} `json:"values,omitempty"`
	// Values runtime.RawExtension `json:"values,omitempty"`

	// Blueprints is a field that specifies what blueprints should be loaded into the chart.
	Blueprints []string `json:"blueprints,omitempty"`
}

// AkStatus defines the observed state of Ak
type AkStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ak is the Schema for the aks API
type Ak struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AkSpec   `json:"spec,omitempty"`
	Status AkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AkList contains a list of Ak
type AkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ak `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ak{}, &AkList{})
}
