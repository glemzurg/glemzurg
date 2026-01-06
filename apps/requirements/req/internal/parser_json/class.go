package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

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

// ToRequirements converts the classInOut to requirements.Class.
func (c classInOut) ToRequirements() requirements.Class {
	class := requirements.Class{
		Key:             c.Key,
		Name:            c.Name,
		Details:         c.Details,
		ActorKey:        c.ActorKey,
		SuperclassOfKey: c.SuperclassOfKey,
		SubclassOfKey:   c.SubclassOfKey,
		UmlComment:      c.UmlComment,
	}
	for _, a := range c.Attributes {
		class.Attributes = append(class.Attributes, a.ToRequirements())
	}
	for _, s := range c.States {
		class.States = append(class.States, s.ToRequirements())
	}
	for _, e := range c.Events {
		class.Events = append(class.Events, e.ToRequirements())
	}
	for _, g := range c.Guards {
		class.Guards = append(class.Guards, g.ToRequirements())
	}
	for _, ac := range c.Actions {
		class.Actions = append(class.Actions, ac.ToRequirements())
	}
	for _, t := range c.Transitions {
		class.Transitions = append(class.Transitions, t.ToRequirements())
	}
	return class
}

// FromRequirements creates a classInOut from requirements.Class.
func FromRequirementsClass(c requirements.Class) classInOut {
	class := classInOut{
		Key:             c.Key,
		Name:            c.Name,
		Details:         c.Details,
		ActorKey:        c.ActorKey,
		SuperclassOfKey: c.SuperclassOfKey,
		SubclassOfKey:   c.SubclassOfKey,
		UmlComment:      c.UmlComment,
	}
	for _, a := range c.Attributes {
		class.Attributes = append(class.Attributes, FromRequirementsAttribute(a))
	}
	for _, s := range c.States {
		class.States = append(class.States, FromRequirementsState(s))
	}
	for _, e := range c.Events {
		class.Events = append(class.Events, FromRequirementsEvent(e))
	}
	for _, g := range c.Guards {
		class.Guards = append(class.Guards, FromRequirementsGuard(g))
	}
	for _, ac := range c.Actions {
		class.Actions = append(class.Actions, FromRequirementsAction(ac))
	}
	for _, t := range c.Transitions {
		class.Transitions = append(class.Transitions, FromRequirementsTransition(t))
	}
	return class
}
