package creating

import (
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

// MarkCreatingTrue sets Creating condition with status True.
func MarkCreatingTrue(object conditions.Object) {
	capiconditions.MarkTrue(object, conditions.Creating)
}

// MarkCreatingFalseWithCreationCompleted sets Creating condition with status
// False, reason CreationCompleted, severity Info and a message informing how
// long the creation took.
func MarkCreatingFalseWithCreationCompleted(object conditions.Object) {
	creationDuration := time.Since(object.GetCreationTimestamp().Time)
	capiconditions.MarkFalse(
		object,
		conditions.Creating,
		conditions.CreationCompletedReason,
		capi.ConditionSeverityInfo,
		"Cluster creation has been completed in %s",
		creationDuration)
}

// MarkCreatingFalseForExistingObject sets Creating condition with status
// False, reason ExistingObject, severity Info and a message informing that the
// object was already created.
func MarkCreatingFalseForExistingObject(object conditions.Object) {
	capiconditions.MarkFalse(
		object,
		conditions.Creating,
		conditions.ExistingObjectReason,
		capi.ConditionSeverityInfo,
		"Object was already created")
}
