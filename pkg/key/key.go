package key

import (
	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	releaseVersion = "release.giantswarm.io/version"
)

func ToCluster(v interface{}) (capi.Cluster, error) {
	if v == nil {
		return capi.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}

	customObjectPointer, ok := v.(*capi.Cluster)
	if !ok {
		return capi.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func ToObjectWithConditions(v interface{}) (conditions.Object, error) {
	if v == nil {
		return nil, microerror.Maskf(wrongTypeError, "expected non-nil conditions.Object, got nil '%T'", v)
	}

	object, ok := v.(conditions.Object)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected 'conditions.Object', got '%T'", v)
	}

	return object, nil
}

func ReleaseVersion(object conditions.Object) string {
	return object.GetLabels()[releaseVersion]
}
