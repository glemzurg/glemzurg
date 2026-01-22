package parser

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// scopeObjectKeys transforms object key references in a steps data structure (map[string]any)
// from short keys (like "bob") to fully qualified scenario object keys.
// This operates on the raw YAML data before it gets parsed into Node objects,
// ensuring that Node objects are always well-formed with complete keys.
func scopeObjectKeys(scenarioKey identity.Key, data map[string]any) error {
	return scopeObjectKeysRecursive(scenarioKey, data)
}

func scopeObjectKeysRecursive(scenarioKey identity.Key, data map[string]any) error {
	// Handle from_object_key
	if fromKey, ok := data["from_object_key"]; ok {
		if fromKeyStr, ok := fromKey.(string); ok {
			fullKey, err := identity.NewScenarioObjectKey(scenarioKey, fromKeyStr)
			if err != nil {
				return err
			}
			data["from_object_key"] = fullKey.String()
		}
	}

	// Handle to_object_key
	if toKey, ok := data["to_object_key"]; ok {
		if toKeyStr, ok := toKey.(string); ok {
			fullKey, err := identity.NewScenarioObjectKey(scenarioKey, toKeyStr)
			if err != nil {
				return err
			}
			data["to_object_key"] = fullKey.String()
		}
	}

	// Recursively process statements
	if statements, ok := data["statements"]; ok {
		if stmtSlice, ok := statements.([]any); ok {
			for _, stmt := range stmtSlice {
				if stmtMap, ok := stmt.(map[string]any); ok {
					if err := scopeObjectKeysRecursive(scenarioKey, stmtMap); err != nil {
						return err
					}
				}
			}
		}
	}

	// Recursively process cases
	if cases, ok := data["cases"]; ok {
		if casesSlice, ok := cases.([]any); ok {
			for _, c := range casesSlice {
				if caseMap, ok := c.(map[string]any); ok {
					// Process statements within each case
					if statements, ok := caseMap["statements"]; ok {
						if stmtSlice, ok := statements.([]any); ok {
							for _, stmt := range stmtSlice {
								if stmtMap, ok := stmt.(map[string]any); ok {
									if err := scopeObjectKeysRecursive(scenarioKey, stmtMap); err != nil {
										return err
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}
