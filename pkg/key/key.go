package key

import (
	"strings"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/conditions-handler/pkg/internal"
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

// isFirstNodePoolUpgradeInProgress checks if the cluster is being upgraded
// from an old/legacy release to the node pools release.
func IsFirstNodePoolUpgradeInProgress(object conditions.Object) bool {
	cluster, err := ToCluster(object)
	if err != nil {
		return false
	}

	upgradingToNodePools, isUpgradingToNodePoolsSet := cluster.GetAnnotations()[internal.UpgradingToNodePools]
	return isUpgradingToNodePoolsSet && strings.ToLower(upgradingToNodePools) == "true"
}
