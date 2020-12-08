package upgrading

import (
	"testing"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

func TestMarkUpgradingTrue(t *testing.T) {
	testName := "MarkUpgradingTrue sets Upgrading condition status to True"
	t.Run(testName, func(t *testing.T) {
		// arrange
		t.Log(testName)
		cluster := &capi.Cluster{}

		// act
		MarkUpgradingTrue(cluster)

		// assert
		if !conditions.IsUpgradingTrue(cluster) {
			gotMessage := internal.SprintComparedCondition(capiconditions.Get(cluster, conditions.Upgrading))
			t.Logf("expected that Upgrading condition status is set to True, got %s", gotMessage)
			t.Fail()
		}
	})
}

func TestMarkUpgradingFalseWithUpgradeCompleted(t *testing.T) {
	testName := "Upgrading condition is set with Status=False, Severity=Info, Reason=UpgradeCompleted"
	t.Run(testName, func(t *testing.T) {
		// arrange
		t.Log(testName)
		cluster := &capi.Cluster{}

		// act
		MarkUpgradingFalseWithUpgradeCompleted(cluster)

		// assert
		expected := conditions.IsUpgradingFalse(cluster, conditions.WithSeverityInfo(), conditions.WithUpgradeCompletedReason())
		if !expected {
			gotMessage := internal.SprintComparedCondition(capiconditions.Get(cluster, conditions.Upgrading))
			t.Logf(
				"expected that Upgrading condition is set with Status=False, Severity=Info, Reason=UpgradeCompleted, got %s",
				gotMessage)
			t.Fail()
		}
	})
}

func TestMarkUpgradingFalseWithUpgradeNotStarted(t *testing.T) {
	testName := "Upgrading condition is set with Status=False, Severity=Info, Reason=UpgradeNotStarted"
	t.Run(testName, func(t *testing.T) {
		// arrange
		t.Log(testName)
		cluster := &capi.Cluster{}

		// act
		MarkUpgradingFalseWithUpgradeNotStarted(cluster)

		// assert
		expected := conditions.IsUpgradingFalse(cluster, conditions.WithSeverityInfo(), conditions.WithUpgradeNotStartedReason())
		if !expected {
			gotMessage := internal.SprintComparedCondition(capiconditions.Get(cluster, conditions.Upgrading))
			t.Logf(
				"expected that Upgrading condition is set with Status=False, Severity=Info, Reason=UpgradeNotStarted, got %s",
				gotMessage)
			t.Fail()
		}
	})
}
