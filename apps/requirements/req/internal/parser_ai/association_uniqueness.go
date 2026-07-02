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
) (*model_class.AssociationUniqueness, error) {
	return convertInputUniqueness(assoc.Uniqueness, fromClassKey, toClassKey, assocFile)
}

func convertInputUniqueness(
	input *inputAssociationUniqueness,
	fromClassKey, toClassKey identity.Key,
	assocFile string,
) (*model_class.AssociationUniqueness, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // optional field, absence is not an error
	}
	fromKeys, err := attributeKeysFromSubKeys(fromClassKey, input.FromAttributes, "uniqueness.from_attributes", assocFile)
	if err != nil {
		return nil, err
	}
	toKeys, err := attributeKeysFromSubKeys(toClassKey, input.ToAttributes, "uniqueness.to_attributes", assocFile)
	if err != nil {
		return nil, err
	}
	if len(fromKeys) == 0 && len(toKeys) == 0 {
		return nil, nil //nolint:nilnil // empty object means no uniqueness constraint
	}
	uniqueness := model_class.NewAssociationUniqueness(fromKeys, toKeys)
	ctx := coreerr.NewContext(assocFile, "uniqueness")
	if err := uniqueness.Validate(ctx); err != nil {
		return nil, mapValidationError(err)
	}
	return &uniqueness, nil
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

func convertUniquenessFromModel(uniqueness *model_class.AssociationUniqueness) *inputAssociationUniqueness {
	if uniqueness == nil {
		return nil
	}
	if len(uniqueness.FromAttributeKeys) == 0 && len(uniqueness.ToAttributeKeys) == 0 {
		return nil
	}
	return &inputAssociationUniqueness{
		FromAttributes: attributeSubKeysFromKeys(uniqueness.FromAttributeKeys),
		ToAttributes:   attributeSubKeysFromKeys(uniqueness.ToAttributeKeys),
	}
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
	if assoc.Uniqueness == nil {
		return nil
	}
	if len(assoc.Uniqueness.FromAttributes) == 0 && len(assoc.Uniqueness.ToAttributes) == 0 {
		return NewParseError(
			ErrAssocUniquenessConstraintInvalid,
			fmt.Sprintf("association '%s' uniqueness needs from_attributes or to_attributes", assocKey),
			assocPath,
		).WithField("uniqueness")
	}
	return nil
}
