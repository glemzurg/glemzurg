package model_class

import (
	"sort"

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
	// Part of the data in a parsed file.
	Attributes   []Attribute   // The attributes of a class.
	Associations []Association // How this class links to other classes.
	States       []model_state.State
	Events       []model_state.Event
	Guards       []model_state.Guard
	Actions      []model_state.Action
	Transitions  []model_state.Transition
	// Helpful data.
	DomainKey string
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

	err = validation.ValidateStruct(&class,
		validation.Field(&class.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("invalid key type '%s' for class", k.KeyType())
			}
			return nil
		})),
		validation.Field(&class.Name, validation.Required),
	)
	if err != nil {
		return Class{}, errors.WithStack(err)
	}

	return class, nil
}

// Sort attributes by indexes first.
const _SUPER_HIGH_INDEX_NUM_FOR_SORT = 100000

func (c *Class) SetAttributes(attributes []Attribute) {

	sort.Slice(attributes, func(i, j int) bool {

		// First, if one has an index and another doesn't use the index.
		// And if they both have indexes sort by the indexes.
		iIndexNum, jIndexNum := uint(_SUPER_HIGH_INDEX_NUM_FOR_SORT), uint(_SUPER_HIGH_INDEX_NUM_FOR_SORT)
		if len(attributes[i].IndexNums) > 0 {
			iIndexNum = attributes[i].IndexNums[0]
		}
		if len(attributes[j].IndexNums) > 0 {
			jIndexNum = attributes[j].IndexNums[0]
		}
		if iIndexNum != jIndexNum {
			return iIndexNum < jIndexNum
		}

		// Non-derived attributes before derived attributes.
		iDerived := attributes[i].DerivationPolicy
		jDerived := attributes[j].DerivationPolicy
		switch {
		case iDerived == "" && jDerived != "":
			return true // i is first.
		case jDerived == "" && iDerived != "":
			return false // j is first.
		}

		// Then order by name.
		return attributes[i].Name < attributes[j].Name
	})

	c.Attributes = attributes
}

func (c *Class) SetStates(states []model_state.State) {

	sort.Slice(states, func(i, j int) bool {
		return states[i].Key.String() < states[j].Key.String()
	})

	c.States = states
}

func (c *Class) SetEvents(events []model_state.Event) {

	sort.Slice(events, func(i, j int) bool {
		return events[i].Key.String() < events[j].Key.String()
	})

	c.Events = events
}

func (c *Class) SetGuards(guards []model_state.Guard) {

	sort.Slice(guards, func(i, j int) bool {
		return guards[i].Key.String() < guards[j].Key.String()
	})

	c.Guards = guards
}

func (c *Class) SetActions(actions []model_state.Action) {

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Key.String() < actions[j].Key.String()
	})

	c.Actions = actions
}

func (c *Class) SetTransitions(transitions []model_state.Transition) {

	sort.Slice(transitions, func(i, j int) bool {
		return transitions[i].Key.String() < transitions[j].Key.String()
	})

	c.Transitions = transitions
}

func (c *Class) SetDomainKey(domainKey string) {
	c.DomainKey = domainKey
}
