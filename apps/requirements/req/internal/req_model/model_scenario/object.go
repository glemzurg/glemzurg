package model_scenario

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
)

const (
	_NAME_STYLE_NAME    = "name"
	_NAME_STYLE_ID      = "id"
	_NAME_STYLE_UNNAMED = "unnamed" // Name must be blank.
)

// Object is an object that participates in a scenario.
type Object struct {
	Key          identity.Key
	ObjectNumber uint         // Order in the scenario diagram.
	Name         string       // The name or id of the object.
	NameStyle    string       // Used to format the name in the diagram.
	ClassKey     identity.Key // The class key this object is an instance of.
	Multi        bool
	UmlComment   string
}

func NewObject(key identity.Key, objectNumber uint, name, nameStyle string, classKey identity.Key, multi bool, umlComment string) (object Object, err error) {

	object = Object{
		Key:          key,
		ObjectNumber: objectNumber,
		Name:         name,
		NameStyle:    nameStyle,
		ClassKey:     classKey,
		Multi:        multi,
		UmlComment:   umlComment,
	}

	if err = object.Validate(); err != nil {
		return Object{}, err
	}

	return object, nil
}

// Validate validates the Object struct.
func (o *Object) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_SCENARIO_OBJECT {
				return errors.Errorf("invalid key type '%s' for scenario object", k.KeyType())
			}
			return nil
		})),
		validation.Field(&o.Name, validation.By(func(value interface{}) error {
			name := value.(string)
			if o.NameStyle == _NAME_STYLE_UNNAMED {
				if name != "" {
					return errors.New("Name must be blank for unnamed style")
				}
			} else {
				if name == "" {
					return errors.New("Name cannot be blank")
				}
			}
			return nil
		})),
		validation.Field(&o.NameStyle, validation.Required, validation.In(_NAME_STYLE_NAME, _NAME_STYLE_ID, _NAME_STYLE_UNNAMED)),
		validation.Field(&o.ClassKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("invalid key type '%s' for class", k.KeyType())
			}
			return nil
		})),
	)
}

// ValidateWithParent validates the Object, its key's parent relationship, and all children.
// The parent must be a Scenario.
func (o *Object) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := o.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := o.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Object has no children with keys that need validation.
	return nil
}

func (so *Object) GetName(class model_class.Class) (name string) {
	switch so.NameStyle {
	case _NAME_STYLE_NAME:
		name = so.Name + ":" + class.Name
	case _NAME_STYLE_ID:
		name = class.Name + " " + so.Name
	case _NAME_STYLE_UNNAMED:
		name = ":" + class.Name
	default:
		panic("unknown name style: " + so.NameStyle)
	}
	if so.Multi {
		name = "*" + name
	}
	return name
}
