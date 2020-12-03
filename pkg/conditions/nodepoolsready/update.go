package nodepoolsready

import (
	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

// update sets NodePoolsReady condition on specified cluster
// by aggregating Ready conditions from specified node pool objects.
//
// If node pool objects are found, cluster NodePoolsReady is set to with status
// False and reason NodePoolsNotFoundReason.
func update(cluster *capi.Cluster, nodePools []capiconditions.Getter) {
	if len(nodePools) == 0 {
		capiconditions.MarkFalse(
			cluster,
			conditions.NodePoolsReady,
			conditions.NodePoolsNotFoundReason,
			capi.ConditionSeverityInfo,
			"Node pools are not found for Cluster %s/%s",
			cluster.Namespace, cluster.Name)
		return
	}

	capiconditions.SetAggregate(
		cluster,
		conditions.NodePoolsReady,
		nodePools,
		capiconditions.WithStepCounter(),
		capiconditions.AddSourceRef())
}
