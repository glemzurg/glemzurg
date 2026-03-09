package model_state

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Action is what happens in a transition between states.
type Action struct {
	Key     identity.Key
	Name    string
	Details string
	// Children
	Parameters  []Parameter         // Typed parameters for this action.
	Requires    []model_logic.Logic // Preconditions to enter this action (must not contain primed variables).
	Guarantees  []model_logic.Logic // Postconditions of this action (primed assignments only, e.g., self.field' = expr).
	SafetyRules []model_logic.Logic // Boolean assertions that must reference primed variables.
}

func NewAction(key identity.Key, name, details string, requires, guarantees, safetyRules []model_logic.Logic, parameters []Parameter) Action {
	return Action{
		Key:         key,
		Name:        name,
		Details:     details,
		Requires:    requires,
		Guarantees:  guarantees,
		SafetyRules: safetyRules,
		Parameters:  parameters,
	}
}

// Validate validates the Action struct.
//
//complexity:cyclo:warn=20,fail=20 Sequential field validation.
func (a *Action) Validate() error {
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return coreerr.New(coreerr.ActionKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTION {
		return coreerr.NewWithValues(coreerr.ActionKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for action", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_ACTION)
	}

	if a.Name == "" {
		return coreerr.New(coreerr.ActionNameRequired, "Name is required", "Name")
	}

	reqLetTargets := make(map[string]bool)
	for i, req := range a.Requires {
		if err := req.Validate(); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
		if req.Type != model_logic.LogicTypeAssessment && req.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(coreerr.ActionRequiresTypeInvalid, fmt.Sprintf("requires %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, req.Type), "Requires", req.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet))
		}
		if req.Type == model_logic.LogicTypeLet {
			if reqLetTargets[req.Target] {
				return coreerr.NewWithValues(coreerr.ActionRequiresDuplicateLet, fmt.Sprintf("requires %d: duplicate let target %q", i, req.Target), "Requires", req.Target, "")
			}
			reqLetTargets[req.Target] = true
		}
	}
	guarTargets := make(map[string]bool)
	for i, guar := range a.Guarantees {
		if err := guar.Validate(); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
		if guar.Type != model_logic.LogicTypeStateChange && guar.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(coreerr.ActionGuaranteeTypeInvalid, fmt.Sprintf("guarantee %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeStateChange, model_logic.LogicTypeLet, guar.Type), "Guarantees", guar.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeStateChange, model_logic.LogicTypeLet))
		}
		// Each guarantee and let must set a unique target.
		if guarTargets[guar.Target] {
			if guar.Type == model_logic.LogicTypeLet {
				return coreerr.NewWithValues(coreerr.ActionGuaranteeDuplicateLet, fmt.Sprintf("guarantee %d: duplicate let target %q", i, guar.Target), "Guarantees", guar.Target, "")
			}
			return coreerr.NewWithValues(coreerr.ActionGuaranteeDuplicateTarget, fmt.Sprintf("guarantee %d: duplicate target %q — each attribute can only be set once per action", i, guar.Target), "Guarantees", guar.Target, "")
		}
		guarTargets[guar.Target] = true
	}
	safetyLetTargets := make(map[string]bool)
	for i, rule := range a.SafetyRules {
		if err := rule.Validate(); err != nil {
			return errors.Wrapf(err, "safety rule %d", i)
		}
		if rule.Type != model_logic.LogicTypeSafetyRule && rule.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(coreerr.ActionSafetyTypeInvalid, fmt.Sprintf("safety rule %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeSafetyRule, model_logic.LogicTypeLet, rule.Type), "SafetyRules", rule.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeSafetyRule, model_logic.LogicTypeLet))
		}
		if rule.Type == model_logic.LogicTypeLet {
			if safetyLetTargets[rule.Target] {
				return coreerr.NewWithValues(coreerr.ActionSafetyDuplicateLet, fmt.Sprintf("safety rule %d: duplicate let target %q", i, rule.Target), "SafetyRules", rule.Target, "")
			}
			safetyLetTargets[rule.Target] = true
		}
	}

	return nil
}

// ValidateWithParent validates the Action, its key's parent relationship, and all children.
// The parent must be a Class.
func (a *Action) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate logic children with action as parent.
	for i := range a.Requires {
		if err := a.Requires[i].ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "requires %d", i)
		}
	}
	for i := range a.Guarantees {
		if err := a.Guarantees[i].ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "guarantee %d", i)
		}
	}
	for i := range a.SafetyRules {
		if err := a.SafetyRules[i].ValidateWithParent(&a.Key); err != nil {
			return errors.Wrapf(err, "safety rule %d", i)
		}
	}
	// Validate all children.
	for i := range a.Parameters {
		if err := a.Parameters[i].ValidateWithParent(); err != nil {
			return err
		}
	}
	return nil
}
