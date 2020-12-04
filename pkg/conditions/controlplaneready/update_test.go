package controlplaneready

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

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
				Type:   capi.ControlPlaneReadyCondition,
				Status: corev1.ConditionFalse,
				Reason: "Something",
				Severity: capi.ConditionSeverityWarning,
				Message: "TC control plane is not ready",
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
			client := newFakeClient()
			handler, err := newControlPlaneReadyHandler(client)
			if err != nil {
				t.Fatal(err)
			}

			// Create all objects required for the test scenario
			ensureCRsExist(t, client, tc)

			// get tested Cluster object
			cluster, err := getTestedCluster(t, client, tc)
			if err != nil {
				t.Fatal(err)
			}

			// act
			err = handler.update(context.Background(), cluster)
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
	scheme := runtime.NewScheme()

	err := capi.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	return fake.NewFakeClientWithScheme(scheme)
}

func loadCR(fName string) (runtime.Object, error) {
	var err error
	var obj runtime.Object

	var bs []byte
	{
		bs, err = ioutil.ReadFile(filepath.Join("testdata", fName))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// First parse kind.
	t := &metav1.TypeMeta{}
	err = yaml.Unmarshal(bs, t)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Then construct correct CR object.
	switch t.Kind {
	case "Cluster":
		obj = new(capi.Cluster)
	case "Machine":
		obj = new(capi.Machine)
	default:
		return nil, microerror.Maskf(unknownKindError, "kind: %s", t.Kind)
	}

	// ...and unmarshal the whole object.
	err = yaml.Unmarshal(bs, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return obj, nil
}

func ensureCRsExist(t *testing.T, client ctrl.Client, tc updateTestCase) {
	// input Cluster object
	o, err := loadCR(tc.clusterManifest)
	if err != nil {
		t.Fatal(err)
	}
	cluster, ok := o.(*capi.Cluster)
	if !ok {
		t.Fatalf("couldn't cast object %T to Cluster", o)
	}
	cluster.CreationTimestamp = metav1.NewTime(time.Now().Add(-tc.clusterAge))

	err = client.Create(context.Background(), cluster)
	if err != nil {
		t.Fatalf("failed to create cluster from input file %s: %#v", tc.clusterManifest, err)
	}

	// input ControlPlane object
	if tc.controlPlaneManifest == "" {
		return
	}

	// input ControlPlane object
	o, err = loadCR(tc.controlPlaneManifest)
	if err != nil {
		t.Fatal(err)
	}
	controlPlane, ok := o.(*capi.Machine)
	if !ok {
		t.Fatalf("couldn't cast object %T to Machine", o)
	}

	err = client.Create(context.Background(), controlPlane)
	if err != nil {
		t.Fatalf("failed to create cluster from input file %s: %#v", tc.controlPlaneManifest, err)
	}
}

func getTestedCluster(t *testing.T, client ctrl.Client, tc updateTestCase) (*capi.Cluster, error) {
	// input Cluster object
	o, err := loadCR(tc.clusterManifest)
	if err != nil {
		t.Fatal(err)
	}
	inputCluster, ok := o.(*capi.Cluster)
	if !ok {
		t.Fatalf("couldn't cast object %T to Cluster", o)
	}

	nsName := types.NamespacedName{
		Namespace: inputCluster.Namespace,
		Name:      inputCluster.Name,
	}

	cluster := &capi.Cluster{}
	err = client.Get(context.Background(), nsName, cluster)
	if err != nil {
		t.Fatalf("err = %#q, want %#v", microerror.JSON(err), nil)
	}

	return cluster, nil
}
