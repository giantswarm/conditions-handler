package creating

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
)

type HandlerConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger

	Name                          string
	UpdateStatusOnConditionChange bool
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
		CtrlClient:                    config.CtrlClient,
		Logger:                        config.Logger,
		UpdateStatusOnConditionChange: config.UpdateStatusOnConditionChange,
		ConditionType:                 conditions.Creating,
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

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) ensureCreated(_ context.Context, object conditions.Object) error {
	update(object)
	return nil
}
