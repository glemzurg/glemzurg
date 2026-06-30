package model_class

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// AssociationUniquenessKey lists attribute keys on each endpoint class that form the
// uniqueness tuple for one association constraint. If only one side has attributes, the
// implied uniqueness on the other side is by instance.
type AssociationUniqueness struct {
	FromAttributeKeys []identity.Key
	ToAttributeKeys   []identity.Key
}

// NewAssociationUniquenessConstraint builds a constraint with MaxCount defaulting to 1.
func NewAssociationUniqueness(fromAttributeKeys, toAttributeKeys []identity.Key) AssociationUniqueness {

	return AssociationUniqueness{
		FromAttributeKeys: fromAttributeKeys,
		ToAttributeKeys:   toAttributeKeys,
	}
}

// Validate validates one association uniqueness constraint structurally.
func (c *AssociationUniqueness) Validate(ctx *coreerr.ValidationContext) error {
	if len(c.FromAttributeKeys) == 0 && len(c.ToAttributeKeys) == 0 {
		return coreerr.New(ctx, coreerr.AssocUniquenessKeyRequired,
			"at least one from_attributes or to_attributes entry is required", "Key")
	}
	return nil
}

// ValidateAttributeReferences checks that attribute keys exist on the endpoint classes.
func (c *AssociationUniqueness) ValidateAttributeReferences(
	ctx *coreerr.ValidationContext,
	fromClass, toClass Class,
) error {
	for i, attrKey := range c.FromAttributeKeys {
		if err := validateEndpointAttributeKey(ctx, attrKey, fromClass, "from_attributes", i); err != nil {
			return err
		}
	}
	for i, attrKey := range c.ToAttributeKeys {
		if err := validateEndpointAttributeKey(ctx, attrKey, toClass, "to_attributes", i); err != nil {
			return err
		}
	}
	return nil
}

func validateEndpointAttributeKey(
	ctx *coreerr.ValidationContext,
	attrKey identity.Key,
	class Class,
	field string,
	index int,
) error {
	if err := attrKey.ValidateWithContext(ctx); err != nil {
		return coreerr.NewWithValues(ctx, coreerr.AssocUniquenessFromAttrNotfound,
			fmt.Sprintf("%s[%d]: attribute key invalid: %s", field, index, err.Error()),
			field, attrKey.String(), class.Key.String())
	}
	if attrKey.KeyType != identity.KEY_TYPE_ATTRIBUTE {
		return coreerr.NewWithValues(ctx, coreerr.AssocUniquenessFromAttrNotfound,
			fmt.Sprintf("%s[%d]: key %q is not an attribute key", field, index, attrKey.String()),
			field, attrKey.String(), class.Key.String())
	}
	if !classHasAttributeKey(class, attrKey) {
		code := coreerr.AssocUniquenessFromAttrNotfound
		if field == "to_attributes" {
			code = coreerr.AssocUniquenessToAttrNotfound
		}
		return coreerr.NewWithValues(ctx, code,
			fmt.Sprintf("%s[%d]: attribute %q not found on class %q", field, index, attrKey.SubKey, class.Name),
			field, attrKey.String(), class.Key.String())
	}
	return nil
}

func classHasAttributeKey(class Class, attrKey identity.Key) bool {
	for _, attr := range class.Attributes {
		if attr.Key == attrKey {
			return true
		}
	}
	return false
}
