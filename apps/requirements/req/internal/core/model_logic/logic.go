package model_logic

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// NotationTLAPlus is the only supported notation for logic specifications.
const NotationTLAPlus = "tla_plus"

// Logic kinds.
const (
	LogicTypeAssessment  = "assessment"   // True/false boolean check (no primed variables).
	LogicTypeStateChange = "state_change" // Primed assignment: attribute' = expression.
	LogicTypeQuery       = "query"        // Defines temporary named return values.
	LogicTypeSafetyRule  = "safety_rule"  // Boolean check referencing both prior and new state (has primed).
	LogicTypeValue       = "value"        // Single unnamed value expression (global functions).
	LogicTypeLet         = "let"          // Local variable definition: target = expression.
)

// validLogicTypes is the set of valid Logic.Type values.
var validLogicTypes = map[string]bool{
	LogicTypeAssessment:  true,
	LogicTypeStateChange: true,
	LogicTypeQuery:       true,
	LogicTypeSafetyRule:  true,
	LogicTypeValue:       true,
	LogicTypeLet:         true,
}

// Logic represents a formal logic specification attached to a model element.
type Logic struct {
	Key            identity.Key              // The key is unique in the whole model, and built on the key of the containing object.
	Type           string                    // One of: assessment, state_change, query, safety_rule, value, let.
	Description    string                    // Required human-readable description.
	Target         string                    // Identifier or attribute to set. Required for state_change and query types.
	Spec           logic_spec.ExpressionSpec // Notation + Specification + Expression (the reusable trio).
	TargetTypeSpec *logic_spec.TypeSpec      // Optional: declared result type of the logic's target.
}

// NewLogic creates a new Logic.
func NewLogic(key identity.Key, logicType, description, target string, spec logic_spec.ExpressionSpec, targetTypeSpec *logic_spec.TypeSpec) Logic {
	return Logic{
		Key:            key,
		Type:           logicType,
		Description:    description,
		Target:         target,
		Spec:           spec,
		TargetTypeSpec: targetTypeSpec,
	}
}

// Validate validates the Logic struct.
func (l *Logic) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := l.Key.ValidateWithContext(ctx); err != nil {
		return err
	}
	// Type is required.
	if l.Type == "" {
		return coreerr.NewWithValues(ctx, coreerr.LogicTypeRequired, "Type is required", "Type", "", "one of: assessment, state_change, query, safety_rule, value, let")
	}
	// Type must be a valid value.
	if !validLogicTypes[l.Type] {
		return coreerr.NewWithValues(ctx, coreerr.LogicTypeInvalid, fmt.Sprintf("Type '%s' is not valid", l.Type), "Type", l.Type, "one of: assessment, state_change, query, safety_rule, value, let")
	}
	// Description is required.
	if l.Description == "" {
		return coreerr.New(ctx, coreerr.LogicDescRequired, "Description is required", "Description")
	}
	// Target validation based on logic type.
	switch l.Type {
	case LogicTypeStateChange, LogicTypeQuery, LogicTypeLet:
		if l.Target == "" {
			return coreerr.NewWithValues(ctx, coreerr.LogicTargetRequired, fmt.Sprintf("logic %q of type %q requires a non-empty target", l.Key.String(), l.Type), "Target", "", "non-empty string")
		}
		// Query and let targets cannot start with "_".
		if (l.Type == LogicTypeQuery || l.Type == LogicTypeLet) && strings.HasPrefix(l.Target, "_") {
			return coreerr.NewWithValues(ctx, coreerr.LogicTargetNoUnderscore, fmt.Sprintf("logic %q of type %q has target %q starting with '_' which is not allowed", l.Key.String(), l.Type, l.Target), "Target", l.Target, "")
		}
	case LogicTypeAssessment, LogicTypeSafetyRule, LogicTypeValue:
		if l.Target != "" {
			return coreerr.NewWithValues(ctx, coreerr.LogicTargetMustBeEmpty, fmt.Sprintf("logic %q of type %q must not have a target, got %q", l.Key.String(), l.Type, l.Target), "Target", l.Target, "empty string")
		}
	}
	// Validate the ExpressionSpec.
	if err := l.Spec.Validate(ctx); err != nil {
		return coreerr.New(ctx, coreerr.LogicSpecInvalid, fmt.Sprintf("logic %q spec: %s", l.Key.String(), err.Error()), "Spec")
	}
	// Validate TargetTypeSpec if present.
	if l.TargetTypeSpec != nil {
		if err := l.TargetTypeSpec.Validate(ctx); err != nil {
			return coreerr.New(ctx, coreerr.LogicTargetTypespecInvalid, fmt.Sprintf("logic %q target type spec: %s", l.Key.String(), err.Error()), "TargetTypeSpec")
		}
	}
	return nil
}

// ValidateWithParent validates the Logic and its key's parent relationship.
func (l *Logic) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	if err := l.Validate(ctx); err != nil {
		return err
	}
	if err := l.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	return nil
}
