package infrastructureready

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexternal "sigs.k8s.io/cluster-api/controllers/external"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
)

type HandlerConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger

	UpdateStatusOnConditionChange bool
}

type Handler struct {
	ctrlClient      ctrl.Client
	internalHandler *internal.Handler
	logger          micrologger.Logger
}

func NewHandler(config HandlerConfig) (*Handler, error) {
	h := &Handler{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	internalHandlerConfig := internal.HandlerConfig{
		CtrlClient:                    config.CtrlClient,
		Logger:                        config.Logger,
		UpdateStatusOnConditionChange: config.UpdateStatusOnConditionChange,
		ConditionType:                 capi.InfrastructureReadyCondition,
		EnsureCreatedFunc:             h.ensureCreated,
		EnsureDeletedFunc:             nil,
	}

	internalHandler, err := internal.NewHandler(internalHandlerConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	h.internalHandler = internalHandler

	return h, nil
}

func (h *Handler) EnsureCreated(ctx context.Context, object conditions.Object) error {
	return h.internalHandler.EnsureCreated(ctx, object)
}

func (h *Handler) EnsureDeleted(ctx context.Context, object conditions.Object) error {
	return h.internalHandler.EnsureDeleted(ctx, object)
}

func (h *Handler) ensureCreated(ctx context.Context, object conditions.Object) error {
	infrastructureRef, err := getInfrastructureRef(object)
	if err != nil {
		return microerror.Mask(err)
	}

	infrastructureObject, err := capiexternal.Get(ctx, h.ctrlClient, infrastructureRef, object.GetNamespace())
	if err != nil {
		return microerror.Mask(err)
	}
	infrastructureObjectGetter := capiconditions.UnstructuredGetter(infrastructureObject)

	updateInfrastructureReadyCondition(object, infrastructureObjectGetter)
	err = updateObjectStatusInfrastructureReady(object)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func getInfrastructureRef(object conditions.Object) (*corev1.ObjectReference, error) {
	clusterPointer, ok := object.(*capi.Cluster)
	if ok {
		return clusterPointer.Spec.InfrastructureRef, nil
	}

	machinePoolPointer, ok := object.(*capiexp.MachinePool)
	if ok {
		return &machinePoolPointer.Spec.Template.Spec.InfrastructureRef, nil
	}

	return nil, microerror.Maskf(wrongTypeError, "expected Cluster or MachinePool, got %T", object)
}

func updateObjectStatusInfrastructureReady(object conditions.Object) error {
	clusterPointer, ok := object.(*capi.Cluster)
	if ok {
		clusterPointer.Status.InfrastructureReady = conditions.IsInfrastructureReadyTrue(object)
		return nil
	}

	machinePoolPointer, ok := object.(*capiexp.MachinePool)
	if ok {
		machinePoolPointer.Status.InfrastructureReady = conditions.IsInfrastructureReadyTrue(object)
	}

	return microerror.Maskf(wrongTypeError, "expected Cluster or MachinePool, got %T", object)
}
