package infrastructureready

import (
	"testing"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

func TestUpdateInfrastructureReady(t *testing.T) {
	testCases := []struct {
		name                 string
		cluster              *capi.Cluster
		infrastructureObject conditions.Object
		expectedCondition    capi.Condition
	}{
		{
			name:                 "case 0: For nil infrastructure object, it sets InfrastructureReady(Status=False, Severity=Warning, Reason=InfrastructureObjectNotFound)",
			cluster:              &capi.Cluster{},
			infrastructureObject: nil,
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.InfrastructureObjectNotFoundReason,
			},
		},
		{
			name: "case 1: For 5min old Cluster and infrastructure object w/o Ready, it sets InfrastructureReady(Status=False, Severity=Info, Reason=WaitingForInfrastructureFallback)",
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-WaitingForInfrastructureWarningThresholdTime / 2)),
				},
			},
			infrastructureObject: &internal.MockProviderCluster{
				Status: internal.MockProviderClusterStatus{
					Conditions: capi.Conditions{},
				},
			},
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   capi.WaitingForInfrastructureFallbackReason,
			},
		},
		{
			name: "case 2: For 20min old Cluster and infrastructure object w/o Ready, it sets InfrastructureReady status to False, Severity=Warning, Reason=WaitingForInfrastructureFallback",
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * WaitingForInfrastructureWarningThresholdTime)),
				},
			},
			infrastructureObject: &internal.MockProviderCluster{
				Status: internal.MockProviderClusterStatus{
					Conditions: capi.Conditions{},
				},
			},
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capi.WaitingForInfrastructureFallbackReason,
			},
		},
		{
			name:    "case 3: For infrastructure object w/ Ready(Status=False), it sets InfrastructureReady(Status=False)",
			cluster: &capi.Cluster{},
			infrastructureObject: &internal.MockProviderCluster{
				Status: internal.MockProviderClusterStatus{
					Conditions: capi.Conditions{
						{
							Type:   capi.ReadyCondition,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			expectedCondition: capi.Condition{
				Type:   capi.InfrastructureReadyCondition,
				Status: corev1.ConditionFalse,
			},
		},
		{
			name:    "case 4: For infrastructure object w/ Ready(Status=True), it sets InfrastructureReady(Status=True)",
			cluster: &capi.Cluster{},
			infrastructureObject: &internal.MockProviderCluster{
				Status: internal.MockProviderClusterStatus{
					Conditions: capi.Conditions{
						{
							Type:   capi.ReadyCondition,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedCondition: capi.Condition{
				Type:   capi.InfrastructureReadyCondition,
				Status: corev1.ConditionTrue,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)

			// act
			updateInfrastructureReadyCondition(tc.cluster, tc.infrastructureObject)

			// assert
			infrastructureReady, ok := conditions.GetInfrastructureReady(tc.cluster)
			if ok {
				if !conditions.AreEquivalent(&infrastructureReady, &tc.expectedCondition) {
					t.Logf(
						"InfrastructureReady was not set correctly, got %s, expected %s",
						internal.SprintComparedCondition(&infrastructureReady),
						internal.SprintComparedCondition(&tc.expectedCondition))
					t.Fail()
				}
			} else {
				t.Log("InfrastructureReady was not set")
				t.Fail()
			}
		})
	}
}
