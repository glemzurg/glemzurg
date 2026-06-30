package model_class

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

const (
	AssociationUniquenessScopePerFromInstance = "per_from_instance"
	AssociationUniquenessScopePerToInstance   = "per_to_instance"
	AssociationUniquenessScopeGlobal          = "global"
)

// AssociationUniquenessKey lists attribute keys on each endpoint class that form the
// uniqueness tuple for one association constraint.
type AssociationUniquenessKey struct {
	FromAttributeKeys []identity.Key
	ToAttributeKeys   []identity.Key
}

// AssociationUniquenessConstraint limits how often the same attribute tuple may appear
// among links in one association, scoped per endpoint instance or globally.
type AssociationUniquenessConstraint struct {
	Scope    string
	Key      AssociationUniquenessKey
	MaxCount uint
}

// NewAssociationUniquenessConstraint builds a constraint with MaxCount defaulting to 1.
func NewAssociationUniquenessConstraint(scope string, key AssociationUniquenessKey, maxCount uint) AssociationUniquenessConstraint {
	if maxCount == 0 {
		maxCount = 1
	}
	return AssociationUniquenessConstraint{
		Scope:    scope,
		Key:      key,
		MaxCount: maxCount,
	}
}

// SetUniquenessConstraints replaces association uniqueness constraints.
func (a *Association) SetUniquenessConstraints(constraints []AssociationUniquenessConstraint) {
	a.UniquenessConstraints = constraints
}

// Validate validates one association uniqueness constraint structurally.
func (c *AssociationUniquenessConstraint) Validate(ctx *coreerr.ValidationContext) error {
	switch c.Scope {
	case AssociationUniquenessScopePerFromInstance,
		AssociationUniquenessScopePerToInstance,
		AssociationUniquenessScopeGlobal:
	default:
		return coreerr.NewWithValues(ctx, coreerr.AssocUniquenessScopeInvalid,
			fmt.Sprintf("scope '%s' is not valid", c.Scope),
			"Scope", c.Scope,
			fmt.Sprintf("one of: %s, %s, %s",
				AssociationUniquenessScopePerFromInstance,
				AssociationUniquenessScopePerToInstance,
				AssociationUniquenessScopeGlobal))
	}
	if len(c.Key.FromAttributeKeys) == 0 && len(c.Key.ToAttributeKeys) == 0 {
		return coreerr.New(ctx, coreerr.AssocUniquenessKeyRequired,
			"at least one from_attributes or to_attributes entry is required", "Key")
	}
	if c.MaxCount < 1 {
		return coreerr.NewWithValues(ctx, coreerr.AssocUniquenessMaxInvalid,
			fmt.Sprintf("max must be at least 1, got %d", c.MaxCount),
			"MaxCount", fmt.Sprintf("%d", c.MaxCount), ">=1")
	}
	return nil
}

// ValidateAttributeReferences checks that attribute keys exist on the endpoint classes.
func (c *AssociationUniquenessConstraint) ValidateAttributeReferences(
	ctx *coreerr.ValidationContext,
	fromClass, toClass Class,
) error {
	for i, attrKey := range c.Key.FromAttributeKeys {
		if err := validateEndpointAttributeKey(ctx, attrKey, fromClass, "from_attributes", i); err != nil {
			return err
		}
	}
	for i, attrKey := range c.Key.ToAttributeKeys {
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

func (a *Association) validateUniquenessConstraints(ctx *coreerr.ValidationContext) error {
	seen := make(map[string]bool)
	for i, constraint := range a.UniquenessConstraints {
		childCtx := ctx.Child("uniquenessConstraint", fmt.Sprintf("%d", i))
		if err := constraint.Validate(childCtx); err != nil {
			return err
		}
		fingerprint := constraint.fingerprint()
		if seen[fingerprint] {
			return coreerr.New(childCtx, coreerr.AssocUniquenessDuplicate,
				"duplicate uniqueness constraint", "UniquenessConstraints")
		}
		seen[fingerprint] = true
	}
	return nil
}

func (c *AssociationUniquenessConstraint) fingerprint() string {
	from := attributeKeysFingerprint(c.Key.FromAttributeKeys)
	to := attributeKeysFingerprint(c.Key.ToAttributeKeys)
	return fmt.Sprintf("%s|%s|%s|%d", c.Scope, from, to, c.MaxCount)
}

func attributeKeysFingerprint(keys []identity.Key) string {
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = key.String()
	}
	return strings.Join(parts, ",")
}

func classHasAttributeKey(class Class, attrKey identity.Key) bool {
	for _, attr := range class.Attributes {
		if attr.Key == attrKey {
			return true
		}
	}
	return false
}
