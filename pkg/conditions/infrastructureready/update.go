package infrastructureready

import (
	"fmt"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
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
func update(object objectWithInfrastructureRef, infrastructureObject capiconditions.Getter) {
	// We need to remove already deprecated ProviderInfrastructureReady condition
	// from clusters that are already upgraded to node pools release, as we are
	// now using InfrastructureReady that is defined in the upstream Cluster API
	// project.
	removeDeprecatedProviderInfrastructureReadyCondition(object)

	objectAge := time.Since(object.GetCreationTimestamp().Time)
	var severity capi.ConditionSeverity
	ageWarningMessage := ""
	if objectAge > conditions.WaitingForInfrastructureWarningThresholdTime {
		// Infrastructure reference should be set and provider-specific
		// infrastructure object should then be reconciled soon after it has
		// been created, which will set its Ready condition.
		// If that does not happen within 10 minutes, that means that something
		// might be wrong, so we set object's InfrastructureReady condition with
		// severity Warning.
		severity = capi.ConditionSeverityWarning
		ageWarningMessage = fmt.Sprintf(" for more than %s", objectAge)
	} else {
		// Otherwise, if it has been less than 10 minutes since object's
		// creation, probably everything is good, and we just have to wait few
		// more minutes.
		// Note: Upstream Cluster API implementation always just sets severity
		// Info when infrastructure object's Ready condition is not set.
		severity = capi.ConditionSeverityInfo
	}

	if object.GetInfrastructureRef() == nil {
		warningMessage :=
			"%s object %s/%s does not have infrastructure reference set%s"

		capiconditions.MarkFalse(
			object,
			capi.InfrastructureReadyCondition,
			conditions.InfrastructureReferenceNotSetReason,
			severity,
			warningMessage,
			object.GetObjectKind(),
			object.GetNamespace(),
			object.GetName(),
			ageWarningMessage)

		return
	}

	if infrastructureObject == nil {
		warningMessage :=
			"Corresponding provider-specific infrastructure object '%s/%s' " +
				"is not found for %s object '%s/%s'%s"

		capiconditions.MarkFalse(
			object,
			capi.InfrastructureReadyCondition,
			conditions.InfrastructureObjectNotFoundReason,
			severity,
			warningMessage,
			object.GetInfrastructureRef().Namespace,
			object.GetInfrastructureRef().Name,
			object.GetObjectKind(),
			object.GetNamespace(),
			object.GetName(),
			ageWarningMessage)

		return
	}

	fallbackToFalse := capiconditions.WithFallbackValue(
		false,
		capi.WaitingForInfrastructureFallbackReason,
		severity,
		fmt.Sprintf("Waiting for infrastructure object '%s/%s' of type %s to have Ready condition set%s",
			object.GetInfrastructureRef().Namespace,
			object.GetInfrastructureRef().Name,
			object.GetInfrastructureRef().Kind,
			ageWarningMessage))

	capiconditions.SetMirror(object, capi.InfrastructureReadyCondition, infrastructureObject, fallbackToFalse)

	// Update deprecated status fields
	infrastructureReady := conditions.IsInfrastructureReadyTrue(object)
	object.SetStatusInfrastructureReady(infrastructureReady)
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
