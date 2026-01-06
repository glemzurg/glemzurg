package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
)

// classInOut is a thing in the system.
type classInOut struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details"`           // Markdown.
	ActorKey        string `json:"actor_key"`         // If this class is an Actor this is the key of that actor.
	SuperclassOfKey string `json:"superclass_of_key"` // If this class is part of a generalization as the superclass.
	SubclassOfKey   string `json:"subclass_of_key"`   // If this class is part of a generalization as a subclass.
	UmlComment      string `json:"uml_comment"`
	// Nested.
	Attributes  []attributeInOut  `json:"attributes"`
	States      []stateInOut      `json:"states"`
	Events      []eventInOut      `json:"events"`
	Guards      []guardInOut      `json:"guards"`
	Actions     []actionInOut     `json:"actions"`
	Transitions []transitionInOut `json:"transitions"`
}

// ToRequirements converts the classInOut to class.Class.
func (c classInOut) ToRequirements() class.Class {
	cls := class.Class{
		Key:             c.Key,
		Name:            c.Name,
		Details:         c.Details,
		ActorKey:        c.ActorKey,
		SuperclassOfKey: c.SuperclassOfKey,
		SubclassOfKey:   c.SubclassOfKey,
		UmlComment:      c.UmlComment,
	}
	for _, a := range c.Attributes {
		cls.Attributes = append(cls.Attributes, a.ToRequirements())
	}
	for _, s := range c.States {
		cls.States = append(cls.States, s.ToRequirements())
	}
	for _, e := range c.Events {
		cls.Events = append(cls.Events, e.ToRequirements())
	}
	for _, g := range c.Guards {
		cls.Guards = append(cls.Guards, g.ToRequirements())
	}
	for _, ac := range c.Actions {
		cls.Actions = append(cls.Actions, ac.ToRequirements())
	}
	for _, t := range c.Transitions {
		cls.Transitions = append(cls.Transitions, t.ToRequirements())
	}
	return cls
}

// FromRequirements creates a classInOut from class.Class.
func FromRequirementsClass(c class.Class) classInOut {
	cls := classInOut{
		Key:             c.Key,
		Name:            c.Name,
		Details:         c.Details,
		ActorKey:        c.ActorKey,
		SuperclassOfKey: c.SuperclassOfKey,
		SubclassOfKey:   c.SubclassOfKey,
		UmlComment:      c.UmlComment,
	}
	for _, a := range c.Attributes {
		cls.Attributes = append(cls.Attributes, FromRequirementsAttribute(a))
	}
	for _, s := range c.States {
		cls.States = append(cls.States, FromRequirementsState(s))
	}
	for _, e := range c.Events {
		cls.Events = append(cls.Events, FromRequirementsEvent(e))
	}
	for _, g := range c.Guards {
		cls.Guards = append(cls.Guards, FromRequirementsGuard(g))
	}
	for _, ac := range c.Actions {
		cls.Actions = append(cls.Actions, FromRequirementsAction(ac))
	}
	for _, t := range c.Transitions {
		cls.Transitions = append(cls.Transitions, FromRequirementsTransition(t))
	}
	return cls
}
