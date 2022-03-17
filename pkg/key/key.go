package key

import (
	"strings"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"

	"github.com/giantswarm/conditions-handler/pkg/errors"
	"github.com/giantswarm/conditions-handler/pkg/internal"
)

const (
	releaseVersion = "release.giantswarm.io/version"
)

func ToClusterPointer(v interface{}) (*capi.Cluster, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got nil", &capi.Cluster{})
	}

	customObjectPointer, ok := v.(*capi.Cluster)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}

	return customObjectPointer, nil
}

func ToMachinePoolPointer(v interface{}) (*capiexp.MachinePool, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got nil", &capiexp.MachinePool{})
	}

	customObjectPointer, ok := v.(*capiexp.MachinePool)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected '%T', got '%T'", &capiexp.MachinePool{}, v)
	}

	return customObjectPointer, nil
}

func ToObjectWithConditions(v interface{}) (conditions.Object, error) {
	if v == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected non-nil conditions.Object, got nil '%T'", v)
	}

	object, ok := v.(conditions.Object)
	if !ok {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected 'conditions.Object', got '%T'", v)
	}

	return object, nil
}

func ReleaseVersion(object conditions.Object) string {
	return object.GetLabels()[releaseVersion]
}

// isFirstNodePoolUpgradeInProgress checks if the cluster is being upgraded
// from an old/legacy release to the node pools release.
func IsFirstNodePoolUpgradeInProgress(object conditions.Object) bool {
	cluster, err := ToClusterPointer(object)
	if err != nil {
		return false
	}

	upgradingToNodePools, isUpgradingToNodePoolsSet := cluster.GetAnnotations()[internal.UpgradingToNodePools]
	return isUpgradingToNodePoolsSet && strings.ToLower(upgradingToNodePools) == "true"
}
