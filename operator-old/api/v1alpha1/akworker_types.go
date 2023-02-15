package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AkWorkerSpec defines the desired state of AkWorker
type AkWorkerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AkWorker. Edit akworker_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AkWorkerStatus defines the observed state of AkWorker
type AkWorkerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AkWorker is the Schema for the akworkers API
type AkWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AkWorkerSpec   `json:"spec,omitempty"`
	Status AkWorkerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AkWorkerList contains a list of AkWorker
type AkWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AkWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AkWorker{}, &AkWorkerList{})
}
