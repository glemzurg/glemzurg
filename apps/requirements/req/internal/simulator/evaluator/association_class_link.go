package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func evalAssociationRelationFieldAccess(
	assocRel *object.AssociationRelation,
	field string,
	bindings *Bindings,
) *EvalResult {
	if field == assocRel.LinkClassMember() {
		return resolveAssociationClassLink(assocRel, bindings)
	}
	linkResult := resolveAssociationClassLink(assocRel, bindings)
	if linkResult.IsError() {
		return linkResult
	}
	return applyFieldChain(linkResult.Value, []string{field})
}

// resolveAssociationClassLink returns the single association-class row for a host traversal.
// With one endpoint it is unambiguous; with many, a uniquely bound local endpoint variable disambiguates.
func resolveAssociationClassLink(assocRel *object.AssociationRelation, bindings *Bindings) *EvalResult {
	endpointCount := assocRel.Endpoints().Size()
	if endpointCount == 0 {
		return NewEvalError("no association rows for association-class member %s", assocRel.LinkClassMember())
	}

	endpoint := soleAssociationEndpoint(assocRel, bindings)
	if endpoint == nil {
		return NewEvalError(
			"association-class member %s requires exactly one association row; found %d",
			assocRel.LinkClassMember(),
			endpointCount,
		)
	}

	link, ok := assocRel.LinkForEndpoint(endpoint)
	if !ok {
		return NewEvalError("missing association-class row for endpoint")
	}
	return NewEvalResult(link)
}

func soleAssociationEndpoint(assocRel *object.AssociationRelation, bindings *Bindings) *object.Record {
	if assocRel.Endpoints().Size() == 1 {
		elem := assocRel.Endpoints().Elements()[0]
		record, ok := elem.(*object.Record)
		if !ok {
			return nil
		}
		return record
	}

	candidates := boundEndpointsInScope(assocRel, bindings)
	if len(candidates) == 1 {
		return candidates[0]
	}
	return nil
}

func boundEndpointsInScope(assocRel *object.AssociationRelation, bindings *Bindings) []*object.Record {
	var candidates []*object.Record
	seen := make(map[*object.Record]struct{})

	for scope := bindings; scope != nil; scope = scope.outer {
		for _, entry := range scope.store {
			if entry.Namespace != NamespaceLocal {
				continue
			}
			record, ok := entry.Value.(*object.Record)
			if !ok || !assocRel.Endpoints().Contains(record) {
				continue
			}
			if _, dup := seen[record]; dup {
				continue
			}
			seen[record] = struct{}{}
			candidates = append(candidates, record)
		}
	}
	return candidates
}
