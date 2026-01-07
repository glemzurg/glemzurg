package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"

// nodeInOut represents a node in the scenario steps tree.
type nodeInOut struct {
	Statements    []nodeInOut `json:"statements" yaml:"statements"`
	Cases         []caseInOut `json:"cases" yaml:"cases"`
	Loop          string      `json:"loop" yaml:"loop"`               // Loop description.
	Description   string      `json:"description" yaml:"description"` // Leaf description.
	FromObjectKey string      `json:"from_object_key" yaml:"from_object_key"`
	ToObjectKey   string      `json:"to_object_key" yaml:"to_object_key"`
	EventKey      string      `json:"event_key" yaml:"event_key"`
	ScenarioKey   string      `json:"scenario_key" yaml:"scenario_key"`
	AttributeKey  string      `json:"attribute_key" yaml:"attribute_key"`
	IsDelete      bool        `json:"is_delete" yaml:"is_delete"`
}

// ToRequirements converts the nodeInOut to model_scenario.Node.
func (n nodeInOut) ToRequirements() model_scenario.Node {

	node := model_scenario.Node{
		Statements:    nil,
		Cases:         nil,
		Loop:          n.Loop,
		Description:   n.Description,
		FromObjectKey: n.FromObjectKey,
		ToObjectKey:   n.ToObjectKey,
		EventKey:      n.EventKey,
		ScenarioKey:   n.ScenarioKey,
		AttributeKey:  n.AttributeKey,
		IsDelete:      n.IsDelete,
	}

	for _, s := range n.Statements {
		node.Statements = append(node.Statements, s.ToRequirements())
	}

	for _, c := range n.Cases {
		node.Cases = append(node.Cases, c.ToRequirements())
	}

	return node
}

// FromRequirementsNode creates a nodeInOut from model_scenario.Node.
func FromRequirementsNode(n model_scenario.Node) nodeInOut {

	node := nodeInOut{
		Statements:    nil,
		Cases:         nil,
		Loop:          n.Loop,
		Description:   n.Description,
		FromObjectKey: n.FromObjectKey,
		ToObjectKey:   n.ToObjectKey,
		EventKey:      n.EventKey,
		ScenarioKey:   n.ScenarioKey,
		AttributeKey:  n.AttributeKey,
		IsDelete:      n.IsDelete,
	}

	for _, s := range n.Statements {
		node.Statements = append(node.Statements, FromRequirementsNode(s))
	}
	for _, c := range n.Cases {
		node.Cases = append(node.Cases, FromRequirementsCase(c))
	}
	return node
}
