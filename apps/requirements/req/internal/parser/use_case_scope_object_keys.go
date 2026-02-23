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

	// Handle event_key (format: "subdomain/class/event/eventname" or "class/event/eventname")
	if eventKey, ok := data["event_key"]; ok {
		if eventKeyStr, ok := eventKey.(string); ok {
			fullKey, err := expandEventKey(subdomainKey, eventKeyStr)
			if err != nil {
				return err
			}
			data["event_key"] = fullKey.String()
		}
	}

	// Handle query_key (format: "class/query/queryname" or "domainFolder/class/query/queryname")
	if queryKey, ok := data["query_key"]; ok {
		if queryKeyStr, ok := queryKey.(string); ok {
			fullKey, err := expandQueryKey(subdomainKey, queryKeyStr)
			if err != nil {
				return err
			}
			data["query_key"] = fullKey.String()
		}
	}

	// Handle scenario_key (format: "scenarioName" for same use case, or "useCase/scenario/scenarioName")
	if scenarioKeyVal, ok := data["scenario_key"]; ok {
		if scenarioKeyStr, ok := scenarioKeyVal.(string); ok {
			fullKey, err := expandScenarioKey(subdomainKey, scenarioKey, scenarioKeyStr)
			if err != nil {
				return err
			}
			data["scenario_key"] = fullKey.String()
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

	return nil
}

// expandAttributeKey converts a short attribute key format to a full identity key.
// Supported formats:
// - "class/attribute" - uses the provided subdomain
// - "domainFolder/class/attribute" - the domainFolder is ignored, uses the current subdomain
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
		// Format: "domainFolder/class/attribute"
		// The domainFolder prefix is informational only - we use the current subdomain.
		// This matches the yaml format used in web_books model (e.g., "01_order_fulfillment/title/title").
		// domainFolder := parts[0] // Ignored
		classSubKey := parts[1]
		attrSubKey := parts[2]

		classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for attribute: %s", shortKey)
		}
		return identity.NewAttributeKey(classKey, attrSubKey)

	default:
		return identity.Key{}, errors.Errorf("invalid attribute key format '%s': expected 'class/attribute' or 'domainFolder/class/attribute'", shortKey)
	}
}

// expandScenarioKey converts a short scenario key format to a full identity key.
// The scenarioKey provides context for the current scenario (its parent use case is used
// when only a bare scenario name is given).
// Supported formats:
// - "scenarioName" - same use case as the current scenario (preferred)
// - "useCase/scenario/scenarioName" - explicit form
func expandScenarioKey(subdomainKey identity.Key, scenarioKey identity.Key, shortKey string) (identity.Key, error) {
	parts := strings.Split(shortKey, "/")

	switch len(parts) {
	case 1:
		// Format: "scenarioName" - same use case as the current scenario.
		useCaseKey, err := identity.ParseKey(scenarioKey.ParentKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to get use case key from scenario key: %s", scenarioKey.String())
		}
		return identity.NewScenarioKey(useCaseKey, parts[0])

	case 3:
		// Format: "useCase/scenario/scenarioName"
		useCaseSubKey := parts[0]
		// parts[1] should be "scenario"
		scenarioSubKey := parts[2]

		useCaseKey, err := identity.NewUseCaseKey(subdomainKey, useCaseSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create use case key for scenario: %s", shortKey)
		}
		return identity.NewScenarioKey(useCaseKey, scenarioSubKey)

	default:
		return identity.Key{}, errors.Errorf("invalid scenario key format '%s': expected 'scenarioName' or 'useCase/scenario/scenarioName'", shortKey)
	}
}

// expandEventKey converts a short event key format to a full identity key.
// Supported formats:
// - "class/eventname" - compact form (preferred)
// - "class/event/eventname" - explicit form
func expandEventKey(subdomainKey identity.Key, shortKey string) (identity.Key, error) {
	parts := strings.Split(shortKey, "/")

	switch len(parts) {
	case 2:
		// Format: "class/eventname"
		classSubKey := parts[0]
		eventSubKey := parts[1]

		classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for event: %s", shortKey)
		}
		return identity.NewEventKey(classKey, eventSubKey)

	case 3:
		// Format: "class/event/eventname"
		classSubKey := parts[0]
		// parts[1] should be "event"
		eventSubKey := parts[2]

		classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for event: %s", shortKey)
		}
		return identity.NewEventKey(classKey, eventSubKey)

	default:
		return identity.Key{}, errors.Errorf("invalid event key format '%s': expected 'class/eventname' or 'class/event/eventname'", shortKey)
	}
}

// expandQueryKey converts a short query key format to a full identity key.
// Supported formats:
// - "class/queryname" - compact form (preferred)
// - "class/query/queryname" - explicit form
func expandQueryKey(subdomainKey identity.Key, shortKey string) (identity.Key, error) {
	parts := strings.Split(shortKey, "/")

	switch len(parts) {
	case 2:
		// Format: "class/queryname"
		classSubKey := parts[0]
		querySubKey := parts[1]

		classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for query: %s", shortKey)
		}
		return identity.NewQueryKey(classKey, querySubKey)

	case 3:
		// Format: "class/query/queryname"
		classSubKey := parts[0]
		// parts[1] should be "query"
		querySubKey := parts[2]

		classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
		if err != nil {
			return identity.Key{}, errors.Wrapf(err, "failed to create class key for query: %s", shortKey)
		}
		return identity.NewQueryKey(classKey, querySubKey)

	default:
		return identity.Key{}, errors.Errorf("invalid query key format '%s': expected 'class/queryname' or 'class/query/queryname'", shortKey)
	}
}
