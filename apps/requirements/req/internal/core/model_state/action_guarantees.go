package model_state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
)

func validateActionGuarantees(ctx *coreerr.ValidationContext, guarantees []model_logic.Logic) error {
	stateChangeTargets := make(map[string]bool)
	deleteTargets := make(map[string]bool)
	letTargets := make(map[string]bool)
	for i, guar := range guarantees {
		childCtx := ctx.Child("guarantees", fmt.Sprintf("%d", i))
		if err := guar.Validate(childCtx); err != nil {
			return err
		}
		if guar.Type != model_logic.LogicTypeStateChange && guar.Type != model_logic.LogicTypeLet && guar.Type != model_logic.LogicTypeDestroy {
			return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeTypeInvalid, fmt.Sprintf("guarantee %d: logic kind must be '%s', '%s', or '%s', got '%s'", i, model_logic.LogicTypeStateChange, model_logic.LogicTypeLet, model_logic.LogicTypeDestroy, guar.Type), "Guarantees", guar.Type, fmt.Sprintf("one of: %s, %s, %s", model_logic.LogicTypeStateChange, model_logic.LogicTypeLet, model_logic.LogicTypeDestroy))
		}
		switch guar.Type {
		case model_logic.LogicTypeLet:
			if letTargets[guar.Target] {
				return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeDuplicateLet, fmt.Sprintf("guarantee %d: duplicate let target %q", i, guar.Target), "Guarantees", guar.Target, "")
			}
			letTargets[guar.Target] = true
		case model_logic.LogicTypeDestroy:
			if deleteTargets[guar.Target] {
				return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeDuplicateDestroyTarget, fmt.Sprintf("guarantee %d: duplicate destroy target %q", i, guar.Target), "Guarantees", guar.Target, "")
			}
			deleteTargets[guar.Target] = true
		case model_logic.LogicTypeStateChange:
			if stateChangeTargets[guar.Target] {
				return coreerr.NewWithValues(childCtx, coreerr.ActionGuaranteeDuplicateTarget, fmt.Sprintf("guarantee %d: duplicate target %q — each attribute can only be set once per action", i, guar.Target), "Guarantees", guar.Target, "")
			}
			stateChangeTargets[guar.Target] = true
		}
	}
	return nil
}
