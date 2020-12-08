package creating

import (
	"testing"

	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

func TestMarkCreatingTrue(t *testing.T) {
	testName := "MarkCreatingTrue sets Creating condition status to True"
	t.Run(testName, func(t *testing.T) {
		// arrange
		t.Log(testName)
		cluster := &capi.Cluster{}

		// act
		MarkCreatingTrue(cluster)

		// assert
		expected := conditions.IsCreatingTrue(cluster)
		if !expected {
			gotMessage := internal.SprintComparedCondition(capiconditions.Get(cluster, conditions.Creating))
			t.Logf("expected that Creating condition status is set to True, got %s", gotMessage)
			t.Fail()
		}
	})
}

func TestMarkCreatingFalseWithCreationCompleted(t *testing.T) {
	testName := "Creating condition is set with Status=False, Severity=Info, Reason=CreationCompleted"
	t.Run(testName, func(t *testing.T) {
		// arrange
		t.Log(testName)
		cluster := &capi.Cluster{}

		// act
		MarkCreatingFalseWithCreationCompleted(cluster)

		// assert
		expected := conditions.IsCreatingFalse(cluster, conditions.WithSeverityInfo(), conditions.WithCreationCompletedReason())
		if !expected {
			gotMessage := internal.SprintComparedCondition(capiconditions.Get(cluster, conditions.Creating))
			t.Logf(
				"expected that Creating condition is set with Status=False, Severity=Info, Reason=CreationCompleted, got %s",
				gotMessage)
			t.Fail()
		}
	})
}

func TestMarkCreatingFalseForExistingObject(t *testing.T) {
	testName := "Creating condition is set with Status=False, Severity=Info, Reason=ExistingObject"
	t.Run(testName, func(t *testing.T) {
		// arrange
		t.Log(testName)
		cluster := &capi.Cluster{}

		// act
		MarkCreatingFalseForExistingObject(cluster)

		// assert
		expected := conditions.IsCreatingFalse(cluster, conditions.WithSeverityInfo(), conditions.WithExistingObjectReason())
		if !expected {
			gotMessage := internal.SprintComparedCondition(capiconditions.Get(cluster, conditions.Creating))
			t.Logf(
				"expected that Creating condition is set with Status=False, Severity=Info, Reason=ExistingObject, got %s",
				gotMessage)
			t.Fail()
		}
	})
}
