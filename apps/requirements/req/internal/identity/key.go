package identity

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Key uniquely identifies an entity in the model.
type Key struct {
	parentKey string  // The parent entity's key.
	keyType   string  // The type of the key, e.g., "class", "association".
	subKey    string  // The unique key of the child entity within its parent and type.
	subKey2   *string // Optional secondary key (e.g., for associations between two domains).
}

func newKey(parentKey, keyType, subKey string) (key Key, err error) {
	return newKeyWithSubKey2(parentKey, keyType, subKey, nil)
}

func newKeyWithSubKey2(parentKey, keyType, subKey string, subKey2 *string) (key Key, err error) {
	parentKey = strings.ToLower(strings.TrimSpace(parentKey))
	keyType = strings.ToLower(strings.TrimSpace(keyType))
	subKey = strings.ToLower(strings.TrimSpace(subKey))

	var subKey2Ptr *string
	if subKey2 != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*subKey2))
		subKey2Ptr = &trimmed
	}

	key = Key{
		parentKey: parentKey,
		keyType:   keyType,
		subKey:    subKey,
		subKey2:   subKey2Ptr,
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

// Validate validates the Key struct.
func (k *Key) Validate() error {
	return validation.ValidateStruct(k,
		validation.Field(&k.keyType, validation.Required, validation.In(
			KEY_TYPE_DOMAIN,
			KEY_TYPE_DOMAIN_ASSOCIATION,
			KEY_TYPE_SUBDOMAIN,
			KEY_TYPE_USE_CASE,
			KEY_TYPE_CLASS,
			KEY_TYPE_STATE,
			KEY_TYPE_EVENT,
			KEY_TYPE_GUARD,
			KEY_TYPE_ACTION,
			KEY_TYPE_TRANSITION,
			KEY_TYPE_GENERALIZATION,
			KEY_TYPE_SCENARIO,
			KEY_TYPE_SCENARIO_OBJECT,
			KEY_TYPE_ACTOR,
			KEY_TYPE_CLASS_ASSOCIATION,
			KEY_TYPE_ATTRIBUTE,
			KEY_TYPE_STATE_ACTION,
		)),
		validation.Field(&k.subKey, validation.Required),
		validation.Field(&k.parentKey, validation.By(func(value interface{}) error {
			parent := value.(string)
			switch k.keyType {
			case KEY_TYPE_DOMAIN, KEY_TYPE_ACTOR:
				// These key types must have blank parentKey.
				if parent != "" {
					return errors.Errorf("parentKey must be blank for '%s' keys, cannot be '%s'", k.keyType, parent)
				}
			case KEY_TYPE_CLASS_ASSOCIATION:
				// Class associations can have blank parentKey (model-level) or non-blank (domain/subdomain level).
				// No validation needed - both are valid.
			default:
				if parent == "" {
					return errors.Errorf("parentKey must be non-blank for '%s' keys", k.keyType)
				}
			}
			return nil
		})),
	)
}

// String returns the string representation of the key.
func (k *Key) String() string {
	var result string
	if k.parentKey != "" {
		result = k.parentKey + "/" + k.keyType + "/" + k.subKey
	} else {
		result = k.keyType + "/" + k.subKey
	}
	if k.subKey2 != nil {
		result = result + "/" + *k.subKey2
	}
	return result
}

// SubKey returns the subKey of the Key.
func (k *Key) SubKey() string {
	return k.subKey
}

// SubKey2 returns the optional subKey2 of the Key.
func (k *Key) SubKey2() *string {
	return k.subKey2
}

// KeyType returns the keyType of the Key.
func (k *Key) KeyType() string {
	return k.keyType
}

// ParentKey returns the parentKey of the Key as a string.
// Returns empty string if this is a root-level key (domain, actor).
func (k *Key) ParentKey() string {
	return k.parentKey
}

// ValidateParent validates that this key is correctly constructed based on the expected parent.
// The parent may be nil if this key type should have no parent (e.g., actor, domain).
// For class associations, the parent is determined by parsing the key structure.
func (k *Key) ValidateParent(parent *Key) error {
	// First validate the key itself.
	if err := k.Validate(); err != nil {
		return err
	}

	switch k.keyType {
	case KEY_TYPE_ACTOR, KEY_TYPE_DOMAIN:
		// These are root keys - parent must be nil.
		if parent != nil {
			return errors.Errorf("key type '%s' should not have a parent, but got parent of type '%s'", k.keyType, parent.keyType)
		}
		if k.parentKey != "" {
			return errors.Errorf("key type '%s' should have empty parentKey, but got '%s'", k.keyType, k.parentKey)
		}

	case KEY_TYPE_SUBDOMAIN:
		// Parent must be a domain.
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_DOMAIN)
		}
		if parent.keyType != KEY_TYPE_DOMAIN {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_DOMAIN, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_DOMAIN_ASSOCIATION:
		// Parent must be a domain (the problem domain).
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_DOMAIN)
		}
		if parent.keyType != KEY_TYPE_DOMAIN {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_DOMAIN, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_USE_CASE, KEY_TYPE_CLASS, KEY_TYPE_GENERALIZATION:
		// Parent must be a subdomain.
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_SUBDOMAIN)
		}
		if parent.keyType != KEY_TYPE_SUBDOMAIN {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_SUBDOMAIN, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_SCENARIO:
		// Parent must be a use case.
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_USE_CASE)
		}
		if parent.keyType != KEY_TYPE_USE_CASE {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_USE_CASE, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_SCENARIO_OBJECT:
		// Parent must be a scenario.
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_SCENARIO)
		}
		if parent.keyType != KEY_TYPE_SCENARIO {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_SCENARIO, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_STATE, KEY_TYPE_EVENT, KEY_TYPE_GUARD, KEY_TYPE_ACTION, KEY_TYPE_TRANSITION, KEY_TYPE_ATTRIBUTE:
		// Parent must be a class.
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_CLASS)
		}
		if parent.keyType != KEY_TYPE_CLASS {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_CLASS, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_STATE_ACTION:
		// Parent must be a state.
		if parent == nil {
			return errors.Errorf("key type '%s' requires a parent of type '%s'", k.keyType, KEY_TYPE_STATE)
		}
		if parent.keyType != KEY_TYPE_STATE {
			return errors.Errorf("key type '%s' requires parent of type '%s', but got '%s'", k.keyType, KEY_TYPE_STATE, parent.keyType)
		}
		if k.parentKey != parent.String() {
			return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
		}

	case KEY_TYPE_CLASS_ASSOCIATION:
		// Class associations are special - the parent can be subdomain, domain, or model (empty).
		// We need to determine the expected parent type by parsing the key structure.
		expectedParentType, err := k.determineClassAssociationParentType()
		if err != nil {
			return err
		}

		switch expectedParentType {
		case "": // Model level - no parent
			if parent != nil {
				return errors.Errorf("model-level class association should not have a parent, but got parent of type '%s'", parent.keyType)
			}
			if k.parentKey != "" {
				return errors.Errorf("model-level class association should have empty parentKey, but got '%s'", k.parentKey)
			}
		case KEY_TYPE_DOMAIN:
			if parent == nil {
				return errors.Errorf("domain-level class association requires a parent of type '%s'", KEY_TYPE_DOMAIN)
			}
			if parent.keyType != KEY_TYPE_DOMAIN {
				return errors.Errorf("domain-level class association requires parent of type '%s', but got '%s'", KEY_TYPE_DOMAIN, parent.keyType)
			}
			if k.parentKey != parent.String() {
				return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
			}
		case KEY_TYPE_SUBDOMAIN:
			if parent == nil {
				return errors.Errorf("subdomain-level class association requires a parent of type '%s'", KEY_TYPE_SUBDOMAIN)
			}
			if parent.keyType != KEY_TYPE_SUBDOMAIN {
				return errors.Errorf("subdomain-level class association requires parent of type '%s', but got '%s'", KEY_TYPE_SUBDOMAIN, parent.keyType)
			}
			if k.parentKey != parent.String() {
				return errors.Errorf("key parentKey '%s' does not match expected parent '%s'", k.parentKey, parent.String())
			}
		}

	default:
		return errors.Errorf("unknown key type '%s'", k.keyType)
	}

	return nil
}

// determineClassAssociationParentType determines what type of parent a class association should have
// by examining the structure of its subKey and subKey2 values.
// Returns "" for model-level, KEY_TYPE_DOMAIN for domain-level, or KEY_TYPE_SUBDOMAIN for subdomain-level.
func (k *Key) determineClassAssociationParentType() (string, error) {
	if k.keyType != KEY_TYPE_CLASS_ASSOCIATION {
		return "", errors.Errorf("determineClassAssociationParentType called on non-class-association key of type '%s'", k.keyType)
	}

	if k.subKey2 == nil {
		return "", errors.New("class association key missing subKey2")
	}

	// Parse the subKey to understand the structure.
	// Model level: subKey is full class path like "domain/x/subdomain/y/class/z"
	// Domain level: subKey is "subdomain/y/class/z"
	// Subdomain level: subKey is "class/z"
	subKeyParts := strings.Split(k.subKey, "/")

	if len(subKeyParts) < 2 {
		return "", errors.Errorf("invalid class association subKey structure: '%s'", k.subKey)
	}

	// Check the first part to determine the level.
	switch subKeyParts[0] {
	case KEY_TYPE_DOMAIN:
		// Model level - subKey starts with "domain/"
		return "", nil
	case KEY_TYPE_SUBDOMAIN:
		// Domain level - subKey starts with "subdomain/"
		return KEY_TYPE_DOMAIN, nil
	case KEY_TYPE_CLASS:
		// Subdomain level - subKey starts with "class/"
		return KEY_TYPE_SUBDOMAIN, nil
	default:
		return "", errors.Errorf("cannot determine class association parent type from subKey '%s'", k.subKey)
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
	return k.parentKey == ""
}

func ParseKey(s string) (key Key, err error) {
	if s == "" {
		return Key{}, errors.New("invalid key format")
	}
	parts := strings.Split(s, "/")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) < 2 {
		return Key{}, errors.New("invalid key format2")
	}

	// Check if this is a key type that uses subKey2 (domain association).
	// Format: parentKey/keyType/subKey/subKey2
	// For domain association: domain/problemdomain/dassociation/problemsubkey/solutionsubkey
	var subKey2 *string
	keyType := parts[len(parts)-2]

	// If this looks like a domain association with subKey2, handle specially
	if len(parts) >= 5 && parts[len(parts)-3] == KEY_TYPE_DOMAIN_ASSOCIATION {
		keyType = parts[len(parts)-3]
		subKey := parts[len(parts)-2]
		subKey2Val := parts[len(parts)-1]
		subKey2 = &subKey2Val
		parentParts := parts[:len(parts)-3]
		parentKey := strings.Join(parentParts, "/")
		return newKeyWithSubKey2(parentKey, keyType, subKey, subKey2)
	}

	// Check if this is a class association with subKey2.
	// Format: parentKey/cassociation/subKey/subKey2
	// where subKey and subKey2 are class paths that end with "class/name".
	// For subdomain parent: parent/cassociation/class/class_a/class/class_b
	// For domain parent: parent/cassociation/subdomain/s_a/class/c_a/subdomain/s_b/class/c_b
	// For model parent: cassociation/domain/d_a/subdomain/s_a/class/c_a/domain/d_b/subdomain/s_b/class/c_b
	// We need to find where "cassociation" is in the parts and then find the second "class" to split.
	for i, part := range parts {
		if part == KEY_TYPE_CLASS_ASSOCIATION {
			// Found cassociation. The subKey and subKey2 are the remaining parts.
			remainingParts := parts[i+1:]
			if len(remainingParts) >= 4 {
				// Find all occurrences of "class" in remainingParts
				classIndices := []int{}
				for j, p := range remainingParts {
					if p == KEY_TYPE_CLASS {
						classIndices = append(classIndices, j)
					}
				}
				// We need at least 2 "class" occurrences (one for each endpoint)
				if len(classIndices) >= 2 {
					// The first class path ends after the first "class/name" pair.
					// The split point is the element AFTER the first class key (i.e., classIndices[0] + 2)
					splitIdx := classIndices[0] + 2
					if splitIdx < len(remainingParts) {
						subKey := strings.Join(remainingParts[:splitIdx], "/")
						subKey2Val := strings.Join(remainingParts[splitIdx:], "/")
						parentParts := parts[:i]
						parentKey := strings.Join(parentParts, "/")
						return newKeyWithSubKey2(parentKey, KEY_TYPE_CLASS_ASSOCIATION, subKey, &subKey2Val)
					}
				}
			}
			break
		}
	}

	// Handle state action key type with format: parentKey/saction/when/subKey
	for i, part := range parts {
		if part == KEY_TYPE_STATE_ACTION && i+2 < len(parts) {
			// Found saction. The subKey is when/subKey (the remaining parts).
			remainingParts := parts[i+1:]
			if len(remainingParts) >= 2 {
				subKey := strings.Join(remainingParts, "/")
				parentParts := parts[:i]
				parentKey := strings.Join(parentParts, "/")
				return newKey(parentKey, KEY_TYPE_STATE_ACTION, subKey)
			}
		}
	}

	// Handle transition key type with format: parentKey/transition/from/event/guard/action/to
	for i, part := range parts {
		if part == KEY_TYPE_TRANSITION && i+5 < len(parts) {
			// Found transition. The subKey is from/event/guard/action/to (the remaining parts).
			remainingParts := parts[i+1:]
			if len(remainingParts) >= 5 {
				subKey := strings.Join(remainingParts, "/")
				parentParts := parts[:i]
				parentKey := strings.Join(parentParts, "/")
				return newKey(parentKey, KEY_TYPE_TRANSITION, subKey)
			}
		}
	}

	subKey := parts[len(parts)-1]
	parentParts := parts[:len(parts)-2]
	parentKey := strings.Join(parentParts, "/")

	return newKey(parentKey, keyType, subKey)
}
