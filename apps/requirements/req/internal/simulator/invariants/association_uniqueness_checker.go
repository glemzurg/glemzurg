package invariants

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

type associationUniquenessBinding struct {
	association model_class.Association
	uniqueness  model_class.AssociationUniqueness
}

// AssociationUniquenessChecker validates association uniqueness on link attribute tuples.
type AssociationUniquenessChecker struct {
	bindings []associationUniquenessBinding
}

// NewAssociationUniquenessChecker builds association uniqueness metadata from schema.
func NewAssociationUniquenessChecker(sch *schema.Schema) *AssociationUniquenessChecker {
	checker := &AssociationUniquenessChecker{}

	sch.ForEachAssociation(func(assoc model_class.Association) {
		if assoc.Uniqueness == nil {
			return
		}
		if !sch.IsClassInScope(assoc.FromClassKey) || !sch.IsClassInScope(assoc.ToClassKey) {
			return
		}
		checker.bindings = append(checker.bindings, associationUniquenessBinding{
			association: assoc,
			uniqueness:  *assoc.Uniqueness,
		})
	})

	return checker
}

// CheckState validates all association uniqueness rules.
func (c *AssociationUniquenessChecker) CheckState(simState *instance.State) ViolationErrors {
	var violations ViolationErrors
	for _, binding := range c.bindings {
		violations = append(violations, c.checkBinding(simState, binding)...)
	}
	return violations
}

func (c *AssociationUniquenessChecker) checkBinding(
	simState *instance.State,
	binding associationUniquenessBinding,
) ViolationErrors {
	links := collectAssociationLinks(simState, binding.association)
	if len(links) == 0 {
		return nil
	}

	partitions := make(map[string][]associationLinkEndpoints)
	for _, link := range links {
		partitionKey := associationUniquenessPartitionKey(binding.uniqueness, link)
		partitions[partitionKey] = append(partitions[partitionKey], link)
	}

	var violations ViolationErrors
	for _, partitionLinks := range partitions {
		counts := make(map[string]int)
		for _, link := range partitionLinks {
			fromInst := simState.GetInstance(link.fromID)
			toInst := simState.GetInstance(link.toID)
			if fromInst == nil || toInst == nil {
				continue
			}
			tupleKey := associationUniquenessTupleKey(fromInst, toInst, binding.uniqueness)
			counts[tupleKey]++
			if counts[tupleKey] > 1 {
				violations = append(violations, NewAssociationUniquenessViolation(AssociationUniquenessViolationParams{
					AssociationName: binding.association.Name,
					FromInstanceID:  link.fromID,
					ToInstanceID:    link.toID,
					ActualCount:     counts[tupleKey],
					RequiredMin:     0,
					RequiredMax:     1,
					Message: fmt.Sprintf(
						"association %q exceeds uniqueness constraint",
						binding.association.Name,
					),
				}))
			}
		}
	}
	return violations
}

type associationLinkEndpoints struct {
	fromID instance.ID
	toID   instance.ID
}

func collectAssociationLinks(
	simState *instance.State,
	assoc model_class.Association,
) []associationLinkEndpoints {
	var links []associationLinkEndpoints
	if assoc.AssociationClassKey != nil {
		simState.ForEachAssociationLinkOfHost(assoc.Key, func(link instance.AssociationLink) {
			links = append(links, associationLinkEndpoints{
				fromID: link.FromEndpointID,
				toID:   link.ToEndpointID,
			})
		})
		return links
	}

	simState.ForEachBinaryLinkOfAssociation(assoc.Key, func(fromID, toID instance.ID) {
		links = append(links, associationLinkEndpoints{
			fromID: fromID,
			toID:   toID,
		})
	})
	return links
}

func associationUniquenessPartitionKey(uniqueness model_class.AssociationUniqueness, link associationLinkEndpoints) string {
	switch {
	case len(uniqueness.FromAttributeKeys) > 0 && len(uniqueness.ToAttributeKeys) > 0:
		return "global"
	case len(uniqueness.ToAttributeKeys) > 0:
		return fmt.Sprintf("from:%d", link.fromID)
	case len(uniqueness.FromAttributeKeys) > 0:
		return fmt.Sprintf("to:%d", link.toID)
	default:
		return "global"
	}
}

func associationUniquenessTupleKey(
	fromInst, toInst *instance.Instance,
	uniqueness model_class.AssociationUniqueness,
) string {
	parts := make([]string, 0, len(uniqueness.FromAttributeKeys)+len(uniqueness.ToAttributeKeys))
	for _, attrKey := range uniqueness.FromAttributeKeys {
		parts = append(parts, indexTupleValueKey(fromInst.GetAttribute(attrKey.SubKey)))
	}
	for _, attrKey := range uniqueness.ToAttributeKeys {
		parts = append(parts, indexTupleValueKey(toInst.GetAttribute(attrKey.SubKey)))
	}
	return strings.Join(parts, "\x00")
}
