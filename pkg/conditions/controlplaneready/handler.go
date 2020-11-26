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
	cluster, ok := object.(*capi.Cluster)
	if !ok {
		microerror.Maskf(wrongTypeError, "expected Cluster, got %T", object)
	}

	controlPlaneObject, err := capiexternal.Get(ctx, h.ctrlClient, cluster.Spec.ControlPlaneRef, object.GetNamespace())
	if err != nil {
		return microerror.Mask(err)
	}
	controlPlaneObjectGetter := capiconditions.UnstructuredGetter(controlPlaneObject)

	updateControlPlaneReadyCondition(cluster, controlPlaneObjectGetter)

	// Update deprecated status fields
	cluster.Status.ControlPlaneReady = conditions.IsInfrastructureReadyTrue(cluster)
	if !cluster.Status.ControlPlaneInitialized {
		cluster.Status.ControlPlaneInitialized = cluster.Status.ControlPlaneReady
	}

	return nil
}
