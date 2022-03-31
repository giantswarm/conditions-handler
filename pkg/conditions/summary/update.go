package summary

import (
	"github.com/giantswarm/conditions/pkg/conditions"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capiconditions "sigs.k8s.io/cluster-api/util/conditions"
)

func update(object conditions.Object, conditionTypesToSummarize []capi.ConditionType, ignoreOptions ...conditions.CheckOption) {
	var conditionsToSummarizeOption capiconditions.MergeOption

	if len(ignoreOptions) > 0 {
		var conditionTypes []capi.ConditionType
		for _, conditionType := range conditionTypesToSummarize {
			condition := capiconditions.Get(object, conditionType)
			ignore := false
			for _, ignoreOption := range ignoreOptions {
				if ignoreOption(condition) {
					ignore = true
					break
				}
			}

			if !ignore {
				conditionTypes = append(conditionTypes, conditionType)
			}
		}
		conditionsToSummarizeOption = capiconditions.WithConditions(conditionTypes...)
	} else {
		conditionsToSummarizeOption = capiconditions.WithConditions(conditionTypesToSummarize...)
	}

	capiconditions.SetSummary(
		object,
		conditionsToSummarizeOption,
		capiconditions.AddSourceRef())
}
