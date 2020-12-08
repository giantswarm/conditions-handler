package internal

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/errors"
)

func ListMachinePoolsByMetadata(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*capiexp.MachinePoolList, error) {
	if obj.Labels[capi.ClusterLabelName] == "" {
		err := microerror.Maskf(errors.InvalidConfigError, "Label %q must not be empty for object %q", capi.ClusterLabelName, obj.GetSelfLink())
		return nil, err
	}

	machinePools, err := ListMachinePoolsByClusterID(ctx, c, obj.Namespace, obj.Labels[capi.ClusterLabelName])
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return machinePools, nil
}

func ListMachinePoolsByClusterID(ctx context.Context, c client.Client, clusterNamespace, clusterID string) (*capiexp.MachinePoolList, error) {
	machinePools := &capiexp.MachinePoolList{}
	err := c.List(ctx, machinePools, client.MatchingLabels{capi.ClusterLabelName: clusterID}, client.InNamespace(clusterNamespace))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return machinePools, nil
}
