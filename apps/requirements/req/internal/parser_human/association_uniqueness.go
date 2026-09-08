package parser_human

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

func uniquenessFromYaml(
	fromClassKey, toClassKey identity.Key,
	associationData map[string]any,
) ([]model_class.AssociationUniqueness, error) {
	if _, found := associationData["uniqueness_constraints"]; found {
		return nil, errors.Errorf("uniqueness_constraints is no longer supported; use uniqueness")
	}

	raw, found := associationData["uniqueness"]
	if !found || raw == nil {
		return nil, nil
	}
	seq, ok := raw.([]any)
	if !ok {
		return nil, errors.Errorf("uniqueness must be a sequence of constraints")
	}
	if len(seq) == 0 {
		return nil, nil
	}

	result := make([]model_class.AssociationUniqueness, 0, len(seq))
	for i, item := range seq {
		mapping, ok := item.(map[string]any)
		if !ok {
			return nil, errors.Errorf("uniqueness[%d] must be a mapping", i)
		}
		constraint, err := uniquenessConstraintFromYaml(fromClassKey, toClassKey, i, mapping)
		if err != nil {
			return nil, err
		}
		result = append(result, constraint)
	}
	return result, nil
}

func uniquenessConstraintFromYaml(
	fromClassKey, toClassKey identity.Key,
	index int,
	mapping map[string]any,
) (model_class.AssociationUniqueness, error) {
	fromField := fmt.Sprintf("uniqueness[%d].from_attributes", index)
	toField := fmt.Sprintf("uniqueness[%d].to_attributes", index)
	fromAttrs, err := yamlStringSlice(mapping, "from_attributes")
	if err != nil {
		return model_class.AssociationUniqueness{}, err
	}
	toAttrs, err := yamlStringSlice(mapping, "to_attributes")
	if err != nil {
		return model_class.AssociationUniqueness{}, err
	}
	fromKeys, err := attributeKeysFromYamlSubKeys(fromClassKey, fromAttrs, fromField)
	if err != nil {
		return model_class.AssociationUniqueness{}, err
	}
	toKeys, err := attributeKeysFromYamlSubKeys(toClassKey, toAttrs, toField)
	if err != nil {
		return model_class.AssociationUniqueness{}, err
	}
	uniqueness := model_class.NewAssociationUniqueness(fromKeys, toKeys)
	ctx := coreerr.NewContext("association", fmt.Sprintf("uniqueness[%d]", index))
	if err := uniqueness.Validate(ctx); err != nil {
		return model_class.AssociationUniqueness{}, errors.WithStack(err)
	}
	return uniqueness, nil
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

func generateAssociationUniquenessYaml(builder *YamlBuilder, uniqueness []model_class.AssociationUniqueness) {
	if len(uniqueness) == 0 {
		return
	}
	var items []*YamlBuilder
	for _, constraint := range uniqueness {
		item := NewYamlBuilder()
		if fromAttrs := attributeSubKeysFromKeys(constraint.FromAttributeKeys); len(fromAttrs) > 0 {
			item.AddSequenceField("from_attributes", fromAttrs)
		}
		if toAttrs := attributeSubKeysFromKeys(constraint.ToAttributeKeys); len(toAttrs) > 0 {
			item.AddSequenceField("to_attributes", toAttrs)
		}
		items = append(items, item)
	}
	builder.AddSequenceOfMappings("uniqueness", items)
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
