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

	if parentKey == "" && keyType == "" {
		keyType = KEY_TYPE_MODEL
	}

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

func NewModelKey(rootKey string) (key Key, err error) {
	return NewKey("", KEY_TYPE_MODEL, rootKey)
}

// Validate validates the Key struct.
func (k *Key) Validate() error {
	return validation.ValidateStruct(k,
		validation.Field(&k.keyType, validation.Required, validation.In(KEY_TYPE_MODEL, KEY_TYPE_SUBDOMAIN, KEY_TYPE_ASSOCIATION, "class", KEY_TYPE_USE_CASE, KEY_TYPE_STATE, KEY_TYPE_EVENT, KEY_TYPE_GUARD, KEY_TYPE_GENERALIZATION, KEY_TYPE_SCENARIO, KEY_TYPE_ACTOR)),
		validation.Field(&k.subKey, validation.Required),
		validation.Field(&k.parentKey, validation.By(func(value interface{}) error {
			parent := value.(string)
			if k.keyType == KEY_TYPE_MODEL {
				if parent != "" {
					return errors.New("parentKey must be blank for model keys")
				}
			} else {
				if parent == "" {
					return errors.New("parentKey must be non-blank for non-model keys")
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
	} else if k.keyType != "" {
		return k.keyType + "/" + k.subKey
	} else {
		return k.subKey
	}
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
	var parentKey, keyType, subKey string
	switch len(parts) {
	case 1:
		keyType = "model"
		subKey = parts[0]
		if subKey == "" {
			return Key{}, errors.New("invalid key format")
		}
	case 2:
		keyType = parts[0]
		subKey = parts[1]
	case 3:
		parentKey = parts[0]
		keyType = parts[1]
		subKey = parts[2]
	default:
		return Key{}, errors.New("invalid key format")
	}

	return NewKey(parentKey, keyType, subKey)
}
