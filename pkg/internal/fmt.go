package internal

import (
	"fmt"

	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func SprintComparedCondition(condition *capi.Condition) string {
	var text string
	if condition != nil {
		text = fmt.Sprintf(
			"%s(Status=%q, Reason=%q, Severity=%q, [optional: Message=%q])",
			condition.Type,
			condition.Status,
			condition.Reason,
			condition.Severity,
			condition.Message)
	} else {
		text = "condition is nil"
	}

	return text
}
