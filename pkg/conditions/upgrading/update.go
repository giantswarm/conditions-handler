package upgrading

import (
	"fmt"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
)

// MarkUpgradingTrue sets Upgrading condition with status True.
func MarkUpgradingTrue(object conditions.Object) {
	capiconditions.MarkTrue(object, conditions.Upgrading)
}

// MarkUpgradingFalseWithUpgradeCompleted sets Upgrading condition with status
// False, reason UpgradeCompleted, severity Info and a message informing how
// long the upgrade took.
func MarkUpgradingFalseWithUpgradeCompleted(object conditions.Object) {
	var upgradeTimeMessage string
	currentUpgradingCondition := capiconditions.Get(object, conditions.Upgrading)
	if currentUpgradingCondition != nil {
		upgradeDuration := time.Since(currentUpgradingCondition.LastTransitionTime.Time)
		upgradeTimeMessage = fmt.Sprintf(" in %s", upgradeDuration)
	} else {
		upgradeTimeMessage = ", but upgrade duration cannot be determined"
	}

	capiconditions.MarkFalse(
		object,
		conditions.Upgrading,
		conditions.UpgradeCompletedReason,
		capi.ConditionSeverityInfo,
		"Upgrade has been completed%s",
		upgradeTimeMessage)
}

// MarkUpgradingFalseWithUpgradeNotStarted sets Upgrading condition with status
// False, reason ExistingObject, severity Info and a message informing that the
// upgrade has not been started.
func MarkUpgradingFalseWithUpgradeNotStarted(object conditions.Object) {
	capiconditions.MarkFalse(
		object,
		conditions.Upgrading,
		conditions.UpgradeNotStartedReason,
		capi.ConditionSeverityInfo,
		"Upgrade has not been started")
}

func update(object conditions.Object) {
	// Case 1: new cluster or node pool is just being created, no upgrade yet.
	if conditions.IsCreatingTrue(object) {
		MarkUpgradingFalseWithUpgradeNotStarted(object)
		return
	}

	// Case 2: Cluster-only check, first cluster upgrade to node pools release,
	// a bit of an edgecase, albeit an important one :)
	if key.IsFirstNodePoolUpgradeInProgress(object) {
		if !conditions.IsUpgradingTrue(object) {
			MarkUpgradingTrue(object)
		}
		return
	}

	// Let's check what was the last release version that we successfully deployed.
	lastDeployedReleaseVersion, isLastDeployedReleaseVersionSet := object.GetAnnotations()[internal.LastDeployedReleaseVersion]
	if !isLastDeployedReleaseVersionSet {
		// Case 3: Last deployed release version annotation is not set at all,
		// which means that cluster or node pool creation has not completed, so
		// no upgrades yet.
		// This case should be already processed by Creating condition handler,
		// and we would have Case 1 from above, but here we check just in case,
		// since the reconciled object maybe does not have Creating condition
		// set.
		MarkUpgradingFalseWithUpgradeNotStarted(object)
		return
	}

	// Let's now check if desired release version is deployed.
	desiredReleaseVersion := key.ReleaseVersion(object)
	desiredReleaseVersionIsDeployed := lastDeployedReleaseVersion == desiredReleaseVersion

	currentUpgrading, isSet := conditions.GetUpgrading(object)

	if !isSet || conditions.IsUnknown(&currentUpgrading) {
		// Case 4: Cluster or node pool is still being created, or it's restored
		// from backup, this case should be very rare and almost never happen.
		if desiredReleaseVersionIsDeployed {
			MarkUpgradingFalseWithUpgradeNotStarted(object)
		} else {
			MarkUpgradingTrue(object)
		}
	} else if conditions.IsTrue(&currentUpgrading) && desiredReleaseVersionIsDeployed {
		// Case 5: Cluster or node pool was being upgraded.
		// Also, last deployed release version for this object is equal to the
		// desired release version, so we can conclude that the upgrade has been
		// completed.
		MarkUpgradingFalseWithUpgradeCompleted(object)
	} else if conditions.IsFalse(&currentUpgrading) && !desiredReleaseVersionIsDeployed {
		// Case 6: Cluster or node pool was not being upgraded.
		// Also, desired release for this cluster is different than the release
		// to which it was previously upgraded or with which was created, so we
		// can conclude that the cluster is in the Upgrading state.
		MarkUpgradingTrue(object)
	}
}
