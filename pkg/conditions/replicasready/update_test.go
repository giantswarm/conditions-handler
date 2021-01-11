package replicasready

import (
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

type updateTestCase struct {
	name                string
	machinePoolManifest string
	expectedCondition   capi.Condition
}

func TestUpdate(t *testing.T) {
	testCases := []updateTestCase{
		{
			name:                "0: MachinePool with Spec.Replicas greater than Status.Replicas (not all replicas are discovered)",
			machinePoolManifest: "machinepool-desired-gt-observed.yaml",
			expectedCondition: capi.Condition{
				Type:     capiexp.ReplicasReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capiexp.WaitingForReplicasReadyReason,
				Message:  "Desired number of replicas is 3, but found 1",
			},
		},
		{
			name:                "1: MachinePool with Status.Replicas greater than Status.ReadyReplicas (not all replicas are ready)",
			machinePoolManifest: "machinepool-replicas-not-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capiexp.ReplicasReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capiexp.WaitingForReplicasReadyReason,
				Message:  "1/3 replicas are ready, 2/3 node references set",
			},
		},
		{
			name:                "2: MachinePool with Status.NodeRef not fully set",
			machinePoolManifest: "machinepool-noderefs-not-set.yaml",
			expectedCondition: capi.Condition{
				Type:     capiexp.ReplicasReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capiexp.WaitingForReplicasReadyReason,
				Message:  "3/3 replicas are ready, 1/3 node references set",
			},
		},
		{
			name:                "3: MachinePool with replicas ready",
			machinePoolManifest: "machinepool-replicasready-true.yaml",
			expectedCondition: capi.Condition{
				Type:   capiexp.ReplicasReadyCondition,
				Status: corev1.ConditionTrue,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			t.Log(tc.name)
			machinePool := loadMachinePool(t, tc)

			// act
			update(&machinePool)
			replicasReady := capiconditions.Get(&machinePool, capiexp.ReplicasReadyCondition)

			// assert
			if replicasReady == nil {
				t.Logf(
					"Condition %s not set, expected %s",
					capiexp.ReplicasReadyCondition,
					internal.SprintComparedCondition(&tc.expectedCondition))
				t.Fail()
			} else {
				if !internal.AreEqualWithIgnoringLastTransitionTime(replicasReady, &tc.expectedCondition) {
					t.Logf(
						"expected %s, got %s",
						internal.SprintComparedCondition(&tc.expectedCondition),
						internal.SprintComparedCondition(replicasReady))
					t.Fail()
				}
			}
		})
	}
}

func loadMachinePool(t *testing.T, tc updateTestCase) capiexp.MachinePool {
	machinePoolCRPath := filepath.Join("testdata", tc.machinePoolManifest)
	o, err := internal.LoadCR(machinePoolCRPath)
	if err != nil {
		t.Fatal(err)
	}

	machinePool, ok := o.(*capiexp.MachinePool)
	if !ok {
		t.Fatalf("couldn't cast object %T to Cluster", o)
	}

	return *machinePool
}
