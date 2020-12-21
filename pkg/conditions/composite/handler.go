package composite

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/errors"
	"github.com/giantswarm/conditions-handler/pkg/handler"
	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
)

type HandlerConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger

	Name     string
	Handlers []handler.Interface
}

type Handler struct {
	ctrlClient ctrl.Client
	logger     micrologger.Logger
	name       string
	handlers   []handler.Interface
}

func NewHandler(config HandlerConfig) (*Handler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(errors.InvalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(errors.InvalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Name == "" {
		return nil, microerror.Maskf(errors.InvalidConfigError, "%T.Name must not be empty", config)
	}
	if config.Handlers == nil {
		return nil, microerror.Maskf(errors.InvalidConfigError, "%T.Handlers must not be empty", config)
	}

	h := &Handler{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
		name:       config.Name,
		handlers:   config.Handlers,
	}

	return h, nil
}

func (h *Handler) EnsureCreated(ctx context.Context, object interface{}) error {
	var err error

	err = internal.LogObjectJson(ctx, h.logger, object)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, handler := range h.handlers {
		err = handler.EnsureCreated(ctx, object)
		if err != nil {
			return microerror.Mask(err)
		}

		objWithConditions, err := key.ToObjectWithConditions(object)
		if err != nil {
			return microerror.Mask(err)
		}

		err = internal.LogObjectJson(ctx, h.logger, objWithConditions.GetConditions())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (h *Handler) EnsureDeleted(ctx context.Context, object interface{}) error {
	var err error
	for _, handler := range h.handlers {
		err = handler.EnsureDeleted(ctx, object)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (h *Handler) Name() string {
	return h.name
}
