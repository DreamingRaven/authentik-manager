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

// AkBlueprintSpec defines the desired state of AkBlueprint
type AkBlueprintSpec struct {

	// File is the name of the file for the blueprint, it can also be a relative path
	File string `json:"file,omitempty"`

	// Blueprint is the blueprint file contents must be an absolute path
	Blueprint string `json:"blueprint,omitempty"`
}

// AkBlueprintStatus defines the observed state of AkBlueprint
type AkBlueprintStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AkBlueprint is the Schema for the akblueprints API
type AkBlueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AkBlueprintSpec   `json:"spec,omitempty"`
	Status AkBlueprintStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AkBlueprintList contains a list of AkBlueprint
type AkBlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AkBlueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AkBlueprint{}, &AkBlueprintList{})
}
