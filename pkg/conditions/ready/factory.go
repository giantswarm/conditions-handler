package ready

import (
	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/conditions-handler/pkg/conditions/handler"
)

func NewClusterReadyHandler(config handler.Config) (*Handler, error) {
	readyHandlerConfig := HandlerConfig{
		CtrlClient:                    config.CtrlClient,
		Logger:                        config.Logger,
		UpdateStatusOnConditionChange: true,
		ConditionsToSummarize: []capi.ConditionType{
			capi.InfrastructureReadyCondition,
			capi.ControlPlaneReadyCondition,
			conditions.NodePoolsReady,
		},
		IgnoreOptions: []conditions.CheckOption{
			ignoreNodePoolsNotFoundInfo(),
		},
	}

	readyHandler, err := NewHandler(readyHandlerConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return readyHandler, nil
}

func NewMachinePoolReadyHandler(config handler.Config) (*Handler, error) {
	readyHandlerConfig := HandlerConfig{
		CtrlClient:                    config.CtrlClient,
		Logger:                        config.Logger,
		UpdateStatusOnConditionChange: true,
		ConditionsToSummarize: []capi.ConditionType{
			capi.InfrastructureReadyCondition,
		},
	}

	readyHandler, err := NewHandler(readyHandlerConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return readyHandler, nil
}

func ignoreNodePoolsNotFoundInfo() conditions.CheckOption {
	return func(condition *capi.Condition) bool {
		return condition != nil &&
			condition.Type == conditions.NodePoolsReady &&
			condition.Status == corev1.ConditionFalse &&
			condition.Reason == conditions.NodePoolsNotFoundReason &&
			condition.Severity == capi.ConditionSeverityInfo
	}
}
