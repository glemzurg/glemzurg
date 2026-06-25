package invariants

import (
	"strings"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
)

// NullableWhenSetSpecification wraps inner so NULL/absent is valid and inner applies only when set.
// Used for nullable attribute and parameter invariants; not for requires or class invariants.
func NullableWhenSetSpecification(subject, inner string) string {
	return "IF " + subject + " = NULL THEN TRUE ELSE " + inner
}

// LogicHasNullableWhenUnsetGuard reports whether expr already treats NULL/absent as vacuously true.
func LogicHasNullableWhenUnsetGuard(expr me.Expression) bool {
	ifte, ok := expr.(*me.IfThenElse)
	if !ok {
		return false
	}
	_, ok = nullCompareParam(ifte.Condition)
	return ok && isTrueLiteral(ifte.Then)
}

// LogicSpecHasNullableWhenUnsetGuard reports whether a logic spec already guards NULL/absent.
func LogicSpecHasNullableWhenUnsetGuard(spec logic_spec.ExpressionSpec) bool {
	if LogicHasNullableWhenUnsetGuard(spec.Expression) {
		return true
	}
	upper := strings.ToUpper(spec.Specification)
	return strings.Contains(upper, "= NULL THEN TRUE")
}

func nullCompareParam(expr me.Expression) (string, bool) {
	cmp, ok := expr.(*me.Compare)
	if !ok || cmp.Op != me.CompareEq {
		return "", false
	}

	localVar, ok := cmp.Left.(*me.LocalVar)
	if !ok || !isEmptySetLiteral(cmp.Right) {
		return "", false
	}

	return localVar.Name, true
}

func isEmptySetLiteral(expr me.Expression) bool {
	literal, ok := expr.(*me.SetLiteral)
	return ok && len(literal.Elements) == 0
}

func isTrueLiteral(expr me.Expression) bool {
	literal, ok := expr.(*me.BoolLiteral)
	return ok && literal.Value
}
