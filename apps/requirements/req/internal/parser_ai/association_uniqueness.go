package parser_ai

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type inputAssociationUniqueness struct {
	FromAttributes []string `json:"from_attributes,omitempty"`
	ToAttributes   []string `json:"to_attributes,omitempty"`
}

func resolveAssociationUniquenessFromInput(
	assoc *inputClassAssociation,
	fromClassKey, toClassKey identity.Key,
	assocFile string,
) ([]model_class.AssociationUniqueness, error) {
	return convertInputUniqueness(assoc.Uniqueness, fromClassKey, toClassKey, assocFile)
}

func convertInputUniqueness(
	input []inputAssociationUniqueness,
	fromClassKey, toClassKey identity.Key,
	assocFile string,
) ([]model_class.AssociationUniqueness, error) {
	if len(input) == 0 {
		return nil, nil
	}
	result := make([]model_class.AssociationUniqueness, 0, len(input))
	for i, item := range input {
		fromField := fmt.Sprintf("uniqueness[%d].from_attributes", i)
		toField := fmt.Sprintf("uniqueness[%d].to_attributes", i)
		fromKeys, err := attributeKeysFromSubKeys(fromClassKey, item.FromAttributes, fromField, assocFile)
		if err != nil {
			return nil, err
		}
		toKeys, err := attributeKeysFromSubKeys(toClassKey, item.ToAttributes, toField, assocFile)
		if err != nil {
			return nil, err
		}
		if len(fromKeys) == 0 && len(toKeys) == 0 {
			return nil, convErr(
				ErrConvAssocUniquenessInvalid,
				fmt.Sprintf("uniqueness[%d] needs from_attributes or to_attributes", i),
				assocFile,
			).WithField(fmt.Sprintf("uniqueness[%d]", i))
		}
		uniqueness := model_class.NewAssociationUniqueness(fromKeys, toKeys)
		ctx := coreerr.NewContext(assocFile, fmt.Sprintf("uniqueness[%d]", i))
		if err := uniqueness.Validate(ctx); err != nil {
			return nil, mapValidationError(err)
		}
		result = append(result, uniqueness)
	}
	return result, nil
}

func attributeKeysFromSubKeys(classKey identity.Key, subKeys []string, field, assocFile string) ([]identity.Key, error) {
	if len(subKeys) == 0 {
		return nil, nil
	}
	keys := make([]identity.Key, 0, len(subKeys))
	for j, subKey := range subKeys {
		attrKey, err := identity.NewAttributeKey(classKey, subKey)
		if err != nil {
			return nil, convErr(
				ErrConvModelValidation,
				fmt.Sprintf("%s[%d]: %s", field, j, err.Error()),
				assocFile,
			).WithField(fmt.Sprintf("%s[%d]", field, j))
		}
		keys = append(keys, attrKey)
	}
	return keys, nil
}

func convertUniquenessFromModel(uniqueness []model_class.AssociationUniqueness) []inputAssociationUniqueness {
	if len(uniqueness) == 0 {
		return nil
	}
	result := make([]inputAssociationUniqueness, 0, len(uniqueness))
	for _, constraint := range uniqueness {
		if len(constraint.FromAttributeKeys) == 0 && len(constraint.ToAttributeKeys) == 0 {
			continue
		}
		result = append(result, inputAssociationUniqueness{
			FromAttributes: attributeSubKeysFromKeys(constraint.FromAttributeKeys),
			ToAttributes:   attributeSubKeysFromKeys(constraint.ToAttributeKeys),
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
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

func validateAssociationUniqueness(assoc *inputClassAssociation, assocKey, assocPath string) error {
	for i, constraint := range assoc.Uniqueness {
		if len(constraint.FromAttributes) == 0 && len(constraint.ToAttributes) == 0 {
			return NewParseError(
				ErrAssocUniquenessConstraintInvalid,
				fmt.Sprintf("association '%s' uniqueness[%d] needs from_attributes or to_attributes", assocKey, i),
				assocPath,
			).WithField(fmt.Sprintf("uniqueness[%d]", i))
		}
	}
	return nil
}
