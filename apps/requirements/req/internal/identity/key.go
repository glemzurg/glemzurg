package identity

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Key uniquely identifies an entity in the model.
type Key struct {
	parentKey string // The parent entity's key.
	keyType   string // The type of the key, e.g., "class", "association".
	subKey    string // The unique key of the child entity within its parent and type.
}

func NewKey(parentKey, keyType, subKey string) (key Key, err error) {
	parentKey = strings.ToLower(strings.TrimSpace(parentKey))
	keyType = strings.ToLower(strings.TrimSpace(keyType))
	subKey = strings.ToLower(strings.TrimSpace(subKey))

	key = Key{
		parentKey: parentKey,
		keyType:   keyType,
		subKey:    subKey,
	}

	err = key.Validate()
	if err != nil {
		return Key{}, errors.WithStack(err)
	}

	return key, nil
}

func NewRootKey(keyType, rootKey string) (key Key, err error) {
	return NewKey("", keyType, rootKey)
}

// Validate validates the Key struct.
func (k *Key) Validate() error {
	return validation.ValidateStruct(k,
		validation.Field(&k.keyType, validation.Required, validation.In(
			KEY_TYPE_DOMAIN,
			KEY_TYPE_SUBDOMAIN,
			KEY_TYPE_ASSOCIATION,
			KEY_TYPE_CLASS,
			KEY_TYPE_USE_CASE,
			KEY_TYPE_STATE,
			KEY_TYPE_EVENT,
			KEY_TYPE_GUARD,
			KEY_TYPE_GENERALIZATION,
			KEY_TYPE_SCENARIO,
			KEY_TYPE_ACTOR,
		)),
		validation.Field(&k.subKey, validation.Required),
		validation.Field(&k.parentKey, validation.By(func(value interface{}) error {
			parent := value.(string)
			switch k.keyType {
			case KEY_TYPE_DOMAIN, KEY_TYPE_USE_CASE:
				if parent != "" {
					return errors.Errorf("parentKey must be blank for '%s' keys", k.keyType)
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
	if k.parentKey != "" {
		return k.parentKey + "/" + k.keyType + "/" + k.subKey
	}
	return k.keyType + "/" + k.subKey
}

// SubKey returns the subKey of the Key.
func (k *Key) SubKey() string {
	return k.subKey
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
	subKey := parts[len(parts)-1]
	keyType := parts[len(parts)-2]
	parentParts := parts[:len(parts)-2]
	parentKey := strings.Join(parentParts, "/")

	return NewKey(parentKey, keyType, subKey)
}
