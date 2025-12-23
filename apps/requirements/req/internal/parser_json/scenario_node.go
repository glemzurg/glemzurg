package parser_json

// Node represents a node in the scenario steps tree.
type Node struct {
	Statements    []Node `json:"statements,omitempty" yaml:"statements,omitempty"`
	Cases         []Case `json:"cases,omitempty" yaml:"cases,omitempty"`
	Loop          string `json:"loop,omitempty" yaml:"loop,omitempty"`               // Loop description.
	Description   string `json:"description,omitempty" yaml:"description,omitempty"` // Leaf description.
	FromObjectKey string `json:"from_object_key,omitempty" yaml:"from_object_key,omitempty"`
	ToObjectKey   string `json:"to_object_key,omitempty" yaml:"to_object_key,omitempty"`
	EventKey      string `json:"event_key,omitempty" yaml:"event_key,omitempty"`
	ScenarioKey   string `json:"scenario_key,omitempty" yaml:"scenario_key,omitempty"`
	AttributeKey  string `json:"attribute_key,omitempty" yaml:"attribute_key,omitempty"`
	IsDelete      bool   `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
	// Helper fields can be added here as needed.
	FromObject *ScenarioObject `json:"-" yaml:"-"`
	ToObject   *ScenarioObject `json:"-" yaml:"-"`
	Event      *Event          `json:"-" yaml:"-"`
	Scenario   *Scenario       `json:"-" yaml:"-"`
	Attribute  *Attribute      `json:"-" yaml:"-"`
}

// Case represents a case in a switch node.
type Case struct {
	Condition  string `json:"condition" yaml:"condition"`
	Statements []Node `json:"statements" yaml:"statements"`
}
