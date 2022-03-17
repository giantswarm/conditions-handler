package factory

import (
	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/conditions-handler/pkg/conditions/composite"
	"github.com/giantswarm/conditions-handler/pkg/conditions/controlplaneready"
	"github.com/giantswarm/conditions-handler/pkg/conditions/creating"
	"github.com/giantswarm/conditions-handler/pkg/conditions/infrastructureready"
	"github.com/giantswarm/conditions-handler/pkg/conditions/nodepoolsready"
	"github.com/giantswarm/conditions-handler/pkg/conditions/summary"
	"github.com/giantswarm/conditions-handler/pkg/conditions/upgrading"
	"github.com/giantswarm/conditions-handler/pkg/handler"
)

// NewClusterConditionsHandler creates a composite handler for reconciling
// MachinePool conditions, which consists of condition handlers for
// InfrastructureReady, ControlPlaneReady, NodePoolsReady, Ready, Creating and
// Upgrading conditions.
func NewClusterConditionsHandler(config handler.Config) (*composite.Handler, error) {
	var err error

	var infrastructureReadyHandler *infrastructureready.Handler
	{
		c := infrastructureready.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "clusterInfrastructureReadyHandler",
			UpdateStatus: false,
		}
		infrastructureReadyHandler, err = infrastructureready.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var controlPlaneReadyHandler *controlplaneready.Handler
	{
		c := controlplaneready.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "clusterControlPlaneReadyHandler",
			UpdateStatus: false,
		}
		controlPlaneReadyHandler, err = controlplaneready.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodePoolsReadyHandler *nodepoolsready.Handler
	{
		c := nodepoolsready.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "clusterNodePoolsReadyHandler",
			UpdateStatus: false,
		}
		nodePoolsReadyHandler, err = nodepoolsready.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var readyHandler *summary.Handler
	{
		c := summary.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			UpdateStatus: false,
			ConditionsToSummarize: []capi.ConditionType{
				capi.InfrastructureReadyCondition,
				capi.ControlPlaneReadyCondition,
				conditions.NodePoolsReady,
			},
			IgnoreOptions: []conditions.CheckOption{
				ignoreNodePoolsNotFoundInfo(),
			},
			Name: "clusterReadyHandler",
		}

		readyHandler, err = summary.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var creatingHandler *creating.Handler
	{
		c := creating.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "clusterCreatingConditionHandler",
			UpdateStatus: false,
		}

		creatingHandler, err = creating.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var upgradingHandler *upgrading.Handler
	{
		c := upgrading.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "clusterUpgradingConditionHandler",
			UpdateStatus: true,
		}

		upgradingHandler, err = upgrading.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConditionsHandler *composite.Handler
	{
		c := composite.HandlerConfig{
			CtrlClient: config.CtrlClient,
			Logger:     config.Logger,
			Name:       config.Name,
			Handlers: []handler.Interface{
				infrastructureReadyHandler,
				controlPlaneReadyHandler,
				nodePoolsReadyHandler,
				readyHandler,
				creatingHandler,
				upgradingHandler,
			},
		}

		clusterConditionsHandler, err = composite.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return clusterConditionsHandler, nil
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
