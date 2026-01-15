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
				if parent != "" {
					return errors.Errorf("parentKey must be blank for '%s' keys, cannot be '%s'", k.keyType, parent)
				}
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

	subKey := parts[len(parts)-1]
	parentParts := parts[:len(parts)-2]
	parentKey := strings.Join(parentParts, "/")

	return newKey(parentKey, keyType, subKey)
}
