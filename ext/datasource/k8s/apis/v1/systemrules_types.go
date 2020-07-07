package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SystemRule struct {
	MetricType   int     `json:"metricType"`
	TriggerCount float64 `json:"triggerCount"`
	Strategy     int     `json:"strategy"`
}

// SystemRulesSpec defines the desired state of SystemRules
type SystemRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Rules []SystemRule `json:"rules"`
}

// SystemRulesStatus defines the observed state of SystemRules
type SystemRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SystemRules is the Schema for the systemrules API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=systemrules,scope=Namespaced
type SystemRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SystemRulesSpec   `json:"spec,omitempty"`
	Status SystemRulesStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SystemRulesList contains a list of SystemRules
type SystemRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SystemRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SystemRules{}, &SystemRulesList{})
}
