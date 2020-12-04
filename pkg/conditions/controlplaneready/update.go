package controlplaneready

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexternal "sigs.k8s.io/cluster-api/controllers/external"
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
func (h *Handler) update(ctx context.Context, cluster *capi.Cluster) error {
	clusterAge := time.Since(cluster.GetCreationTimestamp().Time)
	var severity capi.ConditionSeverity
	ageWarningMessage := ""
	if clusterAge > conditions.WaitingForControlPlaneWarningThresholdTime {
		// Control plane should be reconciled soon after it has been created.
		// If it's Ready condition is not set within 10 minutes, that means
		// that something might be wrong, so we set object's ControlPlaneReady
		// condition with severity Warning.
		severity = capi.ConditionSeverityWarning
		ageWarningMessage = fmt.Sprintf(" for more than %s", clusterAge)
	} else {
		// Otherwise, if it has been less than 10 minutes since object's
		// creation, probably everything is good, and we just have to wait few
		// more minutes.
		// Note: Upstream Cluster API implementation always just sets severity
		// Info when control plane object's Ready condition is not set.
		severity = capi.ConditionSeverityInfo
	}

	if cluster.Spec.ControlPlaneRef == nil {
		warningMessage :=
			"Control plane reference is not set for specified Cluster object '%s/%s'%s"

		capiconditions.MarkFalse(
			cluster,
			capi.ControlPlaneReadyCondition,
			conditions.ControlPlaneReferenceNotSetReason,
			severity,
			warningMessage,
			cluster.GetNamespace(),
			cluster.GetName(),
			ageWarningMessage)

		return nil
	}

	controlPlaneObject, err := h.getControlPlaneObject(ctx, cluster)
	if IsFailedToRetrieveExternalObject(err) {
		warningMessage :=
			"Control plane object '%s/%s' of type %s is not found for specified Cluster object '%s/%s'%s"

		capiconditions.MarkFalse(
			cluster,
			capi.ControlPlaneReadyCondition,
			conditions.ControlPlaneObjectNotFoundReason,
			severity,
			warningMessage,
			cluster.Spec.ControlPlaneRef.Namespace,
			cluster.Spec.ControlPlaneRef.Name,
			cluster.Spec.ControlPlaneRef.Kind,
			cluster.GetNamespace(),
			cluster.GetName(),
			ageWarningMessage)

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	fallbackToFalse := capiconditions.WithFallbackValue(
		false,
		capi.WaitingForControlPlaneFallbackReason,
		severity,
		fmt.Sprintf("Waiting for control plane object '%s/%s' of type %s to have Ready condition set%s",
			cluster.Spec.ControlPlaneRef.Namespace,
			cluster.Spec.ControlPlaneRef.Name,
			cluster.Spec.ControlPlaneRef.Kind,
			ageWarningMessage))

	capiconditions.SetMirror(cluster, capi.ControlPlaneReadyCondition, controlPlaneObject, fallbackToFalse)

	// Update deprecated status fields
	cluster.Status.ControlPlaneReady = conditions.IsControlPlaneReadyTrue(cluster)
	if !cluster.Status.ControlPlaneInitialized {
		cluster.Status.ControlPlaneInitialized = cluster.Status.ControlPlaneReady
	}

	return nil
}

func (h *Handler) getControlPlaneObject(ctx context.Context, cluster *capi.Cluster) (capiconditions.Getter, error) {
	if cluster.Spec.ControlPlaneRef == nil {
		return nil, nil
	}

	controlPlaneObject, err := capiexternal.Get(ctx, h.ctrlClient, cluster.Spec.ControlPlaneRef, cluster.Namespace)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	controlPlaneObjectGetter := capiconditions.UnstructuredGetter(controlPlaneObject)
	return controlPlaneObjectGetter, nil
}
