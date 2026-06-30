package parser_human

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

func uniquenessFromYaml(
	fromClassKey, toClassKey identity.Key,
	associationData map[string]any,
) (*model_class.AssociationUniqueness, error) {
	if _, found := associationData["uniqueness_constraints"]; found {
		return nil, errors.Errorf("uniqueness_constraints is no longer supported; use uniqueness")
	}

	mapping, ok := associationData["uniqueness"].(map[string]any)
	if !ok {
		return nil, nil //nolint:nilnil // optional field, absence is not an error
	}

	fromAttrs, err := yamlStringSlice(mapping, "from_attributes")
	if err != nil {
		return nil, err
	}
	toAttrs, err := yamlStringSlice(mapping, "to_attributes")
	if err != nil {
		return nil, err
	}
	if len(fromAttrs) == 0 && len(toAttrs) == 0 {
		return nil, nil //nolint:nilnil // empty mapping means no uniqueness constraint
	}
	fromKeys, err := attributeKeysFromYamlSubKeys(fromClassKey, fromAttrs, "uniqueness.from_attributes")
	if err != nil {
		return nil, err
	}
	toKeys, err := attributeKeysFromYamlSubKeys(toClassKey, toAttrs, "uniqueness.to_attributes")
	if err != nil {
		return nil, err
	}
	uniqueness := model_class.NewAssociationUniqueness(fromKeys, toKeys)
	ctx := coreerr.NewContext("association", "uniqueness")
	if err := uniqueness.Validate(ctx); err != nil {
		return nil, errors.WithStack(err)
	}
	return &uniqueness, nil
}

func attributeKeysFromYamlSubKeys(classKey identity.Key, subKeys []string, field string) ([]identity.Key, error) {
	if len(subKeys) == 0 {
		return nil, nil
	}
	keys := make([]identity.Key, 0, len(subKeys))
	for j, subKey := range subKeys {
		attrKey, err := identity.NewAttributeKey(classKey, subKey)
		if err != nil {
			return nil, errors.Errorf("%s[%d]: %s", field, j, err.Error())
		}
		keys = append(keys, attrKey)
	}
	return keys, nil
}

func yamlStringSlice(data map[string]any, field string) ([]string, error) {
	raw, ok := data[field]
	if !ok || raw == nil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case []any:
		result := make([]string, 0, len(typed))
		for j, item := range typed {
			str, ok := item.(string)
			if !ok {
				return nil, errors.Errorf("%s[%d] must be a string", field, j)
			}
			result = append(result, str)
		}
		return result, nil
	case []string:
		return append([]string(nil), typed...), nil
	default:
		return nil, errors.Errorf("%s must be a string sequence", field)
	}
}

func generateAssociationUniquenessYaml(builder *YamlBuilder, uniqueness *model_class.AssociationUniqueness) {
	if uniqueness == nil {
		return
	}
	uniquenessBuilder := NewYamlBuilder()
	if fromAttrs := attributeSubKeysFromKeys(uniqueness.FromAttributeKeys); len(fromAttrs) > 0 {
		uniquenessBuilder.AddSequenceField("from_attributes", fromAttrs)
	}
	if toAttrs := attributeSubKeysFromKeys(uniqueness.ToAttributeKeys); len(toAttrs) > 0 {
		uniquenessBuilder.AddSequenceField("to_attributes", toAttrs)
	}
	builder.AddMappingField("uniqueness", uniquenessBuilder)
}

func attributeSubKeysFromKeys(keys []identity.Key) []string {
	if len(keys) == 0 {
		return nil
	}
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.SubKey
	}
	return result
}
