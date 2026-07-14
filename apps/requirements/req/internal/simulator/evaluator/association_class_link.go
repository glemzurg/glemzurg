package evaluator

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func evalAssociationRelationFieldAccess(
	assocRel *object.AssociationRelation,
	field string,
	bindings *Bindings,
) *EvalResult {
	// Step 1: association-class member name → sole/binder scalar AC row, else set of all link rows.
	if field == assocRel.LinkClassMember() {
		if sole := soleAssociationEndpoint(assocRel, bindings); sole != nil {
			link, ok := assocRel.LinkForEndpoint(sole)
			if !ok {
				return NewEvalError("missing association-class row for endpoint")
			}
			return NewEvalResult(link)
		}
		return NewEvalResult(associationClassLinkSet(assocRel))
	}

	// Step 2: project a field present on every endpoint (binder/sole → scalar).
	if projected, ok, err := projectAssociationEndpointField(assocRel, field, bindings); ok {
		return projected
	} else if err != nil {
		return err
	}

	// Step 3: project a field present on every AC link row (binder/sole → scalar).
	if projected, ok, err := projectAssociationLinkField(assocRel, field, bindings); ok {
		return projected
	} else if err != nil {
		return err
	}

	// Step 4: clear failure.
	return NewEvalError("%s", associationFieldProjectionError(assocRel, field))
}

// associationClassLinkSet returns all association-class rows as a set (empty if none).
func associationClassLinkSet(assocRel *object.AssociationRelation) *object.Set {
	set := object.NewSet()
	for _, link := range assocRel.LinkByEndpoint() {
		set.Add(link)
	}
	return set
}

// projectAssociationEndpointField projects field over endpoints when every endpoint has it.
// ok=false means "try next step"; err is a hard failure.
func projectAssociationEndpointField(
	assocRel *object.AssociationRelation,
	field string,
	bindings *Bindings,
) (result *EvalResult, ok bool, err *EvalResult) {
	endpoints := associationEndpointRecords(assocRel)
	if len(endpoints) == 0 {
		return nil, false, NewEvalError("no endpoints to project field %s on association navigation", field)
	}
	if !everyRecordHasField(endpoints, field) {
		return nil, false, nil
	}
	if sole := soleAssociationEndpoint(assocRel, bindings); sole != nil {
		value, ok := object.RecordField(sole, field)
		if !ok {
			return nil, false, NewEvalError("field not found: %s", field)
		}
		return NewEvalResult(value), true, nil
	}
	return NewEvalResult(projectRecordsField(endpoints, field)), true, nil
}

// projectAssociationLinkField projects field over AC link rows when every link has it.
func projectAssociationLinkField(
	assocRel *object.AssociationRelation,
	field string,
	bindings *Bindings,
) (result *EvalResult, ok bool, err *EvalResult) {
	if assocRel.LinkClassMember() == "" {
		return nil, false, nil
	}
	links := associationLinkRecords(assocRel)
	if len(links) == 0 {
		return nil, false, nil
	}
	if !everyRecordHasField(links, field) {
		return nil, false, nil
	}
	if sole := soleAssociationEndpoint(assocRel, bindings); sole != nil {
		link, found := assocRel.LinkForEndpoint(sole)
		if !found {
			return nil, false, NewEvalError("missing association-class row for endpoint")
		}
		value, ok := object.RecordField(link, field)
		if !ok {
			return nil, false, NewEvalError("field not found: %s", field)
		}
		return NewEvalResult(value), true, nil
	}
	return NewEvalResult(projectRecordsField(links, field)), true, nil
}

func associationEndpointRecords(assocRel *object.AssociationRelation) []*object.Record {
	var records []*object.Record
	for _, elem := range assocRel.Endpoints().Elements() {
		rec, ok := elem.(*object.Record)
		if !ok {
			continue
		}
		records = append(records, rec)
	}
	return records
}

func associationLinkRecords(assocRel *object.AssociationRelation) []*object.Record {
	links := make([]*object.Record, 0, len(assocRel.LinkByEndpoint()))
	for _, link := range assocRel.LinkByEndpoint() {
		links = append(links, link)
	}
	return links
}

func everyRecordHasField(records []*object.Record, field string) bool {
	for _, rec := range records {
		if !object.RecordHasField(rec, field) {
			return false
		}
	}
	return true
}

func projectRecordsField(records []*object.Record, field string) *object.Set {
	set := object.NewSet()
	for _, rec := range records {
		value, ok := object.RecordField(rec, field)
		if !ok {
			continue
		}
		set.Add(value)
	}
	return set
}

func associationFieldProjectionError(assocRel *object.AssociationRelation, field string) string {
	endpoints := associationEndpointRecords(assocRel)
	links := associationLinkRecords(assocRel)
	endpointHas := 0
	for _, ep := range endpoints {
		if object.RecordHasField(ep, field) {
			endpointHas++
		}
	}
	linkHas := 0
	for _, link := range links {
		if object.RecordHasField(link, field) {
			linkHas++
		}
	}
	return fmt.Sprintf(
		"cannot project field %s on association navigation (LinkClassMember=%q, endpoints=%d with field=%d, linkRows=%d with field=%d)",
		field,
		assocRel.LinkClassMember(),
		len(endpoints),
		endpointHas,
		len(links),
		linkHas,
	)
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
