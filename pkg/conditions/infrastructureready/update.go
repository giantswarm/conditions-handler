package infrastructureready

import (
	"fmt"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

const (
	// Waiting time during which InfrastructureReady is set to False with
	// severity Info. After this threshold time, severity Warning is used.
	WaitingForInfrastructureWarningThresholdTime = 10 * time.Minute
)

// updateInfrastructureReadyCondition sets InfrastructureReady condition on specified
// object by mirroring Ready condition from specified infrastructure object.
//
// If specified infrastructure object is nil, object InfrastructureReady will
// be set with condition False and Reason InfrastructureObjectNotFoundReason.
//
// If specified infrastructure object's Ready condition is not set, object
// InfrastructureReady will be set with condition False and reason
// WaitingForInfrastructure.
func updateInfrastructureReadyCondition(object conditions.Object, infrastructureObject capiconditions.Getter) {
	if infrastructureObject == nil {
		warningMessage :=
			"Corresponding provider-specific infrastructure object of " +
				"type %T is not found for specified %T object %s/%s"

		capiconditions.MarkFalse(
			object,
			capi.InfrastructureReadyCondition,
			conditions.InfrastructureObjectNotFoundReason,
			capi.ConditionSeverityWarning,
			warningMessage,
			infrastructureObject, object, object.GetNamespace(), object.GetName())

		return
	}

	objectAge := time.Since(object.GetCreationTimestamp().Time)
	var fallbackSeverity capi.ConditionSeverity
	fallbackWarningMessage := ""
	if objectAge > WaitingForInfrastructureWarningThresholdTime {
		// Provider-specific infrastructure object should be reconciled soon
		// after it has been created. If it's Ready condition is not set within
		// 10 minutes, that means that something might be wrong, so we set
		// object's InfrastructureReady condition with severity Warning.
		fallbackSeverity = capi.ConditionSeverityWarning
		fallbackWarningMessage = fmt.Sprintf(" for more than %s", objectAge)
	} else {
		// Otherwise, if it has been less than 10 minutes since object's
		// creation, probably everything is good, and we just have to wait few
		// more minutes.
		// Note: Upstream Cluster API implementation always just sets severity
		// Info when infrastructure object's Ready condition is not set.
		fallbackSeverity = capi.ConditionSeverityInfo
	}

	fallbackToFalse := capiconditions.WithFallbackValue(
		false,
		capi.WaitingForInfrastructureFallbackReason,
		fallbackSeverity,
		fmt.Sprintf("Waiting for infrastructure object of type %T to have Ready condition set%s",
			infrastructureObject, fallbackWarningMessage))

	capiconditions.SetMirror(object, capi.InfrastructureReadyCondition, infrastructureObject, fallbackToFalse)
}
