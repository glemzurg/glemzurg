package model_class

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

// Class is a thing in the system.
type Class struct {
	Key             identity.Key
	Name            string
	Details         string        // Markdown.
	ActorKey        *identity.Key // If this class is an Actor this is the key of that actor.
	SuperclassOfKey *identity.Key // If this class is part of a generalization as the superclass.
	SubclassOfKey   *identity.Key // If this class is part of a generalization as a subclass.
	UmlComment      string
	// Children
	Attributes  map[identity.Key]Attribute // The attributes of a class.
	States      map[identity.Key]model_state.State
	Events      map[identity.Key]model_state.Event
	Guards      map[identity.Key]model_state.Guard
	Actions     map[identity.Key]model_state.Action
	Transitions map[identity.Key]model_state.Transition
}

func NewClass(key identity.Key, name, details string, actorKey, superclassOfKey, subclassOfKey *identity.Key, umlComment string) (class Class, err error) {

	class = Class{
		Key:             key,
		Name:            name,
		Details:         details,
		ActorKey:        actorKey,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
		UmlComment:      umlComment,
	}

	if err = class.Validate(); err != nil {
		return Class{}, err
	}

	return class, nil
}

// Validate validates the Class struct.
func (c *Class) Validate() error {
	if err := validation.ValidateStruct(c,
		validation.Field(&c.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("invalid key type '%s' for class", k.KeyType())
			}
			return nil
		})),
		validation.Field(&c.Name, validation.Required),
	); err != nil {
		return err
	}
	// SuperclassOfKey and SubclassOfKey cannot be the same.
	if c.SuperclassOfKey != nil && c.SubclassOfKey != nil && *c.SuperclassOfKey == *c.SubclassOfKey {
		return errors.New("SuperclassOfKey and SubclassOfKey cannot be the same")
	}
	return nil
}

func (c *Class) SetAttributes(attributes map[identity.Key]Attribute) {
	c.Attributes = attributes
}

func (c *Class) SetStates(states map[identity.Key]model_state.State) {
	c.States = states
}

func (c *Class) SetEvents(events map[identity.Key]model_state.Event) {
	c.Events = events
}

func (c *Class) SetGuards(guards map[identity.Key]model_state.Guard) {
	c.Guards = guards
}

func (c *Class) SetActions(actions map[identity.Key]model_state.Action) {
	c.Actions = actions
}

func (c *Class) SetTransitions(transitions map[identity.Key]model_state.Transition) {
	c.Transitions = transitions
}

// ValidateWithParent validates the Class, its key's parent relationship, and all children.
// The parent must be a Subdomain.
func (c *Class) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := c.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := c.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for _, attr := range c.Attributes {
		if err := attr.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, state := range c.States {
		if err := state.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, event := range c.Events {
		if err := event.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, guard := range c.Guards {
		if err := guard.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, action := range c.Actions {
		if err := action.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	for _, transition := range c.Transitions {
		if err := transition.ValidateWithParent(&c.Key); err != nil {
			return err
		}
	}
	return nil
}
