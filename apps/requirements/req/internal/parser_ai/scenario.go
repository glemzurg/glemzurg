package parser_ai

// inputObject represents an object participating in a scenario.
type inputObject struct {
	ObjectNumber uint   `json:"object_number"`
	Name         string `json:"name,omitempty"`
	NameStyle    string `json:"name_style"`
	ClassKey     string `json:"class_key"`
	Multi        bool   `json:"multi,omitempty"`
	UmlComment   string `json:"uml_comment,omitempty"`
}

// inputStep represents a step in a scenario's step tree.
// Steps form a recursive tree structure (AST) representing the scenario flow.
type inputStep struct {
	StepType      string      `json:"step_type"`
	LeafType      *string     `json:"leaf_type,omitempty"`
	Statements    []inputStep `json:"statements,omitempty"`
	Condition     string      `json:"condition,omitempty"`
	Description   string      `json:"description,omitempty"`
	FromObjectKey *string     `json:"from_object_key,omitempty"`
	ToObjectKey   *string     `json:"to_object_key,omitempty"`
	EventKey      *string     `json:"event_key,omitempty"`
	QueryKey      *string     `json:"query_key,omitempty"`
	ScenarioKey   *string     `json:"scenario_key,omitempty"`
}

// inputScenario represents a scenario JSON file.
// Scenarios document specific flows through a use case (e.g., sequence diagrams).
type inputScenario struct {
	Name    string                   `json:"name"`
	Details string                   `json:"details,omitempty"`
	Objects map[string]*inputObject  `json:"objects,omitempty"`
	Steps   *inputStep               `json:"steps,omitempty"`
}
