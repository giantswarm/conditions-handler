package controlplaneready

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

type updateTestCase struct {
	name                 string
	clusterManifest      string
	controlPlaneManifest string
	expectedCondition    capi.Condition
}

func TestUpdateControlPlaneReady(t *testing.T) {
	testCases := []updateTestCase{
		{
			name:            "case 0: Cluster without control plane reference",
			clusterManifest: "cluster-without-controlplaneref.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.ControlPlaneReferenceNotSetReason,
				Message:  "Control plane reference is not set for specified Cluster (cluster.x-k8s.io/v1alpha3) object 'org-test/test1'",
			},
		},
		{
			name:            "case 1: Cluster with control plane reference and control plane object not found",
			clusterManifest: "cluster-with-controlplaneref.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.ControlPlaneObjectNotFoundReason,
				Message:  "Control plane object 'org-test/test1-cp-0' of kind Machine is not found for specified Cluster (cluster.x-k8s.io/v1alpha3) object 'org-test/test1'",
			},
		},
		{
			name:                 "case 2: Cluster with control plane reference and control plane object without Ready",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			controlPlaneManifest: "controlplane-without-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capi.WaitingForControlPlaneFallbackReason,
				Message:  "Waiting for control plane object 'org-test/test1-cp-0' of kind Machine to have Ready condition set",
			},
		},
		{
			name:                 "case 3: Cluster with control plane reference and control plane object with Ready(Status=False)",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			controlPlaneManifest: "controlplane-with-ready-false.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Reason:   "Something",
				Severity: capi.ConditionSeverityWarning,
				Message:  "TC control plane is not ready",
			},
		},
		{
			name:                 "case 4: Cluster with control plane reference and control plane object with Ready(Status=True)",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			controlPlaneManifest: "controlplane-with-ready-true.yaml",
			expectedCondition: capi.Condition{
				Type:   capi.ControlPlaneReadyCondition,
				Status: corev1.ConditionTrue,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			t.Log(tc.name)
			ctx := context.Background()
			client := newFakeClient()
			handler, err := newControlPlaneReadyHandler(client)
			if err != nil {
				t.Fatal(err)
			}

			// Create all objects required for the test scenario
			EnsureCRsExist(ctx, t, client, tc)

			// get tested Cluster object
			cluster, err := getTestedCluster(ctx, t, client, tc)
			if err != nil {
				t.Fatal(err)
			}

			// Note: Truncating to match Cluster API behaviour.
			timeBeforeUpdate := metav1.NewTime(time.Now().UTC().Truncate(time.Second))

			// act
			err = handler.update(ctx, cluster)
			if err != nil {
				t.Error(err)
			}

			timeAfterUpdate := metav1.NewTime(time.Now().UTC().Truncate(time.Second))

			// assert
			controlPlaneReady, ok := conditions.GetControlPlaneReady(cluster)
			if ok {
				// First let's check of conditions are equal (all fields except
				// LastTransitionTime, since we cannot know the exact time when
				// the condition will be updated).
				if !internal.AreEqualWithIgnoringLastTransitionTime(&controlPlaneReady, &tc.expectedCondition) {
					t.Logf(
						"ControlPlaneReady was not set correctly, got %s, expected %s",
						internal.SprintComparedCondition(&controlPlaneReady),
						internal.SprintComparedCondition(&tc.expectedCondition))
					t.Fail()
				}

				// Now let's check approximately if the condition's last transition
				// time is within expected time range.
				//
				// Expected order of timestamps:
				//   timeBeforeUpdate <= infrastructureReady.LastTransitionTime <= timeAfterUpdate
				//
				// beforeCondition := timeBeforeUpdate <= infrastructureReady.LastTransitionTime
				// afterCondition := infrastructureReady.LastTransitionTime <= timeAfterUpdate
				lastTransitionTime := controlPlaneReady.LastTransitionTime
				beforeCheck := timeBeforeUpdate.Before(&lastTransitionTime) || timeBeforeUpdate.Equal(&lastTransitionTime)
				afterCheck := lastTransitionTime.Before(&timeAfterUpdate) || lastTransitionTime.Equal(&timeAfterUpdate)

				if !(beforeCheck && afterCheck) {
					t.Logf("InfrastructureReady LastTransitionTime is not correct, expected %s <= %s <= %s", timeBeforeUpdate, lastTransitionTime, timeAfterUpdate)
					t.Fail()
				}
			} else {
				t.Log("ControlPlaneReady was not set")
				t.Fail()
			}
		})
	}
}

func newControlPlaneReadyHandler(client ctrl.Client) (*Handler, error) {
	var err error
	var handler *Handler
	{
		var logger micrologger.Logger
		logger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := HandlerConfig{
			CtrlClient:   client,
			Logger:       logger,
			Name:         "controlPlaneReadyTestHandler",
			UpdateStatus: false,
		}
		handler, err = NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return handler, nil
}

func newFakeClient() ctrl.Client {
	return internal.NewFakeClient(capi.AddToScheme)
}

func EnsureCRsExist(ctx context.Context, t *testing.T, client ctrl.Client, tc updateTestCase) {
	clusterCRPath := filepath.Join("testdata", tc.clusterManifest)
	err := internal.EnsureCRExist(ctx, t, client, clusterCRPath)
	if err != nil {
		t.Fatal(err)
	}

	if tc.controlPlaneManifest == "" {
		return
	}

	controlPlaneCRPath := filepath.Join("testdata", tc.controlPlaneManifest)
	err = internal.EnsureCRExist(ctx, t, client, controlPlaneCRPath)
	if err != nil {
		t.Fatal(err)
	}
}

func getTestedCluster(ctx context.Context, t *testing.T, client ctrl.Client, tc updateTestCase) (*capi.Cluster, error) {
	clusterManifestPath := filepath.Join("testdata", tc.clusterManifest)
	return internal.GetTestedCluster(ctx, t, client, clusterManifestPath)
}
