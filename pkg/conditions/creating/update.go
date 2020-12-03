package creating

import (
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
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
		"Creation has been completed in %s",
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

func update(object conditions.Object) {
	// Creating condition is not set or it has Unknown status, let's set it for
	// the first time.
	if conditions.IsCreatingUnknown(object) {
		initialize(object)
		return
	}

	// Creating condition is False, which means that the cluster or node pool
	// creation is completed, so we don't have to update it anymore.
	if conditions.IsCreatingFalse(object) {
		return
	}

	// Creating condition has Status set to True, let's check if the creation
	// has been completed.
	markCreatingFalseIfCreationCompleted(object)
}

func initialize(object conditions.Object) {
	_, isLastDeployedReleaseVersionSet := object.GetAnnotations()[internal.LastDeployedReleaseVersion]

	if isLastDeployedReleaseVersionSet || key.IsFirstNodePoolUpgradeInProgress(object) {
		MarkCreatingFalseForExistingObject(object)
	} else {
		MarkCreatingTrue(object)
	}
}

func markCreatingFalseIfCreationCompleted(object conditions.Object) {
	lastDeployedReleaseVersion, isLastDeployedReleaseVersionSet := object.GetAnnotations()[internal.LastDeployedReleaseVersion]
	if !isLastDeployedReleaseVersionSet {
		// Cluster or node pool creation is not completed, since there is no
		// last deployed release version set.
		return
	}

	desiredReleaseVersion := key.ReleaseVersion(object)

	if lastDeployedReleaseVersion == desiredReleaseVersion {
		// Desired version has been reached, cluster or node pool creation has
		// been completed! :)
		MarkCreatingFalseWithCreationCompleted(object)
	}
}
