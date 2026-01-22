package parser

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// scopeObjectKeys transforms key references in a steps data structure (map[string]any)
// from short keys to fully qualified keys.
// - Object keys (like "bob") become fully qualified scenario object keys.
// - Attribute keys (like "subdomain/class/attr") become fully qualified attribute keys.
// This operates on the raw YAML data before it gets parsed into Node objects,
// ensuring that Node objects are always well-formed with complete keys.
func scopeObjectKeys(scenarioKey identity.Key, subdomainKey identity.Key, data map[string]any) error {
	return scopeObjectKeysRecursive(scenarioKey, subdomainKey, data)
}

func scopeObjectKeysRecursive(scenarioKey identity.Key, subdomainKey identity.Key, data map[string]any) error {
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

	// Handle attribute_key (format: "subdomain/class/attribute" or "class/attribute")
	if attrKey, ok := data["attribute_key"]; ok {
		if attrKeyStr, ok := attrKey.(string); ok {
			fullKey, err := expandAttributeKey(subdomainKey, attrKeyStr)
			if err != nil {
				return err
			}
			data["attribute_key"] = fullKey.String()
		}
	}

	// Recursively process statements
	if statements, ok := data["statements"]; ok {
		if stmtSlice, ok := statements.([]any); ok {
			for _, stmt := range stmtSlice {
				if stmtMap, ok := stmt.(map[string]any); ok {
					if err := scopeObjectKeysRecursive(scenarioKey, subdomainKey, stmtMap); err != nil {
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
									if err := scopeObjectKeysRecursive(scenarioKey, subdomainKey, stmtMap); err != nil {
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

// expandAttributeKey converts a short attribute key format to a full identity key.
// Supported formats:
// - "class/attribute" - uses the provided subdomain
// - "subdomain/class/attribute" - constructs subdomain from domain + subdomain part
func expandAttributeKey(subdomainKey identity.Key, shortKey string) (identity.Key, error) {
	parts := strings.Split(shortKey, "/")

	switch len(parts) {
	case 2:
		// Format: "class/attribute"
		classSubKey := parts[0]
		attrSubKey := parts[1]

		classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for attribute: %s", shortKey)
		}
		return identity.NewAttributeKey(classKey, attrSubKey)

	case 3:
		// Format: "subdomain/class/attribute"
		// The subdomain part is relative to the domain of the current subdomain.
		subdomainSubKey := parts[0]
		classSubKey := parts[1]
		attrSubKey := parts[2]

		// Get the domain key from the current subdomain key.
		domainKeyStr := subdomainKey.ParentKey()
		if domainKeyStr == "" {
			return identity.Key{}, errors.Errorf("could not get domain key from subdomain for attribute: %s", shortKey)
		}
		domainKey, err := identity.ParseKey(domainKeyStr)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to parse domain key for attribute: %s", shortKey)
		}

		// Build the target subdomain key.
		targetSubdomainKey, err := identity.NewSubdomainKey(domainKey, subdomainSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create subdomain key for attribute: %s", shortKey)
		}

		classKey, err := identity.NewClassKey(targetSubdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for attribute: %s", shortKey)
		}
		return identity.NewAttributeKey(classKey, attrSubKey)

	default:
		return identity.Key{}, errors.Errorf("invalid attribute key format '%s': expected 'class/attribute' or 'subdomain/class/attribute'", shortKey)
	}
}
