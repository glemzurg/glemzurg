package parser_human

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
	// Scope object keys (from_object_key, to_object_key).
	if err := scopeObjectKeyField(data, "from_object_key", scenarioKey); err != nil {
		return err
	}
	if err := scopeObjectKeyField(data, "to_object_key", scenarioKey); err != nil {
		return err
	}

	// Scope domain-relative keys (attribute_key, event_key, query_key, scenario_key).
	if err := scopeExpandableKeyField(data, "attribute_key", func(s string) (identity.Key, error) {
		return expandAttributeKey(subdomainKey, s)
	}); err != nil {
		return err
	}
	if err := scopeExpandableKeyField(data, "event_key", func(s string) (identity.Key, error) {
		return expandEventKey(subdomainKey, s)
	}); err != nil {
		return err
	}
	if err := scopeExpandableKeyField(data, "query_key", func(s string) (identity.Key, error) {
		return expandQueryKey(subdomainKey, s)
	}); err != nil {
		return err
	}
	if err := scopeExpandableKeyField(data, "scenario_key", func(s string) (identity.Key, error) {
		return expandScenarioKey(subdomainKey, scenarioKey, s)
	}); err != nil {
		return err
	}

	// Recursively process statements.
	return scopeStatements(scenarioKey, subdomainKey, data)
}

// scopeObjectKeyField expands a short object key field to a fully qualified scenario object key.
func scopeObjectKeyField(data map[string]any, field string, scenarioKey identity.Key) error {
	val, ok := data[field]
	if !ok {
		return nil
	}
	valStr, ok := val.(string)
	if !ok {
		return nil
	}
	fullKey, err := identity.NewScenarioObjectKey(scenarioKey, identity.NormalizeSubKey(valStr))
	if err != nil {
		return err
	}
	data[field] = fullKey.String()
	return nil
}

// scopeExpandableKeyField expands a short key field using the provided expand function.
func scopeExpandableKeyField(data map[string]any, field string, expand func(string) (identity.Key, error)) error {
	val, ok := data[field]
	if !ok {
		return nil
	}
	valStr, ok := val.(string)
	if !ok {
		return nil
	}
	fullKey, err := expand(valStr)
	if err != nil {
		return err
	}
	data[field] = fullKey.String()
	return nil
}

// scopeStatements recursively processes the statements field.
func scopeStatements(scenarioKey, subdomainKey identity.Key, data map[string]any) error {
	statements, ok := data["statements"]
	if !ok {
		return nil
	}
	stmtSlice, ok := statements.([]any)
	if !ok {
		return nil
	}
	for _, stmt := range stmtSlice {
		if stmtMap, ok := stmt.(map[string]any); ok {
			if err := scopeObjectKeysRecursive(scenarioKey, subdomainKey, stmtMap); err != nil {
				return err
			}
		}
	}
	return nil
}

// expandAttributeKey converts a short attribute key format to a full identity key.
// Supported formats:
// - "class/attribute" - uses the provided subdomain
// - "domainFolder/class/attribute" - the domainFolder is ignored, uses the current subdomain.
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
// - "useCase/scenario/scenarioName" - explicit form.
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
// - "class/event/eventname" - explicit form.
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
// - "class/query/queryname" - explicit form.
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
