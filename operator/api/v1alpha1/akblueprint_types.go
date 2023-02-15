package v1alpha1

// appsv1 "k8s.io/api/apps/v1"
// batchv1 "k8s.io/api/batch/v1"
import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Required

// AkBlueprintSpec defines the desired state of AkBlueprint
type AkBlueprintSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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
