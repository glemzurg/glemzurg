package identity

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/pkg/errors"
)

// Key uniquely identifies an entity in the model.
type Key struct {
	ParentKey string `validate:"-"` // The parent entity's key.
	KeyType   string // The type of the key, e.g., "class", "association".
	SubKey    string // The unique key of the child entity within its parent and type.
	SubKey2   string // Optional secondary key (e.g., for associations between two domains). Empty string means not set.
	SubKey3   string // Optional tertiary key (e.g., for association names). Empty string means not set.
}

func newKey(parentKey, keyType, subKey string) (key Key, err error) {
	return newKeyWithSubKey2(parentKey, keyType, subKey, "")
}

func newKeyWithSubKey2(parentKey, keyType, subKey string, subKey2 string) (key Key, err error) {
	return newKeyWithSubKey3(parentKey, keyType, subKey, subKey2, "")
}

func newKeyWithSubKey3(parentKey, keyType, subKey, subKey2, subKey3 string) (key Key, err error) {
	parentKey = strings.ToLower(strings.TrimSpace(parentKey))
	keyType = strings.ToLower(strings.TrimSpace(keyType))
	subKey = strings.ToLower(strings.TrimSpace(subKey))
	subKey2 = strings.ToLower(strings.TrimSpace(subKey2))
	subKey3 = strings.ToLower(strings.TrimSpace(subKey3))

	key = Key{
		ParentKey: parentKey,
		KeyType:   keyType,
		SubKey:    subKey,
		SubKey2:   subKey2,
		SubKey3:   subKey3,
	}

	err = key.Validate()
	if err != nil {
		return Key{}, errors.WithStack(err)
	}

	return key, nil
}

func newRootKey(keyType, rootKey string) (key Key, err error) {
	return newKey("", keyType, rootKey)
}

// validKeyTypes is the set of all valid KeyType values.
var validKeyTypes = map[string]bool{
	KEY_TYPE_ACTOR: true, KEY_TYPE_ACTOR_GENERALIZATION: true,
	KEY_TYPE_DOMAIN: true, KEY_TYPE_DOMAIN_ASSOCIATION: true,
	KEY_TYPE_GLOBAL_FUNCTION: true, KEY_TYPE_INVARIANT: true, KEY_TYPE_NAMED_SET: true,
	KEY_TYPE_SUBDOMAIN: true,
	KEY_TYPE_USE_CASE:  true, KEY_TYPE_USE_CASE_GENERALIZATION: true,
	KEY_TYPE_CLASS: true, KEY_TYPE_CLASS_GENERALIZATION: true,
	KEY_TYPE_CLASS_ASSOCIATION: true,
	KEY_TYPE_ATTRIBUTE:         true, KEY_TYPE_ATTRIBUTE_DERIVATION: true, KEY_TYPE_ATTRIBUTE_INVARIANT: true,
	KEY_TYPE_STATE: true, KEY_TYPE_EVENT: true, KEY_TYPE_GUARD: true,
	KEY_TYPE_ACTION: true, KEY_TYPE_QUERY: true, KEY_TYPE_TRANSITION: true,
	KEY_TYPE_CLASS_INVARIANT: true, KEY_TYPE_STATE_ACTION: true,
	KEY_TYPE_ACTION_REQUIRE: true, KEY_TYPE_ACTION_GUARANTEE: true, KEY_TYPE_ACTION_SAFETY: true,
	KEY_TYPE_QUERY_REQUIRE: true, KEY_TYPE_QUERY_GUARANTEE: true,
	KEY_TYPE_SCENARIO: true, KEY_TYPE_SCENARIO_OBJECT: true, KEY_TYPE_SCENARIO_STEP: true,
}

// identifierPattern is the regex that SubKeys must match for key types that become filenames/directories.
var identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// identifierSubKeyTypes lists key types whose SubKey must be a valid identifier (matches identifierPattern).
// Key types NOT in this set have special SubKey formats (integers, composites, class paths).
var identifierSubKeyTypes = map[string]bool{
	KEY_TYPE_ACTOR:                   true,
	KEY_TYPE_ACTOR_GENERALIZATION:    true,
	KEY_TYPE_NAMED_SET:               true,
	KEY_TYPE_DOMAIN:                  true,
	KEY_TYPE_GLOBAL_FUNCTION:         true,
	KEY_TYPE_SUBDOMAIN:               true,
	KEY_TYPE_CLASS:                   true,
	KEY_TYPE_CLASS_GENERALIZATION:    true,
	KEY_TYPE_USE_CASE:                true,
	KEY_TYPE_USE_CASE_GENERALIZATION: true,
	KEY_TYPE_ATTRIBUTE:               true,
	KEY_TYPE_STATE:                   true,
	KEY_TYPE_EVENT:                   true,
	KEY_TYPE_GUARD:                   true,
	KEY_TYPE_ACTION:                  true,
	KEY_TYPE_QUERY:                   true,
	KEY_TYPE_SCENARIO:                true,
	KEY_TYPE_SCENARIO_OBJECT:         true,
}

// Validate validates the Key struct.
func (k *Key) Validate() error {
	ctx := coreerr.NewContext("key", k.KeyType+"/"+k.SubKey)
	return k.ValidateWithContext(ctx)
}

// ValidateWithContext validates the Key struct using the given validation context.
func (k *Key) ValidateWithContext(ctx *coreerr.ValidationContext) error {
	if k.KeyType == "" {
		return coreerr.NewWithValues(ctx, coreerr.KeyTypeInvalid, "key type is required", "KeyType", "", "non-empty key type")
	}
	if !validKeyTypes[k.KeyType] {
		return coreerr.NewWithValues(ctx, coreerr.KeyTypeInvalid, fmt.Sprintf("key type '%s' is not valid", k.KeyType), "KeyType", k.KeyType, "one of the valid key types")
	}
	if k.SubKey == "" {
		return coreerr.NewWithValues(ctx, coreerr.KeySubkeyRequired, "sub key is required", "SubKey", "", "non-empty sub key")
	}

	// Validate SubKey format for key types that require identifier SubKeys.
	if identifierSubKeyTypes[k.KeyType] {
		if !identifierPattern.MatchString(k.SubKey) {
			return coreerr.NewWithValues(ctx, coreerr.KeySubkeyInvalidFormat,
				fmt.Sprintf("sub key '%s' must match pattern [a-z][a-z0-9_]* for key type '%s'", k.SubKey, k.KeyType),
				"SubKey", k.SubKey, "a value matching ^[a-z][a-z0-9_]*$")
		}
	}

	// Custom ParentKey validation (context-dependent on KeyType).
	switch k.KeyType {
	case KEY_TYPE_DOMAIN, KEY_TYPE_ACTOR, KEY_TYPE_ACTOR_GENERALIZATION, KEY_TYPE_DOMAIN_ASSOCIATION, KEY_TYPE_GLOBAL_FUNCTION, KEY_TYPE_INVARIANT, KEY_TYPE_NAMED_SET:
		// These key types must have blank parentKey.
		if k.ParentKey != "" {
			return coreerr.NewWithValues(ctx, coreerr.KeyParentkeyMustBeBlank,
				fmt.Sprintf("parentKey must be blank for '%s' keys, cannot be '%s'", k.KeyType, k.ParentKey),
				"ParentKey", k.ParentKey, "blank")
		}
	case KEY_TYPE_CLASS_ASSOCIATION:
		// Class associations can have blank parentKey (model-level) or non-blank (domain/subdomain level).
		// No validation needed - both are valid.
	default:
		if k.ParentKey == "" {
			return coreerr.NewWithValues(ctx, coreerr.KeyParentkeyRequired,
				fmt.Sprintf("parentKey must be non-blank for '%s' keys", k.KeyType),
				"ParentKey", "", "non-blank parent key")
		}
	}
	return nil
}

// String returns the string representation of the key.
func (k *Key) String() string {
	var result string
	if k.ParentKey != "" {
		result = k.ParentKey + "/" + k.KeyType + "/" + k.SubKey
	} else {
		result = k.KeyType + "/" + k.SubKey
	}
	if k.SubKey2 != "" {
		result = result + "/" + k.SubKey2
	}
	if k.SubKey3 != "" {
		result = result + "/" + k.SubKey3
	}
	return result
}

// GetSubKey returns the SubKey of the Key.
func (k *Key) GetSubKey() string {
	return k.SubKey
}

// GetSubKey2 returns the optional SubKey2 of the Key.
// Returns empty string if not set.
func (k *Key) GetSubKey2() string {
	return k.SubKey2
}

// GetSubKey3 returns the optional SubKey3 of the Key.
// Returns empty string if not set.
func (k *Key) GetSubKey3() string {
	return k.SubKey3
}

// GetKeyType returns the KeyType of the Key.
func (k *Key) GetKeyType() string {
	return k.KeyType
}

// GetParentKey returns the ParentKey of the Key as a string.
// Returns empty string if this is a root-level key (domain, actor).
func (k *Key) GetParentKey() string {
	return k.ParentKey
}

// ValidateParent validates that this key is correctly constructed based on the expected parent.
// The parent may be nil if this key type should have no parent (e.g., actor, domain).
// For class associations, the parent is determined by parsing the key structure.
//
//complexity:cyclo:warn=60,fail=60 Simple routing switch.
func (k *Key) ValidateParent(parent *Key) error {
	ctx := coreerr.NewContext("key", k.String())
	return k.ValidateParentWithContext(ctx, parent)
}

// ValidateParentWithContext validates the key's parent relationship using the given context.
//
//complexity:cyclo:warn=60,fail=60 Simple routing switch.
func (k *Key) ValidateParentWithContext(ctx *coreerr.ValidationContext, parent *Key) error {
	// First validate the key itself.
	if err := k.ValidateWithContext(ctx); err != nil {
		return err
	}

	switch k.KeyType {
	case KEY_TYPE_ACTOR, KEY_TYPE_ACTOR_GENERALIZATION, KEY_TYPE_DOMAIN, KEY_TYPE_DOMAIN_ASSOCIATION, KEY_TYPE_GLOBAL_FUNCTION, KEY_TYPE_INVARIANT, KEY_TYPE_NAMED_SET:
		return k.validateRootParent(ctx, parent)

	case KEY_TYPE_SUBDOMAIN:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_DOMAIN)

	case KEY_TYPE_USE_CASE, KEY_TYPE_USE_CASE_GENERALIZATION, KEY_TYPE_CLASS, KEY_TYPE_CLASS_GENERALIZATION:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_SUBDOMAIN)

	case KEY_TYPE_SCENARIO:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_USE_CASE)

	case KEY_TYPE_SCENARIO_OBJECT, KEY_TYPE_SCENARIO_STEP:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_SCENARIO)

	case KEY_TYPE_STATE, KEY_TYPE_EVENT, KEY_TYPE_GUARD, KEY_TYPE_ACTION, KEY_TYPE_QUERY, KEY_TYPE_TRANSITION, KEY_TYPE_ATTRIBUTE, KEY_TYPE_CLASS_INVARIANT:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_CLASS)

	case KEY_TYPE_STATE_ACTION:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_STATE)

	case KEY_TYPE_ACTION_REQUIRE, KEY_TYPE_ACTION_GUARANTEE, KEY_TYPE_ACTION_SAFETY:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_ACTION)

	case KEY_TYPE_QUERY_REQUIRE, KEY_TYPE_QUERY_GUARANTEE:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_QUERY)

	case KEY_TYPE_ATTRIBUTE_DERIVATION, KEY_TYPE_ATTRIBUTE_INVARIANT:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_ATTRIBUTE)

	case KEY_TYPE_CLASS_ASSOCIATION:
		return k.validateClassAssociationParent(ctx, parent)

	default:
		return coreerr.NewWithValues(ctx, coreerr.KeyTypeUnknown,
			fmt.Sprintf("unknown key type '%s'", k.KeyType),
			"KeyType", k.KeyType, "a valid key type")
	}
}

// validateRootParent validates that a root-level key has no parent.
func (k *Key) validateRootParent(ctx *coreerr.ValidationContext, parent *Key) error {
	if parent != nil {
		return coreerr.NewWithValues(ctx, coreerr.KeyRootHasParent,
			fmt.Sprintf("key type '%s' should not have a parent, but got parent of type '%s'", k.KeyType, parent.KeyType),
			"Parent", parent.KeyType, "nil")
	}
	if k.ParentKey != "" {
		return coreerr.NewWithValues(ctx, coreerr.KeyRootHasParentkey,
			fmt.Sprintf("key type '%s' should have empty parentKey, but got '%s'", k.KeyType, k.ParentKey),
			"ParentKey", k.ParentKey, "blank")
	}
	return nil
}

// validateRequiredParent validates that a key has a parent of the expected type with matching key value.
func (k *Key) validateRequiredParent(ctx *coreerr.ValidationContext, parent *Key, expectedType string) error {
	if parent == nil {
		return coreerr.NewWithValues(ctx, coreerr.KeyNoParent,
			fmt.Sprintf("key type '%s' requires a parent of type '%s'", k.KeyType, expectedType),
			"Parent", "", expectedType)
	}
	if parent.KeyType != expectedType {
		return coreerr.NewWithValues(ctx, coreerr.KeyWrongParentType,
			fmt.Sprintf("key type '%s' requires parent of type '%s', but got '%s'", k.KeyType, expectedType, parent.KeyType),
			"Parent", parent.KeyType, expectedType)
	}
	if k.ParentKey != parent.String() {
		return coreerr.NewWithValues(ctx, coreerr.KeyParentkeyMismatch,
			fmt.Sprintf("key parentKey '%s' does not match expected parent '%s'", k.ParentKey, parent.String()),
			"ParentKey", k.ParentKey, parent.String())
	}
	return nil
}

// validateClassAssociationParent validates the parent for class association keys.
func (k *Key) validateClassAssociationParent(ctx *coreerr.ValidationContext, parent *Key) error {
	expectedParentType, err := k.determineClassAssociationParentType(ctx)
	if err != nil {
		return err
	}

	switch expectedParentType {
	case "": // Model level - no parent.
		if parent != nil {
			return coreerr.NewWithValues(ctx, coreerr.KeyCassocModelHasParent,
				fmt.Sprintf("model-level class association should not have a parent, but got parent of type '%s'", parent.KeyType),
				"Parent", parent.KeyType, "nil")
		}
		if k.ParentKey != "" {
			return coreerr.NewWithValues(ctx, coreerr.KeyRootHasParentkey,
				fmt.Sprintf("model-level class association should have empty parentKey, but got '%s'", k.ParentKey),
				"ParentKey", k.ParentKey, "blank")
		}
	case KEY_TYPE_DOMAIN:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_DOMAIN)
	case KEY_TYPE_SUBDOMAIN:
		return k.validateRequiredParent(ctx, parent, KEY_TYPE_SUBDOMAIN)
	}
	return nil
}

// determineClassAssociationParentType determines what type of parent a class association should have
// by examining the structure of its subKey and subKey2 values.
// Returns "" for model-level, KEY_TYPE_DOMAIN for domain-level, or KEY_TYPE_SUBDOMAIN for subdomain-level.
func (k *Key) determineClassAssociationParentType(ctx *coreerr.ValidationContext) (string, error) {
	if k.KeyType != KEY_TYPE_CLASS_ASSOCIATION {
		return "", coreerr.NewWithValues(ctx, coreerr.KeyCassocParentUnknown,
			fmt.Sprintf("determineClassAssociationParentType called on non-class-association key of type '%s'", k.KeyType),
			"KeyType", k.KeyType, KEY_TYPE_CLASS_ASSOCIATION)
	}

	if k.SubKey2 == "" {
		return "", coreerr.NewWithValues(ctx, coreerr.KeySubkeyRequired,
			"class association key missing subKey2",
			"SubKey2", "", "non-empty")
	}

	// Parse the subKey to understand the structure.
	// Model level: subKey is full class path like "domain/x/subdomain/y/class/z"
	// Domain level: subKey is "subdomain/y/class/z"
	// Subdomain level: subKey is "class/z"
	subKeyParts := strings.Split(k.SubKey, "/")

	if len(subKeyParts) < 2 {
		return "", coreerr.NewWithValues(ctx, coreerr.KeyCassocParentUnknown,
			fmt.Sprintf("invalid class association subKey structure: '%s'", k.SubKey),
			"SubKey", k.SubKey, "a structured class path")
	}

	// Check the first part to determine the level.
	switch subKeyParts[0] {
	case KEY_TYPE_DOMAIN:
		// Model level - subKey starts with "domain/".
		return "", nil
	case KEY_TYPE_SUBDOMAIN:
		// Domain level - subKey starts with "subdomain/".
		return KEY_TYPE_DOMAIN, nil
	case KEY_TYPE_CLASS:
		// Subdomain level - subKey starts with "class/".
		return KEY_TYPE_SUBDOMAIN, nil
	default:
		return "", coreerr.NewWithValues(ctx, coreerr.KeyCassocParentUnknown,
			fmt.Sprintf("cannot determine class association parent type from subKey '%s'", k.SubKey),
			"SubKey", k.SubKey, "subKey starting with domain/, subdomain/, or class/")
	}
}

// IsParent returns true if the parentKey's string representation is a prefix of this key's string.
// This indicates that parentKey is an ancestor of this key in the hierarchy.
func (k *Key) IsParent(parentKey Key) bool {
	return strings.HasPrefix(k.String(), parentKey.String()+"/")
}

// HasNoParent returns true if this key has no parent component.
// This is true for root-level keys like domain and actor.
func (k *Key) HasNoParent() bool {
	return k.ParentKey == ""
}

// ParseKey parses a string representation back into a Key.
func ParseKey(s string) (key Key, err error) {
	parts, err := splitKeyParts(s)
	if err != nil {
		return Key{}, err
	}

	// Check if this is a domain association key (root-level with subKey2).
	// Format: dassociation/problemSubKey/solutionSubKey
	if len(parts) == 3 && parts[0] == KEY_TYPE_DOMAIN_ASSOCIATION {
		return newKeyWithSubKey2("", KEY_TYPE_DOMAIN_ASSOCIATION, parts[1], parts[2])
	}

	// Try parsing as a class association key.
	if k, ok, err := parseClassAssociationKey(parts); ok || err != nil {
		return k, err
	}

	// Try parsing as a multi-part key type (state action or transition).
	if k, ok, err := parseMultiPartKey(parts); ok || err != nil {
		return k, err
	}

	// Default: simple two-part key at the end.
	keyType := parts[len(parts)-2]
	subKey := parts[len(parts)-1]
	parentParts := parts[:len(parts)-2]
	parentKey := strings.Join(parentParts, "/")

	return newKey(parentKey, keyType, subKey)
}

// splitKeyParts splits a key string into trimmed parts and validates it has at least 2 parts.
func splitKeyParts(s string) ([]string, error) {
	if s == "" {
		return nil, errors.New("invalid key format")
	}
	parts := strings.Split(s, "/")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) < 2 {
		return nil, errors.New("invalid key format")
	}
	return parts, nil
}

// parseClassAssociationKey attempts to parse a class association key from parts.
// Returns (key, true, nil) on success, (Key{}, false, nil) if not a class association,
// or (Key{}, false, err) on parse error.
func parseClassAssociationKey(parts []string) (Key, bool, error) {
	for i, part := range parts {
		if part != KEY_TYPE_CLASS_ASSOCIATION {
			continue
		}
		remainingParts := parts[i+1:]
		if len(remainingParts) < 5 {
			break
		}
		classIndices := findClassIndices(remainingParts)
		if len(classIndices) < 2 {
			break
		}
		splitIdx := classIndices[0] + 2
		secondClassEndIdx := classIndices[1] + 2
		if splitIdx >= len(remainingParts) || secondClassEndIdx >= len(remainingParts) {
			break
		}
		subKey := strings.Join(remainingParts[:splitIdx], "/")
		subKey2Val := strings.Join(remainingParts[splitIdx:secondClassEndIdx], "/")
		subKey3Val := strings.Join(remainingParts[secondClassEndIdx:], "/")
		parentKey := strings.Join(parts[:i], "/")
		k, err := newKeyWithSubKey3(parentKey, KEY_TYPE_CLASS_ASSOCIATION, subKey, subKey2Val, subKey3Val)
		return k, true, err
	}
	return Key{}, false, nil
}

// findClassIndices returns the indices within parts where the value equals KEY_TYPE_CLASS.
func findClassIndices(parts []string) []int {
	var indices []int
	for j, p := range parts {
		if p == KEY_TYPE_CLASS {
			indices = append(indices, j)
		}
	}
	return indices
}

// parseMultiPartKey attempts to parse state action or transition keys whose subKey
// spans multiple slash-separated segments.
// Returns (key, true, nil) on success, (Key{}, false, nil) if not matched.
func parseMultiPartKey(parts []string) (Key, bool, error) {
	for i, part := range parts {
		remaining := parts[i+1:]
		switch {
		case part == KEY_TYPE_STATE_ACTION && len(remaining) >= 2:
			subKey := strings.Join(remaining, "/")
			parentKey := strings.Join(parts[:i], "/")
			k, err := newKey(parentKey, KEY_TYPE_STATE_ACTION, subKey)
			return k, true, err
		case part == KEY_TYPE_TRANSITION && len(remaining) >= 5:
			subKey := strings.Join(remaining, "/")
			parentKey := strings.Join(parts[:i], "/")
			k, err := newKey(parentKey, KEY_TYPE_TRANSITION, subKey)
			return k, true, err
		}
	}
	return Key{}, false, nil
}

// MarshalJSON implements json.Marshaler for Key.
// It marshals the key as its string representation.
func (k Key) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalJSON implements json.Unmarshaler for Key.
// It unmarshals a JSON string into a Key by parsing it.
func (k *Key) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Handle empty string case - return zero-value Key.
	if s == "" {
		*k = Key{}
		return nil
	}

	parsed, err := ParseKey(s)
	if err != nil {
		return err
	}

	*k = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler for Key.
// This is required for Key to be used as a map key in JSON marshalling.
func (k Key) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Key.
// This is required for Key to be used as a map key in JSON unmarshalling.
func (k *Key) UnmarshalText(data []byte) error {
	s := string(data)

	// Handle empty string case - return zero-value Key.
	if s == "" {
		*k = Key{}
		return nil
	}

	parsed, err := ParseKey(s)
	if err != nil {
		return err
	}

	*k = parsed
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Key.
// Only accepts fully formed key strings. Partial keys must be expanded
// before being unmarshaled (e.g., by parser.scopeObjectKeys).
func (k *Key) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	// Handle empty string case - return zero-value Key.
	if s == "" {
		*k = Key{}
		return nil
	}

	// Parse as a full key - partial keys are not accepted.
	parsed, err := ParseKey(s)
	if err != nil {
		return err
	}

	*k = parsed
	return nil
}
