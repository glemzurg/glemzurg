package identity

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Key uniquely identifies an entity in the model.
type Key struct {
	parentKey string // The parent entity's key.
	childType string // The type of the child entity, e.g., "class", "association".
	subKey    string // The unique key of the child entity within its parent and type.
}

func NewKey(parentKey, childType, subKey string) (key Key, err error) {
	parentKey = strings.ToLower(strings.TrimSpace(parentKey))
	childType = strings.ToLower(strings.TrimSpace(childType))
	subKey = strings.ToLower(strings.TrimSpace(subKey))

	key = Key{
		parentKey: parentKey,
		childType: childType,
		subKey:    subKey,
	}

	err = key.Validate()
	if err != nil {
		return Key{}, errors.WithStack(err)
	}

	return key, nil
}

func NewRootKey(rootKey string) (key Key, err error) {
	return NewKey("", "", rootKey)
}

// Validate validates the Key struct.
func (k *Key) Validate() error {
	return validation.ValidateStruct(k,
		validation.Field(&k.subKey, validation.Required),
		validation.Field(&k.parentKey, validation.By(func(value interface{}) error {
			parent := value.(string)
			childType := k.childType
			if (parent == "" && childType != "") || (parent != "" && childType == "") {
				return errors.New("ParentKey and ChildType must both be set or both be blank")
			}
			return nil
		})),
	)
}

// String returns the string representation of the key.
func (k *Key) String() string {
	if k.parentKey != "" && k.childType != "" {
		return k.parentKey + "/" + k.childType + "/" + k.subKey
	}
	return k.subKey
}

// SubKey returns the subKey of the Key.
func (k *Key) SubKey() string {
	return k.subKey
}

// ChildType returns the childType of the Key.
func (k *Key) ChildType() string {
	return k.childType
}

func ParseKey(s string) (key Key, err error) {
	if s == "" {
		return Key{}, errors.New("invalid key format")
	}
	parts := strings.Split(s, "/")
	var parentKey, childType, subKey string
	switch len(parts) {
	case 1:
		subKey = parts[0]
		if subKey == "" {
			return Key{}, errors.New("invalid key format")
		}
	case 3:
		parentKey = parts[0]
		childType = parts[1]
		subKey = parts[2]
	default:
		return Key{}, errors.New("invalid key format")
	}

	return NewKey(parentKey, childType, subKey)
}
