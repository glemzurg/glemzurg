package model_state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ParameterSimulationRule is one alternative sampling path for an action parameter.
// When sampling, every rule whose Requires all hold is eligible; the simulator picks
// one eligible rule uniformly at random and evaluates its Specification.
type ParameterSimulationRule struct {
	Details       string
	Requires      []model_logic.Logic
	Specification *model_logic.Logic
}

// ParameterSimulation holds simulator-only TLA+ for sampling an action parameter value.
// Rules is the ordered list of alternative require/specification pairings.
type ParameterSimulation struct {
	Details string
	Rules   []ParameterSimulationRule
}

// HasSimulation reports whether any simulator sampling metadata is present.
func (s *ParameterSimulation) HasSimulation() bool {
	if s == nil {
		return false
	}
	if s.Details != "" {
		return true
	}
	for i := range s.Rules {
		if s.Rules[i].HasContent() {
			return true
		}
	}
	return false
}

// HasContent reports whether the rule carries requires, a specification, or details.
func (r *ParameterSimulationRule) HasContent() bool {
	return r != nil && (r.Details != "" || len(r.Requires) > 0 || r.Specification != nil)
}

// HasSpecification reports whether the rule can produce a sampled value.
func (r *ParameterSimulationRule) HasSpecification() bool {
	return r != nil && r.Specification != nil && r.Specification.Spec.Specification != ""
}

// Validate validates ParameterSimulation on an action-owned parameter.
func (s *ParameterSimulation) Validate(ctx *coreerr.ValidationContext, paramKey identity.Key) error {
	if s == nil {
		return nil
	}
	for i := range s.Rules {
		ruleCtx := ctx.Child("rules", fmt.Sprintf("%d", i))
		if err := s.Rules[i].Validate(ruleCtx, paramKey); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates one simulation rule.
func (r *ParameterSimulationRule) Validate(ctx *coreerr.ValidationContext, paramKey identity.Key) error {
	if r == nil {
		return nil
	}
	for i, req := range r.Requires {
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
	if r.Specification != nil {
		specCtx := ctx.Child("specification", "")
		if err := r.Specification.Validate(specCtx); err != nil {
			return err
		}
		if r.Specification.Type != model_logic.LogicTypeValue {
			return coreerr.NewWithValues(specCtx, coreerr.ParamSimulationSpecTypeInvalid,
				fmt.Sprintf("simulation specification: logic kind must be '%s', got '%s'", model_logic.LogicTypeValue, r.Specification.Type),
				"Specification", r.Specification.Type, model_logic.LogicTypeValue)
		}
		if err := r.Specification.ValidateWithParent(specCtx, &paramKey); err != nil {
			return err
		}
	}
	return nil
}
