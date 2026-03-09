package model_scenario

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

func NewObject(key identity.Key, objectNumber uint, name, nameStyle string, classKey identity.Key, multi bool, umlComment string) Object {
	return Object{
		Key:          key,
		ObjectNumber: objectNumber,
		Name:         name,
		NameStyle:    nameStyle,
		ClassKey:     classKey,
		Multi:        multi,
		UmlComment:   umlComment,
	}
}

// Validate validates the Object struct.
func (o *Object) Validate() error {
	// Validate the key.
	if err := o.Key.Validate(); err != nil {
		return coreerr.New(coreerr.SobjectKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if o.Key.KeyType != identity.KEY_TYPE_SCENARIO_OBJECT {
		return coreerr.NewWithValues(coreerr.SobjectKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for scenario object", o.Key.KeyType), "Key", o.Key.KeyType, identity.KEY_TYPE_SCENARIO_OBJECT)
	}
	// Validate NameStyle required.
	if o.NameStyle == "" {
		return coreerr.New(coreerr.SobjectNamestyleRequired, "NameStyle is required", "NameStyle")
	}
	// Validate NameStyle is one of the valid values.
	if o.NameStyle != _NAME_STYLE_NAME && o.NameStyle != _NAME_STYLE_ID && o.NameStyle != _NAME_STYLE_UNNAMED {
		return coreerr.NewWithValues(coreerr.SobjectNamestyleInvalid, "NameStyle must be one of: name, id, unnamed", "NameStyle", o.NameStyle, "one of: name, id, unnamed")
	}
	// Validate Name conditionally based on NameStyle.
	if o.NameStyle == _NAME_STYLE_UNNAMED {
		if o.Name != "" {
			return coreerr.NewWithValues(coreerr.SobjectNameMustBeBlank, "Name: Name must be blank for unnamed style", "Name", o.Name, "")
		}
	} else {
		if o.Name == "" {
			return coreerr.New(coreerr.SobjectNameRequired, "Name: Name cannot be blank", "Name")
		}
	}
	// Validate ClassKey.
	if err := o.ClassKey.Validate(); err != nil {
		return coreerr.New(coreerr.SobjectClasskeyInvalid, fmt.Sprintf("ClassKey: %s", err.Error()), "ClassKey")
	}
	if o.ClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return coreerr.NewWithValues(coreerr.SobjectClasskeyTypeInvalid, fmt.Sprintf("classKey: invalid key type '%s' for class", o.ClassKey.KeyType), "ClassKey", o.ClassKey.KeyType, identity.KEY_TYPE_CLASS)
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
		return coreerr.NewWithValues(coreerr.SobjectClassNotfound, fmt.Sprintf("scenario object '%s' references non-existent class '%s'", o.Key.String(), o.ClassKey.String()), "ClassKey", o.ClassKey.String(), "")
	}
	return nil
}

func (o *Object) GetName(class model_class.Class) (name string) {
	switch o.NameStyle {
	case _NAME_STYLE_NAME:
		name = o.Name + ":" + class.Name
	case _NAME_STYLE_ID:
		name = class.Name + " " + o.Name
	case _NAME_STYLE_UNNAMED:
		name = ":" + class.Name
	default:
		panic("unknown name style: " + o.NameStyle)
	}
	if o.Multi {
		name = "*" + name
	}
	return name
}
