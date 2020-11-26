package infrastructureready

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexternal "sigs.k8s.io/cluster-api/controllers/external"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
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
	}

	internalHandler, err := internal.NewHandler(internalHandlerConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	h.internalHandler = internalHandler

	return h, nil
}

func (h *Handler) EnsureCreated(ctx context.Context, object interface{}) error {
	obj, err := key.ToObjectWithConditions(object)
	if err != nil {
		return microerror.Mask(err)
	}

	return h.internalHandler.EnsureCreated(ctx, obj)
}

func (h *Handler) EnsureDeleted(_ context.Context, _ interface{}) error {
	return nil
}

func (h *Handler) ensureCreated(ctx context.Context, object conditions.Object) error {
	obj, err := toObjectWithInfrastructure(object)
	if err != nil {
		return microerror.Mask(err)
	}

	infrastructureObject, err := h.getInfrastructureObject(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	update(obj, infrastructureObject)
	return nil
}

func (h *Handler) getInfrastructureObject(ctx context.Context, object objectWithInfrastructureRef) (capiconditions.Getter, error) {
	infrastructureRef := object.GetInfrastructureRef()
	if infrastructureRef == nil {
		// Infrastructure object is not set
		return nil, nil
	}

	infrastructureObject, err := capiexternal.Get(ctx, h.ctrlClient, object.GetInfrastructureRef(), object.GetNamespace())
	if apierrors.IsNotFound(err) {
		// Infrastructure object is not found, here we don't care why
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// Wrap unstructured object into a capiconditions.Getter
	infrastructureObjectGetter := capiconditions.UnstructuredGetter(infrastructureObject)

	return infrastructureObjectGetter, nil
}

func toObjectWithInfrastructure(object conditions.Object) (objectWithInfrastructureRef, error) {
	if object == nil {
		return nil, microerror.Maskf(wrongTypeError, "expected non-nil conditions.Object, got nil '%T'", object)
	}

	clusterPointer, ok := object.(*capi.Cluster)
	if ok {
		return &clusterWrapper{clusterPointer}, nil
	}

	machinePoolPointer, ok := object.(*capiexp.MachinePool)
	if ok {
		return &machinePoolWrapper{machinePoolPointer}, nil
	}

	return nil, microerror.Maskf(wrongTypeError, "expected Cluster or MachinePool, got %T", object)
}
