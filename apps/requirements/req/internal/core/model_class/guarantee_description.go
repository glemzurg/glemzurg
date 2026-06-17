package model_class

import (
	"strings"
	"unicode"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
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
