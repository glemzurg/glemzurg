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
// Plain associations return a Set of related endpoint records. Host associations
// with an association class return AssociationRelation (endpoint image + link rows);
// set/bag ops coerce that to the endpoint set via CoerceToSet.
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

	endpointRecords := relCtx.GetRelatedRecords(record, relInfo.AssociationKey, relInfo.Reverse)
	endpointElements := recordsToObjects(endpointRecords)

	if relInfo.LinkClassMember == "" {
		return NewEvalResult(object.NewSetFromElements(endpointElements))
	}

	linkByEndpoint := relCtx.GetAssociationClassLinksByEndpoint(
		record,
		relInfo.AssociationKey,
		relInfo.Reverse,
	)

	return NewEvalResult(object.NewAssociationRelation(
		object.NewSetFromElements(endpointElements),
		relInfo.LinkClassMember,
		linkByEndpoint,
	))
}

func recordsToObjects(records []*object.Record) []object.Object {
	elements := make([]object.Object, len(records))
	for i, rec := range records {
		elements[i] = rec
	}
	return elements
}
