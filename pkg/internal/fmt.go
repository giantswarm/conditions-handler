package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func SprintComparedCondition(condition *capi.Condition) string {
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
		text = "condition is nil"
	}

	return text
}

func LogObjectJson(ctx context.Context, logger micrologger.Logger, obj interface{}) error {
	objJson, err := json.Marshal(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	logger.Debugf(ctx, "Object JSON: %s", objJson)
	return nil
}
