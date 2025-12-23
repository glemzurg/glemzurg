package parser_json

// node represents a node in the scenario steps tree.
type node struct {
	Statements    []node  `json:"statements,omitempty" yaml:"statements,omitempty"`
	Cases         []case_ `json:"cases,omitempty" yaml:"cases,omitempty"`
	Loop          string  `json:"loop,omitempty" yaml:"loop,omitempty"`               // Loop description.
	Description   string  `json:"description,omitempty" yaml:"description,omitempty"` // Leaf description.
	FromObjectKey string  `json:"from_object_key,omitempty" yaml:"from_object_key,omitempty"`
	ToObjectKey   string  `json:"to_object_key,omitempty" yaml:"to_object_key,omitempty"`
	EventKey      string  `json:"event_key,omitempty" yaml:"event_key,omitempty"`
	ScenarioKey   string  `json:"scenario_key,omitempty" yaml:"scenario_key,omitempty"`
	AttributeKey  string  `json:"attribute_key,omitempty" yaml:"attribute_key,omitempty"`
	IsDelete      bool    `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
	// Helper fields can be added here as needed.
	FromObject *scenarioObject `json:"-" yaml:"-"`
	ToObject   *scenarioObject `json:"-" yaml:"-"`
	Event      *event          `json:"-" yaml:"-"`
	Scenario   *scenario       `json:"-" yaml:"-"`
	Attribute  *attribute      `json:"-" yaml:"-"`
}
