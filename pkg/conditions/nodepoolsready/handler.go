package nodepoolsready

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
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
		ConditionType:                 conditions.NodePoolsReady,
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

	nodePools, err := h.getNodePools(ctx, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	update(&cluster, nodePools)
	return nil
}

func (h *Handler) getNodePools(ctx context.Context, cluster *capi.Cluster) ([]capiconditions.Getter, error) {
	machinePools, err := internal.ListMachinePoolsByMetadata(ctx, h.ctrlClient, cluster.ObjectMeta)
	if apierrors.IsNotFound(err) || len(machinePools.Items) == 0 {
		// set
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// We need a slice of Getter objects for SetAggregate, so we do a bit of
	// boxing/casting here.
	var machinePoolPointers []capiconditions.Getter
	for _, machinePool := range machinePools.Items {
		machinePoolObj := machinePool
		machinePoolPointers = append(machinePoolPointers, &machinePoolObj)
	}

	return machinePoolPointers, nil
}
