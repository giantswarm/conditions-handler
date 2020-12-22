package infrastructureready

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
	name                   string
	clusterManifest        string
	infrastructureManifest string
	expectedCondition      capi.Condition
}

func TestUpdateInfrastructureReady(t *testing.T) {
	testCases := []updateTestCase{
		{
			name:            "case 0: Cluster without infrastructure ref",
			clusterManifest: "cluster-without-infrastructureref.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.InfrastructureReferenceNotSetReason,
				Message:  "Cluster (cluster.x-k8s.io/v1alpha3) object 'org-test/test1' does not have infrastructure reference set",
			},
		},
		{
			name:            "case 1: Cluster with infrastructure ref and infrastructure object not found",
			clusterManifest: "cluster-with-infrastructureref.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.InfrastructureObjectNotFoundReason,
				Message:  "Corresponding provider-specific infrastructure object 'org-test/test1' is not found for Cluster (cluster.x-k8s.io/v1alpha3) object 'org-test/test1'",
			},
		},
		{
			name:                   "case 2: Cluster with infrastructure ref and infrastructure object without Ready",
			clusterManifest:        "cluster-with-infrastructureref.yaml",
			infrastructureManifest: "infrastructure-without-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capi.WaitingForInfrastructureFallbackReason,
				Message:  "Waiting for infrastructure object 'org-test/test1' of kind MockProviderCluster to have Ready condition set",
			},
		},
		{
			name:                   "case 3: Cluster with infrastructure ref and infrastructure object with Ready(Status=False)",
			clusterManifest:        "cluster-with-infrastructureref.yaml",
			infrastructureManifest: "infrastructure-with-ready-false.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Reason:   "Something",
				Severity: capi.ConditionSeverityWarning,
				Message:  "Infrastructure is not ready",
			},
		},
		{
			name:                   "case 4: Cluster with infrastructure ref and infrastructure object with Ready(Status=True)",
			clusterManifest:        "cluster-with-infrastructureref.yaml",
			infrastructureManifest: "infrastructure-with-ready-true.yaml",
			expectedCondition: capi.Condition{
				Type:   capi.InfrastructureReadyCondition,
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
			handler, err := newInfrastructureReadyHandler(client)
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
			err = handler.update(ctx, &clusterWrapper{cluster})
			if err != nil {
				t.Fatal(err)
			}

			timeAfterUpdate := metav1.NewTime(time.Now().UTC().Truncate(time.Second))

			// assert
			infrastructureReady, ok := conditions.GetInfrastructureReady(cluster)
			if ok {
				// First let's check of conditions are equal (all fields except
				// LastTransitionTime, since we cannot know the exact time when
				// the condition will be updated).
				if !internal.AreEqualWithIgnoringLastTransitionTime(&infrastructureReady, &tc.expectedCondition) {
					t.Logf(
						"InfrastructureReady was not set correctly, got %s, expected %s",
						internal.SprintComparedCondition(&infrastructureReady),
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
				lastTransitionTime := infrastructureReady.LastTransitionTime
				beforeCheck := timeBeforeUpdate.Before(&lastTransitionTime) || timeBeforeUpdate.Equal(&lastTransitionTime)
				afterCheck := lastTransitionTime.Before(&timeAfterUpdate) || lastTransitionTime.Equal(&timeAfterUpdate)

				if !(beforeCheck && afterCheck) {
					t.Logf("InfrastructureReady LastTransitionTime is not correct, expected %s <= %s <= %s", timeBeforeUpdate, lastTransitionTime, timeAfterUpdate)
					t.Fail()
				}
			} else {
				t.Log("InfrastructureReady was not set")
				t.Fail()
			}
		})
	}
}

func EnsureCRsExist(ctx context.Context, t *testing.T, client ctrl.Client, tc updateTestCase) {
	clusterCRPath := filepath.Join("testdata", tc.clusterManifest)
	err := internal.EnsureCRExist(ctx, t, client, clusterCRPath)
	if err != nil {
		t.Fatal(err)
	}

	if tc.infrastructureManifest == "" {
		return
	}

	infrastructureManifestPath := filepath.Join("testdata", tc.infrastructureManifest)
	err = internal.EnsureCRExist(ctx, t, client, infrastructureManifestPath)
	if err != nil {
		t.Fatal(err)
	}
}

func getTestedCluster(ctx context.Context, t *testing.T, client ctrl.Client, tc updateTestCase) (*capi.Cluster, error) {
	clusterManifestPath := filepath.Join("testdata", tc.clusterManifest)
	return internal.GetTestedCluster(ctx, t, client, clusterManifestPath)
}

func newInfrastructureReadyHandler(client ctrl.Client) (*Handler, error) {
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
			Name:         "infrastructureReadyTestHandler",
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
	return internal.NewFakeClient(capi.AddToScheme, internal.AddMockToScheme)
}
