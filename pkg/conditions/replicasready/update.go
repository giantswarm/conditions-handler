package replicasready

import (
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

func update(machinePool *capiexp.MachinePool) {
	// If value is not specified when MachinePool CR is created, the default
	// value is ensured in azure-admission-controller.
	desiredReplicas := *machinePool.Spec.Replicas

	if desiredReplicas > machinePool.Status.Replicas {
		capiconditions.MarkFalse(
			machinePool,
			capiexp.ReplicasReadyCondition,
			capiexp.WaitingForReplicasReadyReason,
			capi.ConditionSeverityWarning,
			"Desired number of replicas is %d, but found %d",
			desiredReplicas,
			machinePool.Status.Replicas)
		return
	}

	// We have found the desired number of replicas.

	// Now check if all found nodes are ready or not, and if all node references
	// are set.
	if machinePool.Status.Replicas != machinePool.Status.ReadyReplicas ||
		len(machinePool.Status.NodeRefs) != int(machinePool.Status.ReadyReplicas) {
		capiconditions.MarkFalse(
			machinePool,
			capiexp.ReplicasReadyCondition,
			capiexp.WaitingForReplicasReadyReason,
			capi.ConditionSeverityWarning,
			"%d/%d replicas are ready, %d/%d node references set",
			machinePool.Status.ReadyReplicas,
			machinePool.Status.Replicas,
			len(machinePool.Status.NodeRefs),
			machinePool.Status.Replicas)
		return
	}

	// Desired number of replicas is ready and all node references are set.
	capiconditions.MarkTrue(machinePool, capiexp.ReplicasReadyCondition)
}
