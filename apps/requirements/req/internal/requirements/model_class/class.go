package model_class

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Class is a thing in the system.
type Class struct {
	Key             string
	Name            string
	Details         string // Markdown.
	ActorKey        string // If this class is an Actor this is the key of that actor.
	SuperclassOfKey string // If this class is part of a generalization as the superclass.
	SubclassOfKey   string // If this class is part of a generalization as a subclass.
	UmlComment      string
	// Part of the data in a parsed file.
	Attributes   []Attribute   // The attributes of a class.
	Associations []Association // How this class links to other classes.
	States       []State
	Events       []Event
	Guards       []Guard
	Actions      []Action
	Transitions  []Transition
	// Helpful data.
	DomainKey string
}

func NewClass(key, name, details, actorKey, superclassOfKey, subclassOfKey, umlComment string) (class Class, err error) {

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
		validation.Field(&class.Key, validation.Required),
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

func (c *Class) SetStates(states []State) {

	sort.Slice(states, func(i, j int) bool {
		return states[i].Key < states[j].Key
	})

	c.States = states
}

func (c *Class) SetEvents(events []Event) {

	sort.Slice(events, func(i, j int) bool {
		return events[i].Key < events[j].Key
	})

	c.Events = events
}

func (c *Class) SetGuards(guards []Guard) {

	sort.Slice(guards, func(i, j int) bool {
		return guards[i].Key < guards[j].Key
	})

	c.Guards = guards
}

func (c *Class) SetActions(actions []Action) {

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Key < actions[j].Key
	})

	c.Actions = actions
}

func (c *Class) SetTransitions(transitions []Transition) {

	sort.Slice(transitions, func(i, j int) bool {
		return transitions[i].Key < transitions[j].Key
	})

	c.Transitions = transitions
}

func (c *Class) SetDomainKey(domainKey string) {
	c.DomainKey = domainKey
}

func createKeyClassLookup(
	classAttributes map[string][]Attribute,
	classStates map[string][]State,
	classEvents map[string][]Event,
	classGuards map[string][]Guard,
	classActions map[string][]Action,
	classTransitions map[string][]Transition,
	byCategory map[string][]Class,
) (lookup map[string]Class) {

	lookup = map[string]Class{}
	for domainKey, items := range byCategory {
		for _, item := range items {

			item.SetDomainKey(domainKey)
			item.SetAttributes(classAttributes[item.Key])
			item.SetStates(classStates[item.Key])
			item.SetEvents(classEvents[item.Key])
			item.SetGuards(classGuards[item.Key])
			item.SetActions(classActions[item.Key])
			item.SetTransitions(classTransitions[item.Key])

			lookup[item.Key] = item
		}
	}
	return lookup
}
