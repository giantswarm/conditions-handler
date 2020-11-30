package internal

import (
	"context"
	"fmt"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
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
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
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
		return
	}

	if conditions.IsUnsupported(object, h.conditionType) {
		return microerror.Maskf(
			conditions.UnsupportedConditionStatusError,
			conditions.UnsupportedConditionStatusErrorMessage(object, h.conditionType))
	}

	initialConditionValue := capiconditions.Get(object, h.conditionType)
	h.logDebug(ctx, "ensuring condition %s", sprintCondition(h.conditionType, initialConditionValue))
	var conditionChanged bool

	defer func() {
		if err == nil {
			if conditionChanged {
				h.logDebug(ctx, "ensured condition %s", sprintCondition(h.conditionType, initialConditionValue))
			} else {
				h.logDebug(ctx, "ensured condition %s, no change", h.conditionType)
			}
		} else {
			h.logWarning(ctx, err, "an error occurred while ensuring condition %s", h.conditionType)
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
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return
}

func (h *Handler) EnsureDeleted(ctx context.Context, object conditions.Object) (err error) {
	if h.ensureDeletedFunc == nil {
		return
	}

	return nil
}

func (h *Handler) logDebug(ctx context.Context, message string, messageArgs ...interface{}) {
	h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf(message, messageArgs...))
}

func (h *Handler) logWarning(ctx context.Context, err error, message string, messageArgs ...interface{}) {
	h.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf(message, messageArgs...), "error", err)
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
