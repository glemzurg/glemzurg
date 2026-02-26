package model_logic

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression"
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
)

// _validate is the shared validator instance for this package.
var _validate = validator.New()

// Logic represents a formal logic specification attached to a model element.
type Logic struct {
	Key           identity.Key                  // The key is unique in the whole model, and built on the key of the containing object.
	Type          string                        `validate:"required,oneof=assessment state_change query safety_rule value"`
	Description   string                        `validate:"required"`
	Target        string                        // Identifier or attribute to set. Required for state_change and query types.
	Notation      string                        `validate:"required,oneof=tla_plus"`
	Specification string                        // Optional logic specification body.
	Expression    model_expression.Expression   // Optional structured expression tree (nil = no expression).
}

// NewLogic creates a new Logic and validates it.
func NewLogic(key identity.Key, logicType, description, target, notation, specification string, expression model_expression.Expression) (logic Logic, err error) {
	logic = Logic{
		Key:           key,
		Type:          logicType,
		Description:   description,
		Target:        target,
		Notation:      notation,
		Specification: specification,
		Expression:    expression,
	}

	if err = logic.Validate(); err != nil {
		return Logic{}, err
	}

	return logic, nil
}

// Validate validates the Logic struct.
func (l *Logic) Validate() error {
	if err := l.Key.Validate(); err != nil {
		return err
	}
	if err := _validate.Struct(l); err != nil {
		return err
	}
	// Target validation based on logic type.
	switch l.Type {
	case LogicTypeStateChange, LogicTypeQuery:
		if l.Target == "" {
			return errors.Errorf("logic %q of type %q requires a non-empty target", l.Key.String(), l.Type)
		}
		// Query targets cannot start with "_".
		if l.Type == LogicTypeQuery && strings.HasPrefix(l.Target, "_") {
			return errors.Errorf("logic %q of type %q has target %q starting with '_' which is not allowed", l.Key.String(), l.Type, l.Target)
		}
	case LogicTypeAssessment, LogicTypeSafetyRule, LogicTypeValue:
		if l.Target != "" {
			return errors.Errorf("logic %q of type %q must not have a target, got %q", l.Key.String(), l.Type, l.Target)
		}
	}
	// Validate expression if present.
	if l.Expression != nil {
		if err := l.Expression.Validate(); err != nil {
			return errors.Wrapf(err, "logic %q expression", l.Key.String())
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
