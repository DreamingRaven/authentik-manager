/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AkApplicationUI defines how an associated application appears in authentiks UI
type AkApplicationUI struct {

	// LaunchURL (optional) what url will be launched from UI, defaults to providers url if empty
	LaunchURL string `json:"launchURL,omitempty"`

	// Icon (optional) url to an image to use in authentiks dashboard
	Icon string `json:"icon,omitempty"`

	// Publisher (optional) name to identify who publishes this application
	Publisher string `json:"publisher,omitempty"`

	// Description (optional) that describes the application itself and what its purpose is
	Description string `json:"decription,omitempty"`
}

// AkApplicationSpec defines the desired state of AkApplication
type AkApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Slug (required) is the internal name to be used in authentik urls
	Slug string `json:"slug,omitempty"`

	// Group (optional) is how authentik knows which applications are to be grouped together
	Group string `json:"group,omitempty"`

	// Provider (required) is a reference to the AkProvider CRD that this application should be related to (not the internal authentik provider)
	Provider string `json:"provider,omitempty"`

	// PolicyEngineMode (required) is whether any or all policies must match to grant access
	PolicyEngineMode string `json:"policyEngineMode,omitempty"`

	// ApplicationUI (optional) is additional settings that define how this application appears in authentiks ui
	ApplicationUI *AkApplicationUI `json:"applicationUI,omitempty"`
}

// AkApplicationStatus defines the observed state of AkApplication
type AkApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AkApplication is the Schema for the akapplications API
type AkApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AkApplicationSpec   `json:"spec,omitempty"`
	Status AkApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AkApplicationList contains a list of AkApplication
type AkApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AkApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AkApplication{}, &AkApplicationList{})
}
