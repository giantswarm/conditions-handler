package infrastructureready

import (
	"context"
	"fmt"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexternal "sigs.k8s.io/cluster-api/controllers/external"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/errors"
)

const (
	deprecatedProviderInfrastructureReadyConditionType capi.ConditionType = "ProviderInfrastructureReady"
)

// update sets InfrastructureReady condition on specified object by mirroring
// Ready condition from specified infrastructure object.
//
// If specified infrastructure object is nil, object InfrastructureReady will
// be set with condition False and Reason InfrastructureObjectNotFoundReason.
//
// If specified infrastructure object's Ready condition is not set, object
// InfrastructureReady will be set with condition False and reason
// WaitingForInfrastructure.
func (h *Handler) update(ctx context.Context, object objectWithInfrastructureRef) error {
	// We need to remove already deprecated ProviderInfrastructureReady condition
	// from clusters that are already upgraded to node pools release, as we are
	// now using InfrastructureReady that is defined in the upstream Cluster API
	// project.
	removeDeprecatedProviderInfrastructureReadyCondition(object)

	gvk := object.GetObjectKind().GroupVersionKind()
	gvkString := fmt.Sprintf("%s (%s)", gvk.Kind, gvk.GroupVersion().String())

	if object.GetInfrastructureRef() == nil {
		warningMessage :=
			"%s object '%s/%s' does not have infrastructure reference set"

		capiconditions.MarkFalse(
			object,
			capi.InfrastructureReadyCondition,
			conditions.InfrastructureReferenceNotSetReason,
			capi.ConditionSeverityWarning,
			warningMessage,
			gvkString,
			object.GetNamespace(),
			object.GetName())

		return nil
	}

	infrastructureObject, err := h.getInfrastructureObject(ctx, object)
	if errors.IsFailedToRetrieveExternalObject(err) {
		warningMessage :=
			"Corresponding provider-specific infrastructure object '%s/%s' " +
				"is not found for %s object '%s/%s'"

		capiconditions.MarkFalse(
			object,
			capi.InfrastructureReadyCondition,
			conditions.InfrastructureObjectNotFoundReason,
			capi.ConditionSeverityWarning,
			warningMessage,
			object.GetInfrastructureRef().Namespace,
			object.GetInfrastructureRef().Name,
			gvkString,
			object.GetNamespace(),
			object.GetName())

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	fallbackToFalse := capiconditions.WithFallbackValue(
		false,
		capi.WaitingForInfrastructureFallbackReason,
		capi.ConditionSeverityWarning,
		fmt.Sprintf("Waiting for infrastructure object '%s/%s' of kind %s to have Ready condition set",
			object.GetInfrastructureRef().Namespace,
			object.GetInfrastructureRef().Name,
			object.GetInfrastructureRef().Kind))

	capiconditions.SetMirror(object, capi.InfrastructureReadyCondition, infrastructureObject, fallbackToFalse)

	// Update deprecated status fields
	infrastructureReady := conditions.IsInfrastructureReadyTrue(object)
	object.SetStatusInfrastructureReady(infrastructureReady)
	return nil
}

func removeDeprecatedProviderInfrastructureReadyCondition(object objectWithInfrastructureRef) {
	deprecatedProviderInfrastructureReadyCondition := capiconditions.Get(object, deprecatedProviderInfrastructureReadyConditionType)
	if deprecatedProviderInfrastructureReadyCondition == nil {
		return
	}

	allConditions := object.GetConditions()
	var filteredConditions capi.Conditions

	for _, c := range allConditions {
		if c.Type != deprecatedProviderInfrastructureReadyConditionType {
			filteredConditions = append(filteredConditions, c)
		}
	}

	object.SetConditions(filteredConditions)
}

func (h *Handler) getInfrastructureObject(ctx context.Context, object objectWithInfrastructureRef) (capiconditions.Getter, error) {
	infrastructureRef := object.GetInfrastructureRef()
	if infrastructureRef == nil {
		// Infrastructure object is not set
		return nil, nil
	}

	infrastructureObject, err := capiexternal.Get(ctx, h.ctrlClient, object.GetInfrastructureRef(), object.GetNamespace())
	if apierrors.IsNotFound(err) {
		// Infrastructure object is not found, here we don't care why
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// Wrap unstructured object into a capiconditions.Getter
	infrastructureObjectGetter := capiconditions.UnstructuredGetter(infrastructureObject)

	return infrastructureObjectGetter, nil
}
