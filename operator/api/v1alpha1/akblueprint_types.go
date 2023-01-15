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

// appsv1 "k8s.io/api/apps/v1"
// batchv1 "k8s.io/api/batch/v1"
import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AkBlueprintSpec defines the desired state of AkBlueprint
type AkBlueprintSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ConfigMap is the name of the cm that holds the blueprints that should be mounted
	ConigMap string `json:"configMap,omitempty"`

	// Blueprint is the filepath the blueprint should be loaded into
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
