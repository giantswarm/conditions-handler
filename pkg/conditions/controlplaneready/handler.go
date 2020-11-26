package controlplaneready

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexternal "sigs.k8s.io/cluster-api/controllers/external"
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
		ConditionType:                 capi.ControlPlaneReadyCondition,
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
	cluster, err := key.ToCluster(object)
	if err != nil {
		return microerror.Mask(err)
	}

	return h.internalHandler.EnsureCreated(ctx, &cluster)
}

func (h *Handler) EnsureDeleted(_ context.Context, _ interface{}) error {
	return nil
}

func (h *Handler) ensureCreated(ctx context.Context, object conditions.Object) error {
	cluster, err := key.ToCluster(object)
	if err != nil {
		return microerror.Mask(err)
	}

	controlPlane, err := h.getControlPlaneObject(ctx, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	update(&cluster, controlPlane)
	return nil
}

func (h *Handler) getControlPlaneObject(ctx context.Context, cluster *capi.Cluster) (capiconditions.Getter, error) {
	controlPlaneObject, err := capiexternal.Get(ctx, h.ctrlClient, cluster.Spec.ControlPlaneRef, cluster.Namespace)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	controlPlaneObjectGetter := capiconditions.UnstructuredGetter(controlPlaneObject)
	return controlPlaneObjectGetter, nil
}
