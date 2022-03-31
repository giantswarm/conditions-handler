package internal

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/conditions-handler/pkg/errors"
)

// AreEqualWithIgnoringLastTransitionTime is checking if two conditions are have
// all fields equal, except LastTransitionTime.
func AreEqualWithIgnoringLastTransitionTime(c1, c2 *capi.Condition) bool {
	dummyLastTransition := metav1.NewTime(time.Now())

	condition1 := *c1
	condition1.LastTransitionTime = dummyLastTransition

	condition2 := *c2
	condition2.LastTransitionTime = dummyLastTransition

	return conditions.AreEqual(&condition1, &condition2)
}

func LoadCR(manifestPath string) (ctrl.Object, error) {
	var err error
	var obj ctrl.Object

	var bs []byte
	{
		bs, err = ioutil.ReadFile(manifestPath)
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
	case "MachinePool":
		obj = new(capiexp.MachinePool)
	case "MockProviderCluster":
		obj = new(MockProviderCluster)
	default:
		return nil, microerror.Maskf(errors.UnknownKindError, "kind: %s", t.Kind)
	}

	// ...and unmarshal the whole object.
	err = yaml.Unmarshal(bs, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return obj, nil
}

func EnsureCRExist(ctx context.Context, t *testing.T, client ctrl.Client, manifestPath string) error {
	o, err := LoadCR(manifestPath)
	if err != nil {
		return microerror.Mask(err)
	}

	err = client.Create(ctx, o)
	if err != nil {
		return microerror.Maskf(errors.InvalidConfigError, "failed to create cluster from input file %s: %#v", manifestPath, err)
	}

	return nil
}

func NewFakeClient(addToSchemeFuncs ...func(s *runtime.Scheme) error) ctrl.Client {
	scheme := runtime.NewScheme()
	for _, f := range addToSchemeFuncs {
		err := f(scheme)
		if err != nil {
			panic(err)
		}
	}

	return fake.NewClientBuilder().WithScheme(scheme).Build()
}

func GetTestedCluster(ctx context.Context, t *testing.T, client ctrl.Client, manifestPath string) (*capi.Cluster, error) {
	o, err := LoadCR(manifestPath)
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
	err = client.Get(ctx, nsName, cluster)
	if err != nil {
		t.Fatalf("err = %#q, want %#v", microerror.JSON(err), nil)
	}

	return cluster, nil
}
