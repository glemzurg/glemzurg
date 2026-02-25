package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// lookupRelation checks if a field name corresponds to a relation on the given class.
// It checks both forward relations (.Name) and reverse relations (._Name).
// Returns nil if no relation is found.
func lookupRelation(classKey, fieldName string, relCtx *RelationContext) *RelationInfo {
	if relCtx == nil || classKey == "" {
		return nil
	}

	// The GetRelation method handles both forward and reverse lookup
	return relCtx.GetRelation(classKey, fieldName)
}

// evalRelationTraversal evaluates a relation field access on a record.
// It returns a Set of related records by querying the link table.
func evalRelationTraversal(record *object.Record, relInfo *RelationInfo, relCtx *RelationContext) *EvalResult {
	if record == nil {
		return NewEvalError("cannot traverse relation on nil record")
	}
	if relInfo == nil {
		return NewEvalError("relation info is nil")
	}
	if relCtx == nil {
		return NewEvalError("relation context is nil")
	}

	// Get related records from the link table
	relatedRecords := relCtx.GetRelatedRecords(record, relInfo.AssociationKey, relInfo.Reverse)

	// Convert to a Set of Objects
	elements := make([]object.Object, len(relatedRecords))
	for i, rec := range relatedRecords {
		elements[i] = rec
	}

	return NewEvalResult(object.NewSetFromElements(elements))
}

// evalRelationTraversalOnObject evaluates a relation field access on any object.
// If the object is a Record, it performs the relation traversal.
// Otherwise, returns an error.
func evalRelationTraversalOnObject(obj object.Object, relInfo *RelationInfo, relCtx *RelationContext) *EvalResult {
	record, ok := obj.(*object.Record)
	if !ok {
		return NewEvalError("cannot traverse relation on non-record object: %s", obj.Type())
	}
	return evalRelationTraversal(record, relInfo, relCtx)
}
