package internal

import (
	"context"
	"fmt"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/errors"
)

type HandlerConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger

	UpdateStatus      bool
	ConditionType     capi.ConditionType
	EnsureCreatedFunc func(ctx context.Context, object conditions.Object) error
	EnsureDeletedFunc func(ctx context.Context, object conditions.Object) error
}

type Handler struct {
	ctrlClient ctrl.Client
	logger     micrologger.Logger

	conditionType     capi.ConditionType
	updateStatus      bool
	ensureCreatedFunc func(ctx context.Context, object conditions.Object) error
	ensureDeletedFunc func(ctx context.Context, object conditions.Object) error
}

func NewHandler(config HandlerConfig) (*Handler, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(errors.InvalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(errors.InvalidConfigError, "%T.Logger must not be empty", config)
	}

	h := &Handler{
		ctrlClient:        config.CtrlClient,
		logger:            config.Logger,
		conditionType:     config.ConditionType,
		updateStatus:      config.UpdateStatus,
		ensureCreatedFunc: config.EnsureCreatedFunc,
		ensureDeletedFunc: config.EnsureDeletedFunc,
	}

	return h, nil
}

func (h *Handler) EnsureCreated(ctx context.Context, object conditions.Object) (err error) {
	if h.ensureCreatedFunc == nil {
		return nil
	}

	if conditions.IsUnsupported(object, h.conditionType) {
		return microerror.Maskf(
			conditions.UnsupportedConditionStatusError,
			conditions.UnsupportedConditionStatusErrorMessage(object, h.conditionType))
	}

	initialConditionValue := capiconditions.Get(object, h.conditionType)
	h.logger.Debugf(ctx, "ensuring condition %s", sprintCondition(h.conditionType, initialConditionValue))
	var conditionChanged bool

	defer func() {
		if err == nil {
			if conditionChanged {
				h.logger.Debugf(ctx, "ensured condition %s", sprintCondition(h.conditionType, initialConditionValue))
			} else {
				h.logger.Debugf(ctx, "ensured condition %s, no change", h.conditionType)
			}
		} else {
			h.logger.Errorf(ctx, err, "an error occurred while ensuring condition %s", h.conditionType)
		}
	}()

	err = h.ensureCreatedFunc(ctx, object)
	if err != nil {
		return microerror.Mask(err)
	}

	currentConditionValue := capiconditions.Get(object, h.conditionType)
	conditionChanged = !conditions.AreEqual(initialConditionValue, currentConditionValue)

	if h.updateStatus {
		err = h.ctrlClient.Status().Update(ctx, object)
		if apierrors.IsConflict(err) {
			h.logger.Debugf(ctx, "conflict trying to save object in k8s API concurrently", "stack", microerror.JSON(microerror.Mask(err)))
			h.logger.Debugf(ctx, "cancelling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (h *Handler) EnsureDeleted(_ context.Context, _ conditions.Object) (err error) {
	if h.ensureDeletedFunc == nil {
		return
	}

	return nil
}

func sprintCondition(conditionType capi.ConditionType, condition *capi.Condition) string {
	var text string
	if condition != nil {
		text = fmt.Sprintf(
			"%s(Status=%q, Reason=%q, Severity=%q, Message=%q)",
			condition.Type,
			condition.Status,
			condition.Reason,
			condition.Severity,
			condition.Message)
	} else {
		text = fmt.Sprintf("%s(not set)", conditionType)
	}

	return text
}
