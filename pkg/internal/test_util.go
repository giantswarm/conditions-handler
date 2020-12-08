package internal

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/conditions-handler/pkg/errors"
)

func LoadCR(manifestPath string) (runtime.Object, error) {
	var err error
	var obj runtime.Object

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

func EnsureCRExist(ctx context.Context, t *testing.T, client ctrl.Client, manifestPath string, modifier func(object runtime.Object)) error {
	o, err := LoadCR(manifestPath)
	if err != nil {
		return microerror.Mask(err)
	}
	if modifier != nil {
		modifier(o)
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

	return fake.NewFakeClientWithScheme(scheme)
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
