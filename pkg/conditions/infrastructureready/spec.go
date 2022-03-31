package infrastructureready

import (
	"github.com/giantswarm/conditions/pkg/conditions"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
)

type objectWithInfrastructureRef interface {
	conditions.Object
	GetInfrastructureRef() *corev1.ObjectReference
	SetStatusInfrastructureReady(ready bool)
}

type clusterWrapper struct {
	*capi.Cluster
}

func (c *clusterWrapper) GetInfrastructureRef() *corev1.ObjectReference {
	return c.Spec.InfrastructureRef
}

func (c *clusterWrapper) SetStatusInfrastructureReady(value bool) {
	c.Status.InfrastructureReady = value
}

type machinePoolWrapper struct {
	*capiexp.MachinePool
}

func (mp *machinePoolWrapper) GetInfrastructureRef() *corev1.ObjectReference {
	return &mp.Spec.Template.Spec.InfrastructureRef
}

func (mp *machinePoolWrapper) SetStatusInfrastructureReady(value bool) {
	mp.Status.InfrastructureReady = value
}
