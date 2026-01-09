package identity

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Key uniquely identifies an entity in the model.
type Key struct {
	ParentKey string // The parent entity's key.
	ChildType string // The type of the child entity, e.g., "class", "association".
	SubKey    string // The unique key of the child entity within its parent and type.
}

func NewKey(parentKey, childType, subKey string) (key Key, err error) {
	parentKey = strings.ToLower(strings.TrimSpace(parentKey))
	childType = strings.ToLower(strings.TrimSpace(childType))
	subKey = strings.ToLower(strings.TrimSpace(subKey))

	key = Key{
		ParentKey: parentKey,
		ChildType: childType,
		SubKey:    subKey,
	}

	err = key.Validate()
	if err != nil {
		return Key{}, errors.WithStack(err)
	}

	return key, nil
}

// Validate validates the Key struct.
func (k Key) Validate() error {
	return validation.ValidateStruct(&k,
		validation.Field(&k.SubKey, validation.Required),
		validation.Field(&k.ParentKey, validation.By(func(value interface{}) error {
			parent := value.(string)
			childType := k.ChildType
			if (parent == "" && childType != "") || (parent != "" && childType == "") {
				return errors.New("ParentKey and ChildType must both be set or both be blank")
			}
			return nil
		})),
	)
}

// String returns the string representation of the key.
func (k Key) String() string {
	if k.ParentKey != "" && k.ChildType != "" {
		return k.ParentKey + "/" + k.ChildType + "/" + k.SubKey
	}
	return k.SubKey
}
