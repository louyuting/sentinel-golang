package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type FlowRule struct {
	Resource          string  `json:"resource"`
	LimitOrigin       string  `json:"limitOrigin"`
	MetricType        int     `json:"metricType"`
	Count             float64 `json:"count"`
	RelationStrategy  int     `json:"relationStrategy"`
	ControlBehavior   int     `json:"controlBehavior"`
	RefResource       string  `json:"refResource"`
	WarmUpPeriodSec   uint32  `json:"warmUpPeriodSec"`
	MaxQueueingTimeMs uint32  `json:"maxQueueingTimeMs"`
}

// FlowRulesSpec defines the desired state of FlowRules
type FlowRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Rules []FlowRule `json:"rules"`
}

// FlowRulesStatus defines the observed state of FlowRules
type FlowRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlowRules is the Schema for the flowrules API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=flowrules,scope=Namespaced
type FlowRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowRulesSpec   `json:"spec,omitempty"`
	Status FlowRulesStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlowRulesList contains a list of FlowRules
type FlowRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowRules{}, &FlowRulesList{})
}
