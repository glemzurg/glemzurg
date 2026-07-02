package actions

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// peerFieldDistinctFromParamPattern is the simulator's structural match for
// ∀ v ∈ Class : v.field # Param — evaluated over peers excluding self.
type peerFieldDistinctFromParamPattern struct {
	ClassKey    identity.Key
	ClassName   string
	VarName     string
	FieldSubKey string
	ParamName   string
}

func detectPeerFieldDistinctFromParam(expr me.Expression) (peerFieldDistinctFromParamPattern, bool) {
	q, ok := expr.(*me.Quantifier)
	if !ok || q.Kind != me.QuantifierForall {
		return peerFieldDistinctFromParamPattern{}, false
	}

	classRef, ok := q.Domain.(*me.ClassRef)
	if !ok {
		return peerFieldDistinctFromParamPattern{}, false
	}

	cmp, ok := q.Predicate.(*me.Compare)
	if !ok || cmp.Op != me.CompareNeq {
		return peerFieldDistinctFromParamPattern{}, false
	}

	fieldAccess, ok := cmp.Left.(*me.FieldAccess)
	if !ok {
		return peerFieldDistinctFromParamPattern{}, false
	}
	localVar, ok := fieldAccess.Base.(*me.LocalVar)
	if !ok || localVar.Name != q.Variable {
		return peerFieldDistinctFromParamPattern{}, false
	}

	paramVar, ok := cmp.Right.(*me.LocalVar)
	if !ok {
		return peerFieldDistinctFromParamPattern{}, false
	}

	return peerFieldDistinctFromParamPattern{
		ClassKey:    classRef.ClassKey,
		ClassName:   classRef.Name,
		VarName:     q.Variable,
		FieldSubKey: fieldAccess.Field,
		ParamName:   paramVar.Name,
	}, true
}

func assessPeerFieldDistinctFromParam(pattern peerFieldDistinctFromParamPattern, bindings *evaluator.Bindings) bool {
	classSetVal, ok := bindings.GetValue(pattern.ClassName)
	if !ok {
		return false
	}
	classSet, ok := classSetVal.(*object.Set)
	if !ok {
		return false
	}

	paramVal, ok := bindings.GetValue(pattern.ParamName)
	if !ok {
		return false
	}

	selfRecord := bindings.Self()
	for _, elem := range classSet.Elements() {
		peerRecord, ok := elem.(*object.Record)
		if !ok {
			continue
		}
		if selfRecord != nil && peerRecord.Equals(selfRecord) {
			continue
		}
		fieldVal := peerRecord.Get(pattern.FieldSubKey)
		if peerFieldValueConflicts(fieldVal, paramVal) {
			return false
		}
	}
	return true
}

func peerFieldValueConflicts(peerVal, paramVal object.Object) bool {
	if object.IsNull(peerVal) && object.IsNull(paramVal) {
		return true
	}
	if object.IsNull(peerVal) || object.IsNull(paramVal) {
		return false
	}
	return evaluator.ObjectsEqual(peerVal, paramVal)
}
