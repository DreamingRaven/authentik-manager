package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AkServerSpec defines the desired state of AkServer
type AkServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AkServer. Edit akserver_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AkServerStatus defines the observed state of AkServer
type AkServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AkServer is the Schema for the akservers API
type AkServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AkServerSpec   `json:"spec,omitempty"`
	Status AkServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AkServerList contains a list of AkServer
type AkServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AkServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AkServer{}, &AkServerList{})
}
