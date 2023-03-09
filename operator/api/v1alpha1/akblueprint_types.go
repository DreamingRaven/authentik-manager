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

// AkBlueprintSpec defines the desired state of AkBlueprint
type AkBlueprintSpec struct {

	// Blueprint is a container for a complete single authentik blueprint yaml spec
	Blueprint BP `json:"blueprint,omitempty"`
}

// BP is a whole blueprint struct containing the full structure of an authentik blueprint
// https://goauthentik.io/developer-docs/blueprints/v1/structure#structure
type BP struct {
	// Version is the version of this blueprint
	Version int `json:"version"`

	// Metadata block specifying labels and names of the blueprint
	Metadata BPMeta `json:"metadata"`

	// Context (optional) authentik default context (whatever that means)
	Context json.RawMessage `json:"context,omitempty"`

	// Entries lists models we want to use via this blueprint
	Entries []BPModel `json:"entries"`
}

// BPMeta is the metadata of an authentik blueprint as authentik likes
type BPMeta struct {

	// Labels (optional) ke-value store for special labels
	// https://goauthentik.io/developer-docs/blueprints/v1/structure#special-labels
	Labels json.RawMessage `json:"labels,omitempty"`

	// Name of the authentik blueprint for authentik to register
	Name string `json:"name"`
}

// BPModel is a rough outline of the structure of models authentik likes in its blueprints
type BPModel struct {

	// Model "app.model" notation of which model from authentik to call
	Model string `json:"model"`

	//+kubebuilder:validation:Enum="present";"create";"absent"
	//+kubebuilder:default:=present
	// State (optional) desired state of this model when loaded from "present", "create", "absent"
	// present: (default) keeps the object in sync with its definition in this blueprint
	// create: only creates the initial object with its values here
	// absent: deletes the object
	State string `json:"state,omitempty"`

	// Conditions (optional) a list of conditions which if all match the model will be activated. If not the model will be inactive
	Conditions []string `json:"conditions,omitempty"`

	// Identifiers key-value identifiers to allow filtering of this stage, and identifying it
	Identifiers json.RawMessage `json:"identifiers"`

	// Id (optional) is similar to identifiers except is optional and is just an ID to reference this model using !KeyOf syntax in authentik
	Id string `json:"id,omitempty"`

	// Attrs is a list of settings / options / overrides of the defaults of this model
	Attrs json.RawMessage `json:"attrs,omitempty"`
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
