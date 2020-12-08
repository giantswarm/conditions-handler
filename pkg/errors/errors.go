package errors

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var InvalidConfigError = &microerror.Error{
	Kind: "InvalidConfigError",
}

// IsInvalidConfig asserts InvalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == InvalidConfigError
}

var UnknownKindError = &microerror.Error{
	Kind: "UnknownKindError",
}

// IsUnknownKindError asserts UnknownKindError.
func IsUnknownKindError(err error) bool {
	return microerror.Cause(err) == UnknownKindError
}

func IsFailedToRetrieveExternalObject(err error) bool {
	if err == nil {
		return false
	}

	errorString := err.Error()
	if strings.Contains(errorString, "failed to retrieve") &&
		strings.Contains(errorString, "external object") {
		return true
	} else {
		return false
	}
}

var WrongTypeError = &microerror.Error{
	Kind: "WrongTypeError",
}

// IsWrongTypeError asserts WrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == WrongTypeError
}
