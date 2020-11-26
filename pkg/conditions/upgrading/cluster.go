package upgrading

import (
	"strings"

	"github.com/giantswarm/conditions/pkg/conditions"

	"github.com/giantswarm/conditions-handler/pkg/key"
)

const (
	// upgradingToNodePools is set to True during the first cluster upgrade to node pools release.
	upgradingToNodePools = "release.giantswarm.io/upgrading-to-node-pools"
)

// isFirstNodePoolUpgradeInProgress checks if the cluster is being upgraded
// from an old/legacy release to the node pools release.
func isFirstNodePoolUpgradeInProgress(object conditions.Object) bool {
	cluster, err := key.ToCluster(object)
	if err != nil {
		return false
	}

	upgradingToNodePools, isUpgradingToNodePoolsSet := cluster.GetAnnotations()[upgradingToNodePools]
	return isUpgradingToNodePoolsSet && strings.ToLower(upgradingToNodePools) == "true"
}
