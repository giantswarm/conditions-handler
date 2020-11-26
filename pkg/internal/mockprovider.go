package internal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

// MockProviderCluster is the Schema for the mockazureclusters API
type MockProviderCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MockProviderClusterSpec   `json:"spec,omitempty"`
	Status MockProviderClusterStatus `json:"status,omitempty"`
}

// MockProviderClusterSpec defines the desired state of MockProviderCluster
type MockProviderClusterSpec struct{}

// MockProviderClusterStatus defines the observed state of MockProviderCluster
type MockProviderClusterStatus struct {
	Conditions capi.Conditions `json:"conditions,omitempty"`
}

// GetConditions returns the list of conditions for an MockProviderCluster API object.
func (c *MockProviderCluster) GetConditions() capi.Conditions {
	return c.Status.Conditions
}

// SetConditions will set the given conditions on an MockProviderCluster object
func (c *MockProviderCluster) SetConditions(conditions capi.Conditions) {
	c.Status.Conditions = conditions
}

func (c *MockProviderCluster) GetObjectKind() schema.ObjectKind { return c }

func (c *MockProviderCluster) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(c.TypeMeta.APIVersion, c.TypeMeta.Kind)
}

func (c *MockProviderCluster) DeepCopyObject() runtime.Object {
	panic("MockProviderCluster does not support DeepCopy")
}
