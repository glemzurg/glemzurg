package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
)

// Thin forwarders so call sites with a local variable named instance can still
// reach instance package functions (Go shadows the package name).

func checkParameterTypeSpecs(
	params []model_state.Parameter,
	sourceKey identity.Key,
	sourceName string,
	sourceKind string,
	instanceID instance.ID,
	classKey identity.Key,
) instance.ViolationErrors {
	return instance.CheckParameterTypeSpecs(params, sourceKey, sourceName, sourceKind, instanceID, classKey)
}
