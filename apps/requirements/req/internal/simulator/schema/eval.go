package schema

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
)

// NewEvalContext loads lowered model expressions into a registry-backed eval
// context so derived attributes and other logic can call model globals.
// Uses the private owned model; callers never receive *core.Model.
func (s *Schema) NewEvalContext() (*evaluator.EvalContext, error) {
	if s == nil || s.model == nil {
		return nil, fmt.Errorf("schema.NewEvalContext: schema has no model")
	}
	result := model_bridge.NewLoader().LoadFromModel(s.model)
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
