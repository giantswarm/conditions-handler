package key

import (
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
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
