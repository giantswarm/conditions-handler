package controlplaneready

import (
	"testing"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

func TestUpdateControlPlaneReady(t *testing.T) {
	testCases := []struct {
		name               string
		cluster            *capi.Cluster
		controlPlaneObject conditions.Object
		expectedCondition  capi.Condition
	}{
		{
			name:               "case 0: For nil control plane object, it sets ControlPlaneReady status to False, Severity=Warning, Reason=ControlPlaneObjectNotFound",
			cluster:            &capi.Cluster{},
			controlPlaneObject: nil,
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.ControlPlaneObjectNotFoundReason,
			},
		},
		{
			name: "case 1: For 5min old Cluster and control plane object w/o Ready, it sets ControlPlaneReady status to False, Severity=Info, Reason=WaitingForControlPlaneFallback",
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-conditions.WaitingForControlPlaneWarningThresholdTime / 2)),
				},
			},
			controlPlaneObject: &capi.Machine{
				Status: capi.MachineStatus{
					Conditions: capi.Conditions{},
				},
			},
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   capi.WaitingForControlPlaneFallbackReason,
			},
		},
		{
			name: "case 2: For 20min old Cluster and control plane object w/o Ready, it sets ControlPlaneReady status to False, Severity=Warning, Reason=WaitingForControlPlaneFallback",
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * conditions.WaitingForControlPlaneWarningThresholdTime)),
				},
			},
			controlPlaneObject: &capi.Machine{
				Status: capi.MachineStatus{
					Conditions: capi.Conditions{},
				},
			},
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capi.WaitingForControlPlaneFallbackReason,
			},
		},
		{
			name:    "case 3: For control plane object w/ Ready(Status=False), it sets ControlPlaneReady(Status=False)",
			cluster: &capi.Cluster{},
			controlPlaneObject: &capi.Machine{
				Status: capi.MachineStatus{
					Conditions: capi.Conditions{
						{
							Type:   capi.ReadyCondition,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			expectedCondition: capi.Condition{
				Type:   capi.ControlPlaneReadyCondition,
				Status: corev1.ConditionFalse,
			},
		},
		{
			name:    "case 4: For control plane object w/ Ready(Status=True), it sets ControlPlaneReady(Status=True)",
			cluster: &capi.Cluster{},
			controlPlaneObject: &capi.Machine{
				Status: capi.MachineStatus{
					Conditions: capi.Conditions{
						{
							Type:   capi.ReadyCondition,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedCondition: capi.Condition{
				Type:   capi.ControlPlaneReadyCondition,
				Status: corev1.ConditionTrue,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.name)

			// act
			update(tc.cluster, tc.controlPlaneObject)

			// assert
			controlPlaneReady, ok := conditions.GetControlPlaneReady(tc.cluster)
			if ok {
				if !conditions.AreEquivalent(&controlPlaneReady, &tc.expectedCondition) {
					t.Logf(
						"ControlPlaneReady was not set correctly, got %s, expected %s",
						internal.SprintComparedCondition(&controlPlaneReady),
						internal.SprintComparedCondition(&tc.expectedCondition))
					t.Fail()
				}
			} else {
				t.Log("ControlPlaneReady was not set")
				t.Fail()
			}
		})
	}
}
