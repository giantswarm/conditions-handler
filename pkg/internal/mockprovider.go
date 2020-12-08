package internal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	group   = "mock.giantswarm.io"
	version = "v1alpha1"
)

// knownTypes is the full list of objects to register with the scheme. It
// should contain all zero values of custom objects and custom object lists
// in the group version.
var knownTypes = []runtime.Object{
	&MockProviderCluster{},
	&MockProviderClusterList{},
}

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   group,
	Version: version,
}

var (
	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddMockToScheme is used by the generated client.
	AddMockToScheme = schemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, knownTypes...)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// MockProviderCluster is the Schema for the mockproviderclusters API
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
	if c == nil {
		return nil
	}
	out := new(MockProviderCluster)
	c.DeepCopyInto(out)
	return out
}

func (c *MockProviderCluster) DeepCopyInto(out *MockProviderCluster) {
	*out = *c
}

type MockProviderClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []MockProviderCluster `json:"items"`
}

func (c *MockProviderClusterList) GetObjectKind() schema.ObjectKind { return c }

func (c *MockProviderClusterList) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(c.TypeMeta.APIVersion, c.TypeMeta.Kind)
}

func (c *MockProviderClusterList) DeepCopyObject() runtime.Object {
	panic("MockProviderCluster does not support DeepCopy")
}
