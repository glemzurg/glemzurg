package model_logic

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// IsAssociationClassReify reports whether this guarantee reifies an association class.
// Detected by a non-empty endpoint_selector; Target is the association class TLA name.
func IsAssociationClassReify(l Logic) bool {
	return strings.TrimSpace(l.EndpointSelectorSpec.Specification) != ""
}

// SetEndpointSelectorSpec sets the expression naming the far-side endpoint for AC reify.
func (l *Logic) SetEndpointSelectorSpec(spec logic_spec.ExpressionSpec) {
	l.EndpointSelectorSpec = spec
}

func validateLogicAssociationClassFields(ctx *coreerr.ValidationContext, l *Logic) error {
	if !IsAssociationClassReify(*l) {
		return nil
	}
	// Association-class reify: action guarantee state_change only.
	if l.Type != LogicTypeStateChange {
		return coreerr.New(ctx, coreerr.LogicAssocClassContextInvalid,
			"endpoint_selector requires type state_change", "Type")
	}
	if l.Key.KeyType != identity.KEY_TYPE_ACTION_GUARANTEE {
		return coreerr.New(ctx, coreerr.LogicAssocClassContextInvalid,
			"endpoint_selector may only appear in action guarantees", "Key")
	}
	if strings.TrimSpace(l.Target) == "" {
		return coreerr.NewWithValues(ctx, coreerr.LogicTargetRequired,
			"association-class reify requires a non-empty target (association class name)",
			"Target", "", "association class TLA name")
	}
	if strings.TrimSpace(l.Spec.Specification) == "" {
		return coreerr.New(ctx, coreerr.LogicAssocClassSpecRequired,
			"association-class reify requires a creation specification", "Spec.Specification")
	}
	return nil
}
