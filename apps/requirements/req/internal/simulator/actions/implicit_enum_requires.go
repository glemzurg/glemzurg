package actions

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

const (
	implicitEnumRequireSubKeyPrefix = "implicit_enum_"
	logicOwnerKindAction            = "action"
	logicOwnerKindQuery             = "query"
)

// EffectiveRequires returns explicit requires plus synthesized enum-membership assessments
// for enumeration parameters not already constrained in explicit requires.
func EffectiveRequires(
	ownerKey identity.Key,
	ownerKind string,
	params []model_state.Parameter,
	explicitRequires []model_logic.Logic,
) ([]model_logic.Logic, error) {
	implicit, err := implicitEnumRequires(ownerKey, ownerKind, params, explicitRequires)
	if err != nil {
		return nil, err
	}
	if len(implicit) == 0 {
		return explicitRequires, nil
	}
	combined := make([]model_logic.Logic, 0, len(explicitRequires)+len(implicit))
	combined = append(combined, explicitRequires...)
	combined = append(combined, implicit...)
	return combined, nil
}

func implicitEnumRequires(
	ownerKey identity.Key,
	ownerKind string,
	params []model_state.Parameter,
	explicitRequires []model_logic.Logic,
) ([]model_logic.Logic, error) {
	if len(params) == 0 {
		return nil, nil
	}

	explicitEnum := extractParameterConstraints(explicitRequires).enumValues
	paramNames := parameterNames(params)
	classKey, err := identity.ParseKey(ownerKey.ParentKey)
	if err != nil {
		return nil, fmt.Errorf("implicit enum requires: owner %q: %w", ownerKey.String(), err)
	}

	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: paramNames,
	}
	pf := convert.NewExpressionParseFunc(ctx)

	var implicit []model_logic.Logic
	ordinal := 0
	for _, param := range params {
		if explicitEnum != nil {
			if _, covered := explicitEnum[param.Name]; covered {
				continue
			}
		}
		values := model_data_type.EnumerationValues(param.DataType)
		if len(values) == 0 {
			continue
		}

		specText := enumMembershipSpecification(param.Name, values, param.Nullable)
		spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specText, pf)
		if err != nil {
			return nil, fmt.Errorf(
				"implicit enum require for parameter %q (owner %q): %w",
				param.Name, ownerKey.String(), err,
			)
		}

		requireKey, err := newOwnerRequireKey(ownerKey, ownerKind, fmt.Sprintf("%s%d", implicitEnumRequireSubKeyPrefix, ordinal))
		if err != nil {
			return nil, err
		}
		ordinal++

		implicit = append(implicit, model_logic.NewLogic(
			requireKey,
			model_logic.LogicTypeAssessment,
			fmt.Sprintf("Parameter %q must be one of the allowed enumeration values.", param.Name),
			"",
			spec,
			nil,
		))
	}

	return implicit, nil
}

func newOwnerRequireKey(ownerKey identity.Key, ownerKind, subKey string) (identity.Key, error) {
	switch ownerKind {
	case logicOwnerKindAction:
		return identity.NewActionRequireKey(ownerKey, subKey)
	case logicOwnerKindQuery:
		return identity.NewQueryRequireKey(ownerKey, subKey)
	default:
		return identity.Key{}, fmt.Errorf("unsupported owner kind %q for implicit enum requires", ownerKind)
	}
}

func enumMembershipSpecification(paramName string, values []string, nullable bool) string {
	membership := fmt.Sprintf(`%s \in %s`, paramName, formatTLAPlusStringSet(values))
	if !nullable {
		return membership
	}
	return fmt.Sprintf(`IF %s = NULL THEN TRUE ELSE %s`, paramName, membership)
}

func formatTLAPlusStringSet(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(quoted, ", ") + "}"
}

// HasImplicitEnumRequires reports whether params include parsed enumeration rules that
// would synthesize at least one implicit require.
func HasImplicitEnumRequires(params []model_state.Parameter, explicitRequires []model_logic.Logic) bool {
	explicitEnum := extractParameterConstraints(explicitRequires).enumValues
	for _, param := range params {
		if explicitEnum != nil {
			if _, covered := explicitEnum[param.Name]; covered {
				continue
			}
		}
		if len(model_data_type.EnumerationValues(param.DataType)) > 0 {
			return true
		}
	}
	return false
}
