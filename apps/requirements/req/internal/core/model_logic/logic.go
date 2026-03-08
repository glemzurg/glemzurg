package model_logic

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
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
	Spec           model_spec.ExpressionSpec // Notation + Specification + Expression (the reusable trio).
	TargetTypeSpec *model_spec.TypeSpec      // Optional: declared result type of the logic's target.
}

// NewLogic creates a new Logic and validates it.
func NewLogic(key identity.Key, logicType, description, target string, spec model_spec.ExpressionSpec, targetTypeSpec *model_spec.TypeSpec) (logic Logic, err error) {
	logic = Logic{
		Key:            key,
		Type:           logicType,
		Description:    description,
		Target:         target,
		Spec:           spec,
		TargetTypeSpec: targetTypeSpec,
	}

	if err = logic.Validate(); err != nil {
		return Logic{}, err
	}

	return logic, nil
}

// Validate validates the Logic struct.
func (l *Logic) Validate() error {
	// Validate the key.
	if err := l.Key.Validate(); err != nil {
		return err
	}
	// Type is required.
	if l.Type == "" {
		return coreerr.NewWithValues(coreerr.LogicTypeRequired, "Type is required", "Type", "", "one of: assessment, state_change, query, safety_rule, value, let")
	}
	// Type must be a valid value.
	if !validLogicTypes[l.Type] {
		return coreerr.NewWithValues(coreerr.LogicTypeInvalid, fmt.Sprintf("Type '%s' is not valid", l.Type), "Type", l.Type, "one of: assessment, state_change, query, safety_rule, value, let")
	}
	// Description is required.
	if l.Description == "" {
		return coreerr.New(coreerr.LogicDescRequired, "Description is required", "Description")
	}
	// Target validation based on logic type.
	switch l.Type {
	case LogicTypeStateChange, LogicTypeQuery, LogicTypeLet:
		if l.Target == "" {
			return coreerr.NewWithValues(coreerr.LogicTargetRequired, fmt.Sprintf("logic %q of type %q requires a non-empty target", l.Key.String(), l.Type), "Target", "", "non-empty string")
		}
		// Query and let targets cannot start with "_".
		if (l.Type == LogicTypeQuery || l.Type == LogicTypeLet) && strings.HasPrefix(l.Target, "_") {
			return coreerr.NewWithValues(coreerr.LogicTargetNoUnderscore, fmt.Sprintf("logic %q of type %q has target %q starting with '_' which is not allowed", l.Key.String(), l.Type, l.Target), "Target", l.Target, "")
		}
	case LogicTypeAssessment, LogicTypeSafetyRule, LogicTypeValue:
		if l.Target != "" {
			return coreerr.NewWithValues(coreerr.LogicTargetMustBeEmpty, fmt.Sprintf("logic %q of type %q must not have a target, got %q", l.Key.String(), l.Type, l.Target), "Target", l.Target, "empty string")
		}
	}
	// Validate the ExpressionSpec.
	if err := l.Spec.Validate(); err != nil {
		return coreerr.New(coreerr.LogicSpecInvalid, fmt.Sprintf("logic %q spec: %s", l.Key.String(), err.Error()), "Spec")
	}
	// Validate TargetTypeSpec if present.
	if l.TargetTypeSpec != nil {
		if err := l.TargetTypeSpec.Validate(); err != nil {
			return coreerr.New(coreerr.LogicTargetTypespecInvalid, fmt.Sprintf("logic %q target type spec: %s", l.Key.String(), err.Error()), "TargetTypeSpec")
		}
	}
	return nil
}

// ValidateWithParent validates the Logic and its key's parent relationship.
func (l *Logic) ValidateWithParent(parent *identity.Key) error {
	if err := l.Validate(); err != nil {
		return err
	}
	if err := l.Key.ValidateParent(parent); err != nil {
		return err
	}
	return nil
}
