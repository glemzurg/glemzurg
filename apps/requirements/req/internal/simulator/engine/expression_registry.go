package engine

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
)

// setupExpressionRegistry loads lowered model expressions into a registry-backed
// eval context so derived attributes and other logic can call model globals.
func setupExpressionRegistry(model *core.Model) (*evaluator.EvalContext, error) {
	result := model_bridge.NewLoader().LoadFromModel(model)
	if result.HasErrors() {
		return nil, fmt.Errorf("expression registry: %s", formatLoadErrors(result.Errors))
	}
	return &evaluator.EvalContext{
		IRRegistry: registry.NewRuntimeAdapter(result.Registry),
	}, nil
}

func formatLoadErrors(errs []error) string {
	if len(errs) == 0 {
		return "unknown load error"
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}
