package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
)

// PopulateDerivedAttributeCallersFromModel records which classes reference each
// derived attribute in lowered logic. External derived attributes (no simulatable
// in-scope caller) belong on the simulation surface, like queries.
func PopulateDerivedAttributeCallersFromModel(model *core.Model, catalog *ClassCatalog) {
	derivedKeys := buildDerivedAttributeKeyIndex(model)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				recordClassDerivedAttributeCallers(class, derivedKeys, catalog)
			}
		}
	}
}

func buildDerivedAttributeKeyIndex(model *core.Model) map[identity.Key]bool {
	derivedKeys := make(map[identity.Key]bool)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				for _, attr := range class.Attributes {
					if attr.DerivationPolicy != nil {
						derivedKeys[attr.Key] = true
					}
				}
			}
		}
	}
	return derivedKeys
}

func recordClassDerivedAttributeCallers(
	class model_class.Class,
	derivedKeys map[identity.Key]bool,
	catalog *ClassCatalog,
) {
	callerClassKey := class.Key

	recordLogicsDerivedAttributeCallers(catalog, callerClassKey, class.Invariants, derivedKeys)
	for _, guard := range class.Guards {
		recordLogicDerivedAttributeCallers(catalog, callerClassKey, guard.Logic, derivedKeys)
	}
	for _, action := range class.Actions {
		recordLogicsDerivedAttributeCallers(catalog, callerClassKey, action.Requires, derivedKeys)
		recordLogicsDerivedAttributeCallers(catalog, callerClassKey, action.Guarantees, derivedKeys)
		recordLogicsDerivedAttributeCallers(catalog, callerClassKey, action.SafetyRules, derivedKeys)
	}
	for _, query := range class.Queries {
		recordLogicsDerivedAttributeCallers(catalog, callerClassKey, query.Requires, derivedKeys)
		recordLogicsDerivedAttributeCallers(catalog, callerClassKey, query.Guarantees, derivedKeys)
	}
	for _, attr := range class.Attributes {
		if attr.DerivationPolicy != nil {
			recordLogicDerivedAttributeCallers(catalog, callerClassKey, *attr.DerivationPolicy, derivedKeys)
		}
		recordLogicsDerivedAttributeCallers(catalog, callerClassKey, attr.Invariants, derivedKeys)
	}
}

func recordLogicsDerivedAttributeCallers(
	catalog *ClassCatalog,
	callerClassKey identity.Key,
	logics []model_logic.Logic,
	derivedKeys map[identity.Key]bool,
) {
	for _, logic := range logics {
		recordLogicDerivedAttributeCallers(catalog, callerClassKey, logic, derivedKeys)
	}
}

func recordLogicDerivedAttributeCallers(
	catalog *ClassCatalog,
	callerClassKey identity.Key,
	logic model_logic.Logic,
	derivedKeys map[identity.Key]bool,
) {
	expr := logic.Spec.Expression
	if expr == nil {
		return
	}
	for attrKey := range model_bridge.CollectAttributeRefs(expr) {
		if derivedKeys[attrKey] {
			catalog.addAttributeCaller(attrKey, callerClassKey)
		}
	}
}
