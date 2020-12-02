package controlplaneready

import (
	"fmt"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

// update sets ControlPlaneReady condition on specified
// cluster by mirroring Ready condition from specified control plane object.
//
// If specified control plane object is nil, object ControlPlaneReady will
// be set with condition False and Reason ControlPlaneObjectNotFoundReason.
//
// If specified control plane object's Ready condition is not set, object
// ControlPlaneReady will be set with condition False and reason
// WaitingForControlPlane.
func update(cluster *capi.Cluster, controlPlaneObject capiconditions.Getter) {
	if controlPlaneObject == nil {
		warningMessage :=
			"Control plane object of type %T is not found for specified %T object %s/%s"

		capiconditions.MarkFalse(
			cluster,
			capi.ControlPlaneReadyCondition,
			conditions.ControlPlaneObjectNotFoundReason,
			capi.ConditionSeverityWarning,
			warningMessage,
			controlPlaneObject, cluster, cluster.GetNamespace(), cluster.GetName())

		return
	}

	clusterAge := time.Since(cluster.GetCreationTimestamp().Time)
	var fallbackSeverity capi.ConditionSeverity
	fallbackWarningMessage := ""
	if clusterAge > conditions.WaitingForControlPlaneWarningThresholdTime {
		// Control plane should be reconciled soon after it has been created.
		// If it's Ready condition is not set within 10 minutes, that means
		// that something might be wrong, so we set object's ControlPlaneReady
		// condition with severity Warning.
		fallbackSeverity = capi.ConditionSeverityWarning
		fallbackWarningMessage = fmt.Sprintf(" for more than %s", clusterAge)
	} else {
		// Otherwise, if it has been less than 10 minutes since object's
		// creation, probably everything is good, and we just have to wait few
		// more minutes.
		// Note: Upstream Cluster API implementation always just sets severity
		// Info when control plane object's Ready condition is not set.
		fallbackSeverity = capi.ConditionSeverityInfo
	}

	fallbackToFalse := capiconditions.WithFallbackValue(
		false,
		capi.WaitingForControlPlaneFallbackReason,
		fallbackSeverity,
		fmt.Sprintf("Waiting for control plane object of type %T to have Ready condition set%s",
			controlPlaneObject, fallbackWarningMessage))

	capiconditions.SetMirror(cluster, capi.ControlPlaneReadyCondition, controlPlaneObject, fallbackToFalse)

	// Update deprecated status fields
	cluster.Status.ControlPlaneReady = conditions.IsControlPlaneReadyTrue(cluster)
	if !cluster.Status.ControlPlaneInitialized {
		cluster.Status.ControlPlaneInitialized = cluster.Status.ControlPlaneReady
	}
}
