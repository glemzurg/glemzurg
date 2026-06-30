package model_class

import (
	"strings"
	"unicode"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ComputedSimpleActionGuaranteeDescription derives "Set <attribute name>" for a
// state_change guarantee that assigns a single-word expression to a class attribute.
func ComputedSimpleActionGuaranteeDescription(guarantee model_logic.Logic, attributes []Attribute) (string, bool) {
	if guarantee.Type != model_logic.LogicTypeStateChange {
		return "", false
	}
	if guarantee.Target == "" || !isSingleWordSpecification(guarantee.Spec.Specification) {
		return "", false
	}
	attrName := attributeDisplayNameForTarget(attributes, guarantee.Target)
	if attrName == "" {
		return "", false
	}
	return "Set " + attrName, true
}

// ComputedAssociationSetAddGuaranteeDescription derives "Add to <association name>"
// for a state_change guarantee that unions {_new(...)} onto an outgoing association field.
func ComputedAssociationSetAddGuaranteeDescription(guarantee model_logic.Logic, associations map[identity.Key]Association) (string, bool) {
	if guarantee.Type != model_logic.LogicTypeStateChange || guarantee.Target == "" {
		return "", false
	}
	spec := strings.TrimSpace(guarantee.Spec.Specification)
	if !IsAssociationSetAddSpecification(spec) {
		return "", false
	}
	for _, assoc := range associations {
		if AssociationTLAFieldName(assoc.Name) != guarantee.Target {
			continue
		}
		if spec != guarantee.Target && !strings.HasPrefix(spec, guarantee.Target) {
			return "", false
		}
		return "Add to " + assoc.Name, true
	}
	return "", false
}

// ComputedActionGuaranteeDescription returns a human description for a guarantee when
// the authored TLA matches a known shorthand pattern.
func ComputedActionGuaranteeDescription(guarantee model_logic.Logic, attributes []Attribute, associations map[identity.Key]Association) (string, bool) {
	if desc, ok := ComputedSimpleActionGuaranteeDescription(guarantee, attributes); ok {
		return desc, true
	}
	if desc, ok := ComputedAssociationSetMapGuaranteeDescription(guarantee, associations); ok {
		return desc, true
	}
	if desc, ok := ComputedAssociationDestroyGuaranteeDescription(guarantee, associations); ok {
		return desc, true
	}
	return ComputedAssociationSetAddGuaranteeDescription(guarantee, associations)
}

// ComputedAssociationDestroyGuaranteeDescription derives "Set <association name>"
// for a destroy guarantee on an outgoing association field, matching attribute-set display.
func ComputedAssociationDestroyGuaranteeDescription(guarantee model_logic.Logic, associations map[identity.Key]Association) (string, bool) {
	if guarantee.Type != model_logic.LogicTypeDestroy || guarantee.Target == "" {
		return "", false
	}
	assocName := associationDisplayNameForTarget(associations, guarantee.Target)
	if assocName == "" {
		return "", false
	}
	return "Set " + assocName, true
}

// ComputedAssociationSetMapGuaranteeDescription derives "Update <association name>"
// for a state_change guarantee that maps a peer event over an outgoing association field.
func ComputedAssociationSetMapGuaranteeDescription(guarantee model_logic.Logic, associations map[identity.Key]Association) (string, bool) {
	if guarantee.Type != model_logic.LogicTypeStateChange || guarantee.Target == "" {
		return "", false
	}
	spec := strings.TrimSpace(guarantee.Spec.Specification)
	if !IsAssociationSetMapSpecification(spec) && !IsAssociationAddOrUpdateSpecification(spec) {
		return "", false
	}
	for _, assoc := range associations {
		if AssociationTLAFieldName(assoc.Name) != guarantee.Target {
			continue
		}
		return "Update " + assoc.Name, true
	}
	return "", false
}

func isAssociationSetAddSpecification(specification string) bool {
	if specification == "" {
		return false
	}
	lower := strings.ToLower(specification)
	hasUnion := strings.Contains(lower, `\union`) || strings.Contains(specification, "∪")
	hasNew := strings.Contains(specification, `_new(`) || strings.Contains(specification, model_state.EventTLANameNew+`(`)
	return hasUnion && hasNew
}

func associationDisplayNameForTarget(associations map[identity.Key]Association, target string) string {
	for _, assoc := range associations {
		if AssociationTLAFieldName(assoc.Name) != target {
			continue
		}
		if assoc.Name != "" {
			return assoc.Name
		}
	}
	return ""
}

func attributeDisplayNameForTarget(attributes []Attribute, target string) string {
	normalizedTarget := identity.NormalizeSubKey(target)
	for _, attr := range attributes {
		if attr.Key.SubKey == normalizedTarget && attr.Name != "" {
			return attr.Name
		}
	}
	return ""
}

func isSingleWordSpecification(specification string) bool {
	specification = strings.TrimSpace(specification)
	if specification == "" {
		return false
	}
	for i, r := range specification {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
