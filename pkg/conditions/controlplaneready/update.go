package controlplaneready

import (
	"context"
	"fmt"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexternal "sigs.k8s.io/cluster-api/controllers/external"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/errors"
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
	gvk := cluster.GetObjectKind().GroupVersionKind()
	gvkString := fmt.Sprintf("%s (%s)", gvk.Kind, gvk.GroupVersion().String())

	if cluster.Spec.ControlPlaneRef == nil {
		warningMessage :=
			"Control plane reference is not set for specified %s object '%s/%s'"

		capiconditions.MarkFalse(
			cluster,
			capi.ControlPlaneReadyCondition,
			conditions.ControlPlaneReferenceNotSetReason,
			capi.ConditionSeverityWarning,
			warningMessage,
			gvkString,
			cluster.GetNamespace(),
			cluster.GetName())

		return nil
	}

	controlPlaneObject, err := h.getControlPlaneObject(ctx, cluster)
	if errors.IsFailedToRetrieveExternalObject(err) {
		warningMessage :=
			"Control plane object '%s/%s' of kind %s is not found for specified %s object '%s/%s'"

		capiconditions.MarkFalse(
			cluster,
			capi.ControlPlaneReadyCondition,
			conditions.ControlPlaneObjectNotFoundReason,
			capi.ConditionSeverityWarning,
			warningMessage,
			cluster.Spec.ControlPlaneRef.Namespace,
			cluster.Spec.ControlPlaneRef.Name,
			cluster.Spec.ControlPlaneRef.Kind,
			gvkString,
			cluster.GetNamespace(),
			cluster.GetName())

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	fallbackToFalse := capiconditions.WithFallbackValue(
		false,
		capi.WaitingForControlPlaneFallbackReason,
		capi.ConditionSeverityWarning,
		fmt.Sprintf("Waiting for control plane object '%s/%s' of kind %s to have Ready condition set",
			cluster.Spec.ControlPlaneRef.Namespace,
			cluster.Spec.ControlPlaneRef.Name,
			cluster.Spec.ControlPlaneRef.Kind))

	capiconditions.SetMirror(cluster, capi.ControlPlaneReadyCondition, controlPlaneObject, fallbackToFalse)

	// Update deprecated status fields
	cluster.Status.ControlPlaneReady = conditions.IsControlPlaneReadyTrue(cluster)

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
