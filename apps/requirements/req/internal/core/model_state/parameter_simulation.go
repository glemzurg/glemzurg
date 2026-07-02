package model_state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ParameterSimulation holds simulator-only TLA+ for sampling an action parameter value.
// Requires are boolean assessments that must hold before the action is eligible;
// Specification evaluates to the parameter value when requires pass.
type ParameterSimulation struct {
	Details       string
	Requires      []model_logic.Logic
	Specification *model_logic.Logic
}

// HasSimulation reports whether any simulator sampling metadata is present.
func (s *ParameterSimulation) HasSimulation() bool {
	return s != nil && (len(s.Requires) > 0 || s.Specification != nil)
}

// Validate validates ParameterSimulation on an action-owned parameter.
func (s *ParameterSimulation) Validate(ctx *coreerr.ValidationContext, paramKey identity.Key) error {
	if s == nil {
		return nil
	}
	for i, req := range s.Requires {
		reqCtx := ctx.Child("requires", fmt.Sprintf("%d", i))
		if err := req.Validate(reqCtx); err != nil {
			return err
		}
		if req.Type != model_logic.LogicTypeAssessment {
			return coreerr.NewWithValues(reqCtx, coreerr.ParamSimulationRequireTypeInvalid,
				fmt.Sprintf("simulation requires %d: logic kind must be '%s', got '%s'", i, model_logic.LogicTypeAssessment, req.Type),
				"Requires", req.Type, model_logic.LogicTypeAssessment)
		}
		if err := req.ValidateWithParent(reqCtx, &paramKey); err != nil {
			return err
		}
	}
	if s.Specification != nil {
		specCtx := ctx.Child("specification", "")
		if err := s.Specification.Validate(specCtx); err != nil {
			return err
		}
		if s.Specification.Type != model_logic.LogicTypeValue {
			return coreerr.NewWithValues(specCtx, coreerr.ParamSimulationSpecTypeInvalid,
				fmt.Sprintf("simulation specification: logic kind must be '%s', got '%s'", model_logic.LogicTypeValue, s.Specification.Type),
				"Specification", s.Specification.Type, model_logic.LogicTypeValue)
		}
		if err := s.Specification.ValidateWithParent(specCtx, &paramKey); err != nil {
			return err
		}
	}
	return nil
}
