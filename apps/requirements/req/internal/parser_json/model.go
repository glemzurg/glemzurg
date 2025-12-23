package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
)

// Model is the documentation summary of a set of requirements.
type Model struct {
	Key     string
	Name    string
	Details string // Markdown.
	// Nested structure.
	Actors             []Actor
	Domains            []Domain
	DomainAssociations []DomainAssociation
	Associations       []Association // Associations between classes that span domains.
}

// An actor is a external user of this system, either a person or another system.
type Actor struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Type       string // "person" or "system"
	UmlComment string
}

// Domain is a root category of the model.
type Domain struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Realized   bool   // If this domain has no semantic model because it is existing already, so only design in this domain.
	UmlComment string
	// Nested.
	Subdomains []Subdomain
}

// Subdomain is a nested category of the model.
type Subdomain struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Nested.
	Generalizations []Generalization // Generalizations for the classes and use cases in this subdomain.
	Classes         []Class
	UseCases        []UseCase
	Associations    []Association // Associations between classes in this subdomain.
}

// Class is a thing in the system.
type Class struct {
	Key             string
	Name            string
	Details         string // Markdown.
	ActorKey        string // If this class is an Actor this is the key of that actor.
	SuperclassOfKey string // If this class is part of a generalization as the superclass.
	SubclassOfKey   string // If this class is part of a generalization as a subclass.
	UmlComment      string
	// Nested.
	Attributes  []Attribute
	States      []State
	Events      []Event
	Guards      []Guard
	Actions     []Action
	Transitions []Transition
}

// Attribute is a member of a class.
type Attribute struct {
	Key              string
	Name             string
	Details          string // Markdown.
	DataTypeRules    string // What are the bounds of this data type.
	DerivationPolicy string // If this is a derived attribute, how is it derived.
	Nullable         bool   // Is this attribute optional.
	UmlComment       string
	// Part of the data in a parsed file.
	IndexNums []uint              // The indexes this attribute is part of.
	DataType  *data_type.DataType // If the DataTypeRules can be parsed, this is the resulting data type.
}

// Association is how two classes relate to each other.
type Association struct {
	Key                 string
	Name                string
	Details             string       // Markdown.
	FromClassKey        string       // The class on one end of the association.
	FromMultiplicity    Multiplicity // The multiplicity from one end of the association.
	ToClassKey          string       // The class on the other end of the association.
	ToMultiplicity      Multiplicity // The multiplicity on the other end of the association.
	AssociationClassKey string       // Any class that points to this association.
	UmlComment          string
}

// Multiplicity is how two classes relate to each other.
type Multiplicity struct {
	LowerBound  uint // Zero is "any".
	HigherBound uint // Zero is "any".
}

// State is a particular set of values in a state, distinct from all other states in the state.
type State struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Nested.
	Actions []StateAction
}

// Event is what triggers a transition between states.
type Event struct {
	Key        string
	Name       string
	Details    string
	Parameters []EventParameter
}

// EventParameter is a parameter for events.
type EventParameter struct {
	Name   string
	Source string // Where the values for this parameter are coming from.
}

// Guard is a constraint on an event in a state machine.
type Guard struct {
	Key     string
	Name    string // A simple unique name for a guard, for internal use.
	Details string // How the details of the guard are represented, what shows in the uml.
}

// Action is what happens in a transition between states.
type Action struct {
	Key        string
	Name       string
	Details    string
	Requires   []string // To enter this action.
	Guarantees []string
}

// Transition is a move between two states.
type Transition struct {
	Key          string
	FromStateKey string
	EventKey     string
	GuardKey     string
	ActionKey    string
	ToStateKey   string
	UmlComment   string
}

// StateAction is a action that triggers when a state is entered or exited or happens perpetually.
type StateAction struct {
	Key       string
	ActionKey string
	When      string
}

// UseCase is a user story for the system.
type UseCase struct {
	Key        string
	Name       string
	Details    string // Markdown.
	Level      string // How high cocept or tightly focused the user case is.
	ReadOnly   bool   // This is a user story that does not change the state of the system.
	UmlComment string
	// Nested.
	Actors    map[string]UseCaseActor
	Scenarios []Scenario
	// Helpful data.
	DomainKey string
}

// UseCaseActor is an actor who acts in a user story.
type UseCaseActor struct {
	UmlComment string
}

// Scenario is a documented scenario for a use case, such as a sequence diagram.
type Scenario struct {
	Key     string
	Name    string
	Details string      // Markdown.
	Steps   interface{} // The "abstract syntax tree" of the scenario. (Node type, simplified)
	// Nested.
	Objects []ScenarioObject
}

// ScenarioObject is an object that participates in a scenario.
type ScenarioObject struct {
	Key          string
	ObjectNumber uint   // Order in the scenario diagram.
	Name         string // The name or id of the object.
	NameStyle    string // Used to format the name in the diagram.
	ClassKey     string // The class key this object is an instance of.
	Multi        bool
	UmlComment   string
}

// Generalization is how two or more things in the system build on each other (like a super type and sub type).
type Generalization struct {
	Key        string
	Name       string
	Details    string // Markdown.
	IsComplete bool   // Are the specializations complete, or can an instantiation of this generalization exist without a specialization.
	IsStatic   bool   // Are the specializations static and unchanging or can they change during runtime.
	UmlComment string
}

// DomainAssociation is when a domain enforces requirements on another domain.
type DomainAssociation struct {
	Key               string // The key of unique in the model.
	ProblemDomainKey  string // The domain that enforces requirements on the other domain.
	SolutionDomainKey string // The domain that has requirements enforced upon it.
	UmlComment        string
}
