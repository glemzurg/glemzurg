package evaluator

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalAssociationRef evaluates a bare association name in class scope (e.g. IsSubdividedInto
// in { x \in IsSubdividedInto.CurrencyWalletDefinition : … }).
// Host associations with an association class return AssociationRelation so AC member
// access matches self.IsSubdividedInto.CurrencyWalletDefinition; plain associations
// return a set of endpoint records.
func evalAssociationRef(n *me.AssociationRef, bindings *Bindings) *EvalResult {
	self, relCtx, errResult := associationRefSelfAndContext(bindings)
	if errResult != nil {
		return errResult
	}
	assocKey := AssociationKey(n.AssociationKey.String())
	classKey := associationRefClassKey(self, relCtx, bindings)
	if relInfo := relCtx.FindRelationByAssociationKey(classKey, assocKey); relInfo != nil {
		return evalRelationTraversal(self, relInfo, relCtx)
	}
	// Unknown relation metadata: degrade to endpoint set (same as pre-AC plain nav).
	records := relCtx.GetRelatedRecords(self, assocKey, false)
	return NewEvalResult(object.NewSetFromElements(recordsToObjects(records)))
}

func associationRefSelfAndContext(bindings *Bindings) (*object.Record, *RelationContext, *EvalResult) {
	self := bindings.Self()
	if self == nil {
		return nil, nil, NewEvalError("association reference requires self")
	}
	relCtx := bindings.RelationContext()
	if relCtx == nil {
		return nil, nil, NewEvalError("association reference requires relation context")
	}
	return self, relCtx, nil
}

func associationRefClassKey(self *object.Record, relCtx *RelationContext, bindings *Bindings) string {
	if classKey, ok := relCtx.ClassKeyForRecord(self); ok {
		return classKey
	}
	return bindings.SelfClassKey()
}
