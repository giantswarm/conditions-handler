package infrastructureready

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/errors"
	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
)

type HandlerConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger

	Name         string
	UpdateStatus bool
}

type Handler struct {
	ctrlClient      ctrl.Client
	internalHandler *internal.Handler
	logger          micrologger.Logger
	name            string
}

func NewHandler(config HandlerConfig) (*Handler, error) {
	h := &Handler{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		name:       config.Name,
	}

	internalHandlerConfig := internal.HandlerConfig{
		CtrlClient:        config.CtrlClient,
		Logger:            config.Logger,
		UpdateStatus:      config.UpdateStatus,
		ConditionType:     capi.InfrastructureReadyCondition,
		EnsureCreatedFunc: h.ensureCreated,
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

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) ensureCreated(ctx context.Context, object conditions.Object) error {
	obj, err := toObjectWithInfrastructure(object)
	if err != nil {
		return microerror.Mask(err)
	}

	err = h.update(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func toObjectWithInfrastructure(object conditions.Object) (objectWithInfrastructureRef, error) {
	if object == nil {
		return nil, microerror.Maskf(errors.WrongTypeError, "expected non-nil conditions.Object, got nil '%T'", object)
	}

	clusterPointer, ok := object.(*capi.Cluster)
	if ok {
		return &clusterWrapper{clusterPointer}, nil
	}

	machinePoolPointer, ok := object.(*capiexp.MachinePool)
	if ok {
		return &machinePoolWrapper{machinePoolPointer}, nil
	}

	return nil, microerror.Maskf(errors.WrongTypeError, "expected Cluster or MachinePool, got %T", object)
}
