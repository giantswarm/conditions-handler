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
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

type updateTestCase struct {
	name                   string
	clusterManifest        string
	clusterAge             time.Duration
	infrastructureManifest string
	expectedCondition      capi.Condition
}

func TestUpdateInfrastructureReady(t *testing.T) {
	testCases := []updateTestCase{
		{
			name:            "case 0: For 5min old cluster without infrastructure ref, it sets InfrastructureReady(Status=False, Severity=Info, Reason=InfrastructureObjectNotSet)",
			clusterManifest: "cluster-without-infrastructureref.yaml",
			clusterAge:      conditions.WaitingForControlPlaneWarningThresholdTime / 2,
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   conditions.InfrastructureReferenceNotSetReason,
			},
		},
		{
			name:            "case 1: For 20min old cluster without infrastructure ref, it sets InfrastructureReady(Status=False, Severity=Warning, Reason=InfrastructureObjectNotSet)",
			clusterManifest: "cluster-without-infrastructureref.yaml",
			clusterAge:      2 * conditions.WaitingForControlPlaneWarningThresholdTime,
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.InfrastructureReferenceNotSetReason,
			},
		},
		{
			name:            "case 2: For 5min old cluster and nil infrastructure object, it sets InfrastructureReady(Status=False, Severity=Info, Reason=InfrastructureObjectNotFound)",
			clusterManifest: "cluster-with-infrastructureref.yaml",
			clusterAge:      conditions.WaitingForControlPlaneWarningThresholdTime / 2,
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   conditions.InfrastructureObjectNotFoundReason,
			},
		},
		{
			name:            "case 3: For 20min old cluster and nil infrastructure object, it sets InfrastructureReady(Status=False, Severity=Warning, Reason=InfrastructureObjectNotFound)",
			clusterManifest: "cluster-with-infrastructureref.yaml",
			clusterAge:      2 * conditions.WaitingForControlPlaneWarningThresholdTime,
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.InfrastructureObjectNotFoundReason,
			},
		},
		{
			name:                   "case 4: For 5min old Cluster and infrastructure object w/o Ready, it sets InfrastructureReady(Status=False, Severity=Info, Reason=WaitingForInfrastructureFallback)",
			clusterManifest:        "cluster-with-infrastructureref.yaml",
			clusterAge:             conditions.WaitingForControlPlaneWarningThresholdTime / 2,
			infrastructureManifest: "infrastructure-without-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   capi.WaitingForInfrastructureFallbackReason,
			},
		},
		{
			name:                   "case 5: For 20min old Cluster and infrastructure object w/o Ready, it sets InfrastructureReady status to False, Severity=Warning, Reason=WaitingForInfrastructureFallback",
			clusterManifest:        "cluster-with-infrastructureref.yaml",
			clusterAge:             2 * conditions.WaitingForControlPlaneWarningThresholdTime,
			infrastructureManifest: "infrastructure-without-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.InfrastructureReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capi.WaitingForInfrastructureFallbackReason,
			},
		},
		{
			name:                   "case 6: For infrastructure object w/ Ready(Status=False), it sets InfrastructureReady(Status=False)",
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
			name:                   "case 7: For infrastructure object w/ Ready(Status=True), it sets InfrastructureReady(Status=True)",
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

			// act
			err = handler.update(ctx, &clusterWrapper{cluster})
			if err != nil {
				t.Fatal(err)
			}

			// assert
			infrastructureReady, ok := conditions.GetInfrastructureReady(cluster)
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

func EnsureCRsExist(ctx context.Context, t *testing.T, client ctrl.Client, tc updateTestCase) {
	clusterCRPath := filepath.Join("testdata", tc.clusterManifest)
	err := internal.EnsureCRExist(ctx, t, client, clusterCRPath, func(o runtime.Object) {
		cluster, ok := o.(*capi.Cluster)
		if !ok {
			t.Fatalf("couldn't cast object %T to Cluster", o)
		}
		cluster.CreationTimestamp = metav1.NewTime(time.Now().Add(-tc.clusterAge))
	})
	if err != nil {
		t.Fatal(err)
	}

	if tc.infrastructureManifest == "" {
		return
	}

	infrastructureManifestPath := filepath.Join("testdata", tc.infrastructureManifest)
	err = internal.EnsureCRExist(ctx, t, client, infrastructureManifestPath, nil)
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
