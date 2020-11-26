package upgrading

import (
	"fmt"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
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
		upgradeTimeMessage = fmt.Sprintf("in %s", upgradeDuration)
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
