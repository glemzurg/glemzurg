package model_state

import (
	"fmt"

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
func (a *Action) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := a.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.ActionKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTION {
		return coreerr.NewWithValues(ctx, coreerr.ActionKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for action", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_ACTION)
	}

	if a.Name == "" {
		return coreerr.New(ctx, coreerr.ActionNameRequired, "Name is required", "Name")
	}
	if badChar := coreerr.ValidateNameChars(a.Name); badChar != "" {
		return coreerr.NewWithValues(ctx, coreerr.ActionNameInvalidChars, fmt.Sprintf("Name contains invalid character %q", badChar), "Name", a.Name, "A-Za-z0-9 space hyphen underscore")
	}

	reqLetTargets := make(map[string]bool)
	for i, req := range a.Requires {
		childCtx := ctx.Child("requires", fmt.Sprintf("%d", i))
		if err := req.Validate(childCtx); err != nil {
			return err
		}
		if req.Type != model_logic.LogicTypeAssessment && req.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(childCtx, coreerr.ActionRequiresTypeInvalid, fmt.Sprintf("requires %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeAssessment, model_logic.LogicTypeLet, req.Type), "Requires", req.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeAssessment, model_logic.LogicTypeLet))
		}
		if req.Type == model_logic.LogicTypeLet {
			if reqLetTargets[req.Target] {
				return coreerr.NewWithValues(childCtx, coreerr.ActionRequiresDuplicateLet, fmt.Sprintf("requires %d: duplicate let target %q", i, req.Target), "Requires", req.Target, "")
			}
			reqLetTargets[req.Target] = true
		}
	}
	guarTargets := make(map[string]bool)
	for i, guar := range a.Guarantees {
		childCtx := ctx.Child("guarantees", fmt.Sprintf("%d", i))
		if err := guar.Validate(childCtx); err != nil {
			return err
		}
		if guar.Type != model_logic.LogicTypeStateChange && guar.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeTypeInvalid, fmt.Sprintf("guarantee %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeStateChange, model_logic.LogicTypeLet, guar.Type), "Guarantees", guar.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeStateChange, model_logic.LogicTypeLet))
		}
		// Each guarantee and let must set a unique target.
		if guarTargets[guar.Target] {
			if guar.Type == model_logic.LogicTypeLet {
				return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeDuplicateLet, fmt.Sprintf("guarantee %d: duplicate let target %q", i, guar.Target), "Guarantees", guar.Target, "")
			}
			return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeDuplicateTarget, fmt.Sprintf("guarantee %d: duplicate target %q — each attribute can only be set once per action", i, guar.Target), "Guarantees", guar.Target, "")
		}
		guarTargets[guar.Target] = true
	}
	safetyLetTargets := make(map[string]bool)
	for i, rule := range a.SafetyRules {
		childCtx := ctx.Child("safetyRules", fmt.Sprintf("%d", i))
		if err := rule.Validate(childCtx); err != nil {
			return err
		}
		if rule.Type != model_logic.LogicTypeSafetyRule && rule.Type != model_logic.LogicTypeLet {
			return coreerr.NewWithValues(childCtx, coreerr.ActionSafetyTypeInvalid, fmt.Sprintf("safety rule %d: logic kind must be '%s' or '%s', got '%s'", i, model_logic.LogicTypeSafetyRule, model_logic.LogicTypeLet, rule.Type), "SafetyRules", rule.Type, fmt.Sprintf("one of: %s, %s", model_logic.LogicTypeSafetyRule, model_logic.LogicTypeLet))
		}
		if rule.Type == model_logic.LogicTypeLet {
			if safetyLetTargets[rule.Target] {
				return coreerr.NewWithValues(childCtx, coreerr.ActionSafetyDuplicateLet, fmt.Sprintf("safety rule %d: duplicate let target %q", i, rule.Target), "SafetyRules", rule.Target, "")
			}
			safetyLetTargets[rule.Target] = true
		}
	}

	return nil
}

// ValidateWithParent validates the Action, its key's parent relationship, and all children.
// The parent must be a Class.
func (a *Action) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Validate logic children with action as parent.
	for i := range a.Requires {
		childCtx := ctx.Child("requires", fmt.Sprintf("%d", i))
		if err := a.Requires[i].ValidateWithParent(childCtx, &a.Key); err != nil {
			return err
		}
	}
	for i := range a.Guarantees {
		childCtx := ctx.Child("guarantees", fmt.Sprintf("%d", i))
		if err := a.Guarantees[i].ValidateWithParent(childCtx, &a.Key); err != nil {
			return err
		}
	}
	for i := range a.SafetyRules {
		childCtx := ctx.Child("safetyRules", fmt.Sprintf("%d", i))
		if err := a.SafetyRules[i].ValidateWithParent(childCtx, &a.Key); err != nil {
			return err
		}
	}
	// Validate all children.
	for i := range a.Parameters {
		childCtx := ctx.Child("parameter", fmt.Sprintf("%d", i))
		if err := a.Parameters[i].ValidateWithParent(childCtx); err != nil {
			return err
		}
	}
	return nil
}
