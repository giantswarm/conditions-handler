package factory

import (
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"

	"github.com/giantswarm/conditions-handler/pkg/conditions/composite"
	"github.com/giantswarm/conditions-handler/pkg/conditions/creating"
	"github.com/giantswarm/conditions-handler/pkg/conditions/infrastructureready"
	"github.com/giantswarm/conditions-handler/pkg/conditions/replicasready"
	"github.com/giantswarm/conditions-handler/pkg/conditions/summary"
	"github.com/giantswarm/conditions-handler/pkg/conditions/upgrading"
	"github.com/giantswarm/conditions-handler/pkg/handler"
)

// NewMachinePoolConditionsHandler creates a composite handler for reconciling
// MachinePool conditions, which consists of condition handlers for
// InfrastructureReady, Ready, Creating and Upgrading conditions.
func NewMachinePoolConditionsHandler(config handler.Config) (*composite.Handler, error) {
	var err error

	var infrastructureReadyHandler *infrastructureready.Handler
	{
		c := infrastructureready.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "machinePoolInfrastructureReadyHandler",
			UpdateStatus: false,
		}
		infrastructureReadyHandler, err = infrastructureready.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var replicasReadyHandler *replicasready.Handler
	{
		c := replicasready.HandlerConfig{
			CtrlClient:   config.CtrlClient,
			Logger:       config.Logger,
			Name:         "machinePoolReplicasReadyHandler",
			UpdateStatus: false,
		}
		replicasReadyHandler, err = replicasready.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var readyHandler *summary.Handler
	{
		c := summary.HandlerConfig{
			CtrlClient:           config.CtrlClient,
			Logger:               config.Logger,
			UpdateStatus:         false,
			SummaryConditionType: capi.ReadyCondition,
			ConditionsToSummarize: []capi.ConditionType{
				capi.InfrastructureReadyCondition,
				capiexp.ReplicasReadyCondition,
			},
			Name: "machinePoolReadyHandler",
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
			Name:         "machinePoolCreatingConditionHandler",
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
			Name:         "machinePoolUpgradingConditionHandler",
			UpdateStatus: true,
		}

		upgradingHandler, err = upgrading.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var machinePoolConditionsHandler *composite.Handler
	{
		c := composite.HandlerConfig{
			CtrlClient: config.CtrlClient,
			Logger:     config.Logger,
			Name:       config.Name,
			Handlers: []handler.Interface{
				infrastructureReadyHandler,
				replicasReadyHandler,
				readyHandler,
				creatingHandler,
				upgradingHandler,
			},
		}

		machinePoolConditionsHandler, err = composite.NewHandler(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return machinePoolConditionsHandler, nil
}
