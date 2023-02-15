package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//+kubebuilder:validation:Optional

// AkSpec defines the desired state of Ak
type AkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Naming is a struct that defines how the authentik resources should be named
	Naming AkNaming `json:"naming,omitempty"`

	// Domain is a struct that defines the domain authentik should operate within
	Domain AkDomain `json:"domain,omitempty"`

	// Secret is the definition of which secret to use and how to manage it
	Secret AkSecret `json:"secret,omitempty"`

	// Smtp controls how and where to connect to an SMTP server for authentik
	Smtp AkSmtp `json:"smtp,omitempty"`

	// Ingress contains settings for ingress route control
	Ingress AkIngress `json:"ingress,omitempty"`

	//+kubebuilder:validation:Enum="tyranny";"republic"
	//+kubebuilder:default:=tyranny
	// Mode specifies if the controller is going to make things happen one way or another or if it is going to try to cooperate with existing installations
	Mode string `json:"mode,omitempty"`
}

type AkNaming struct {

	//+kubebuilder:default:=authentik
	// Base is the prefix name to use e.g authentik-server or something-server
	Base string `json:"base,omitempty"`

	//+kubebuilder:default:=server
	// Server is the suffix name to append to the server specific deployment
	Server string `json:"server,omitempty"`

	//+kubebuilder:default:=worker
	// Worker is the suffix name to append to the celery specific deployment
	Worker string `json:"worker,omitempty"`
}

type AkDomain struct {

	//+kubebuilder:default:=example.org
	// Base is the root domain name e.g example.org
	Base string `json:"base,omitempty"`

	//+kubebuilder:default:=auth.example.org
	// Full is the full domain name including subdomain which authentik should listen for
	Full string `json:"full,omitempty"`
}

type AkSecret struct {

	//+kubebuilder:default:=auth
	// Name is the name of the secret to look for in the same namespace or to generate
	Name string `json:"name,omitempty"`

	//+kubebuilder:default:=true
	// Generate is a flag to tell the operator if it should manage its own secret
	Generate bool `json:"generate,omitempty"`
}

type AkSmtp struct {

	// Username the name which should be use to authenticate to the SMTP server
	Username string `json:"username,omitempty"`

	// Port on SMTP host to connect to
	Port int `json:"port,omitempty"`

	// Host to connect to
	Host string `json:"host,omitempty"`

	// From who the emails sent via this SMTP email server should be from
	From string `json:"from,omitempty"`

	// key the key to look up in the .spec.secret.name secret
	Key string `json:"key,omitempty"`
}

type AkIngress struct {

	//+kubebuilder:default:=true
	// Enabled dictates whether this controller should manage authentiks ingress resources or not
	Enabled bool `json:"enabled,omitempty"`
}

// AkStatus defines the observed state of Ak
type AkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions"`
}

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
