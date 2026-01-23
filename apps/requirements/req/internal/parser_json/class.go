package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
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

// ToRequirements converts the classInOut to model_class.Class.
func (c classInOut) ToRequirements() (model_class.Class, error) {
	key, err := identity.ParseKey(c.Key)
	if err != nil {
		return model_class.Class{}, err
	}

	// Handle optional pointer fields - empty string means nil
	var actorKey *identity.Key
	if c.ActorKey != "" {
		k, err := identity.ParseKey(c.ActorKey)
		if err != nil {
			return model_class.Class{}, err
		}
		actorKey = &k
	}

	var superclassOfKey *identity.Key
	if c.SuperclassOfKey != "" {
		k, err := identity.ParseKey(c.SuperclassOfKey)
		if err != nil {
			return model_class.Class{}, err
		}
		superclassOfKey = &k
	}

	var subclassOfKey *identity.Key
	if c.SubclassOfKey != "" {
		k, err := identity.ParseKey(c.SubclassOfKey)
		if err != nil {
			return model_class.Class{}, err
		}
		subclassOfKey = &k
	}

	class := model_class.Class{
		Key:             key,
		Name:            c.Name,
		Details:         c.Details,
		ActorKey:        actorKey,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
		UmlComment:      c.UmlComment,
	}
	for _, a := range c.Attributes {
		attr, err := a.ToRequirements()
		if err != nil {
			return model_class.Class{}, err
		}
		if class.Attributes == nil {
			class.Attributes = make(map[identity.Key]model_class.Attribute)
		}
		class.Attributes[attr.Key] = attr
	}
	for _, s := range c.States {
		state, err := s.ToRequirements()
		if err != nil {
			return model_class.Class{}, err
		}
		if class.States == nil {
			class.States = make(map[identity.Key]model_state.State)
		}
		class.States[state.Key] = state
	}
	for _, e := range c.Events {
		event, err := e.ToRequirements()
		if err != nil {
			return model_class.Class{}, err
		}
		if class.Events == nil {
			class.Events = make(map[identity.Key]model_state.Event)
		}
		class.Events[event.Key] = event
	}
	for _, g := range c.Guards {
		guard, err := g.ToRequirements()
		if err != nil {
			return model_class.Class{}, err
		}
		if class.Guards == nil {
			class.Guards = make(map[identity.Key]model_state.Guard)
		}
		class.Guards[guard.Key] = guard
	}
	for _, ac := range c.Actions {
		action, err := ac.ToRequirements()
		if err != nil {
			return model_class.Class{}, err
		}
		if class.Actions == nil {
			class.Actions = make(map[identity.Key]model_state.Action)
		}
		class.Actions[action.Key] = action
	}
	for _, t := range c.Transitions {
		transition, err := t.ToRequirements()
		if err != nil {
			return model_class.Class{}, err
		}
		if class.Transitions == nil {
			class.Transitions = make(map[identity.Key]model_state.Transition)
		}
		class.Transitions[transition.Key] = transition
	}
	return class, nil
}

// FromRequirements creates a classInOut from model_class.Class.
func FromRequirementsClass(c model_class.Class) classInOut {
	// Handle optional pointer fields - nil means empty string
	var actorKey, superclassOfKey, subclassOfKey string
	if c.ActorKey != nil {
		actorKey = c.ActorKey.String()
	}
	if c.SuperclassOfKey != nil {
		superclassOfKey = c.SuperclassOfKey.String()
	}
	if c.SubclassOfKey != nil {
		subclassOfKey = c.SubclassOfKey.String()
	}

	class := classInOut{
		Key:             c.Key.String(),
		Name:            c.Name,
		Details:         c.Details,
		ActorKey:        actorKey,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
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
