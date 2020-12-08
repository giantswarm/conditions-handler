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
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

type updateTestCase struct {
	name                 string
	clusterManifest      string
	clusterAge           time.Duration
	controlPlaneManifest string
	expectedCondition    capi.Condition
}

func TestUpdateControlPlaneReady(t *testing.T) {
	testCases := []updateTestCase{
		{
			name:            "case 0: For For 5min old Cluster without control plane reference, it sets ControlPlaneReady status to False, Severity=Info, Reason=ControlPlaneReferenceNotSet",
			clusterManifest: "cluster-without-controlplaneref.yaml",
			clusterAge:      conditions.WaitingForControlPlaneWarningThresholdTime / 2,
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   conditions.ControlPlaneReferenceNotSetReason,
			},
		},
		{
			name:            "case 1: For For 20min old Cluster without control plane reference, it sets ControlPlaneReady status to False, Severity=Warning, Reason=ControlPlaneReferenceNotSet",
			clusterManifest: "cluster-without-controlplaneref.yaml",
			clusterAge:      2 * conditions.WaitingForControlPlaneWarningThresholdTime,
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.ControlPlaneReferenceNotSetReason,
			},
		},
		{
			name:            "case 2: For 5min old Cluster with control plane reference and control plane object not found, it sets ControlPlaneReady status to False, Severity=Info, Reason=ControlPlaneObjectNotFound",
			clusterManifest: "cluster-with-controlplaneref.yaml",
			clusterAge:      conditions.WaitingForControlPlaneWarningThresholdTime / 2,
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   conditions.ControlPlaneObjectNotFoundReason,
			},
		},
		{
			name:            "case 3: For 20min old Cluster with control plane reference and control plane object not found, it sets ControlPlaneReady status to False, Severity=Warning, Reason=ControlPlaneObjectNotFound",
			clusterManifest: "cluster-with-controlplaneref.yaml",
			clusterAge:      2 * conditions.WaitingForControlPlaneWarningThresholdTime,
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   conditions.ControlPlaneObjectNotFoundReason,
			},
		},
		{
			name:                 "case 4: For 5min old Cluster and control plane object w/o Ready, it sets ControlPlaneReady status to False, Severity=Info, Reason=WaitingForControlPlaneFallback",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			clusterAge:           conditions.WaitingForControlPlaneWarningThresholdTime / 2,
			controlPlaneManifest: "controlplane-without-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityInfo,
				Reason:   capi.WaitingForControlPlaneFallbackReason,
			},
		},
		{
			name:                 "case 5: For 20min old Cluster and control plane object w/o Ready, it sets ControlPlaneReady status to False, Severity=Warning, Reason=WaitingForControlPlaneFallback",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			clusterAge:           2 * conditions.WaitingForControlPlaneWarningThresholdTime,
			controlPlaneManifest: "controlplane-without-ready.yaml",
			expectedCondition: capi.Condition{
				Type:     capi.ControlPlaneReadyCondition,
				Status:   corev1.ConditionFalse,
				Severity: capi.ConditionSeverityWarning,
				Reason:   capi.WaitingForControlPlaneFallbackReason,
			},
		},
		{
			name:                 "case 6: For control plane object w/ Ready(Status=False), it sets ControlPlaneReady(Status=False)",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			clusterAge:           2 * conditions.WaitingForControlPlaneWarningThresholdTime,
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
			name:                 "case 7: For control plane object w/ Ready(Status=True), it sets ControlPlaneReady(Status=True)",
			clusterManifest:      "cluster-with-controlplaneref.yaml",
			clusterAge:           2 * conditions.WaitingForControlPlaneWarningThresholdTime,
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

			// act
			err = handler.update(ctx, cluster)
			if err != nil {
				t.Error(err)
			}

			// assert
			controlPlaneReady, ok := conditions.GetControlPlaneReady(cluster)
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

	if tc.controlPlaneManifest == "" {
		return
	}

	controlPlaneCRPath := filepath.Join("testdata", tc.controlPlaneManifest)
	err = internal.EnsureCRExist(ctx, t, client, controlPlaneCRPath, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func getTestedCluster(ctx context.Context, t *testing.T, client ctrl.Client, tc updateTestCase) (*capi.Cluster, error) {
	clusterManifestPath := filepath.Join("testdata", tc.clusterManifest)
	return internal.GetTestedCluster(ctx, t, client, clusterManifestPath)
}
