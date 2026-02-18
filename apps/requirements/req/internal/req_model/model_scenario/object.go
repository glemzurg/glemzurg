package model_scenario

import (
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
	NameStyle    string       `validate:"required,oneof=name id unnamed"` // Used to format the name in the diagram.
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
	// Validate the key.
	if err := o.Key.Validate(); err != nil {
		return err
	}
	if o.Key.KeyType != identity.KEY_TYPE_SCENARIO_OBJECT {
		return errors.Errorf("Key: invalid key type '%s' for scenario object.", o.Key.KeyType)
	}
	// Validate Name conditionally based on NameStyle.
	if o.NameStyle == _NAME_STYLE_UNNAMED {
		if o.Name != "" {
			return errors.New("Name: Name must be blank for unnamed style")
		}
	} else {
		if o.Name == "" {
			return errors.New("Name: Name cannot be blank")
		}
	}
	// Validate NameStyle (required + oneof).
	if err := _validate.Struct(o); err != nil {
		return err
	}
	// Validate ClassKey.
	if err := o.ClassKey.Validate(); err != nil {
		return errors.Wrap(err, "ClassKey")
	}
	if o.ClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return errors.Errorf("ClassKey: invalid key type '%s' for class.", o.ClassKey.KeyType)
	}
	return nil
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

// ValidateReferences validates that the object's ClassKey references a real class.
// The class must exist in the classes map (classes from the same subdomain as the use case).
func (o *Object) ValidateReferences(classes map[identity.Key]bool) error {
	if !classes[o.ClassKey] {
		return errors.Errorf("scenario object '%s' references non-existent class '%s'", o.Key.String(), o.ClassKey.String())
	}
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
