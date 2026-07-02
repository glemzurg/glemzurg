package evaluator

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func evalAssociationRef(n *me.AssociationRef, bindings *Bindings) *EvalResult {
	self := bindings.Self()
	if self == nil {
		return NewEvalError("association reference requires self")
	}
	relCtx := bindings.RelationContext()
	if relCtx == nil {
		return NewEvalError("association reference requires relation context")
	}

	assocKey := AssociationKey(n.AssociationKey.String())
	records := relCtx.GetRelatedRecords(self, assocKey, false)
	return NewEvalResult(object.NewSetFromElements(recordsToObjects(records)))
}
