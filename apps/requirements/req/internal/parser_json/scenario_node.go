package parser_json

// nodeInOut represents a node in the scenario steps tree.
type nodeInOut struct {
	Statements    []nodeInOut `json:"statements,omitempty" yaml:"statements,omitempty"`
	Cases         []caseInOut `json:"cases,omitempty" yaml:"cases,omitempty"`
	Loop          string      `json:"loop,omitempty" yaml:"loop,omitempty"`               // Loop description.
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"` // Leaf description.
	FromObjectKey string      `json:"from_object_key,omitempty" yaml:"from_object_key,omitempty"`
	ToObjectKey   string      `json:"to_object_key,omitempty" yaml:"to_object_key,omitempty"`
	EventKey      string      `json:"event_key,omitempty" yaml:"event_key,omitempty"`
	ScenarioKey   string      `json:"scenario_key,omitempty" yaml:"scenario_key,omitempty"`
	AttributeKey  string      `json:"attribute_key,omitempty" yaml:"attribute_key,omitempty"`
	IsDelete      bool        `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
}
