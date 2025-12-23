package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

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

// ToRequirements converts the nodeInOut to requirements.Node.
func (n nodeInOut) ToRequirements() requirements.Node {
	statements := make([]requirements.Node, len(n.Statements))
	for i, s := range n.Statements {
		statements[i] = s.ToRequirements()
	}
	cases := make([]requirements.Case, len(n.Cases))
	for i, c := range n.Cases {
		cases[i] = c.ToRequirements()
	}
	return requirements.Node{
		Statements:    statements,
		Cases:         cases,
		Loop:          n.Loop,
		Description:   n.Description,
		FromObjectKey: n.FromObjectKey,
		ToObjectKey:   n.ToObjectKey,
		EventKey:      n.EventKey,
		ScenarioKey:   n.ScenarioKey,
		AttributeKey:  n.AttributeKey,
		IsDelete:      n.IsDelete,
	}
}

// FromRequirementsNode creates a nodeInOut from requirements.Node.
func FromRequirementsNode(n requirements.Node) nodeInOut {
	statements := make([]nodeInOut, len(n.Statements))
	for i, s := range n.Statements {
		statements[i] = FromRequirementsNode(s)
	}
	cases := make([]caseInOut, len(n.Cases))
	for i, c := range n.Cases {
		cases[i] = FromRequirementsCase(c)
	}
	return nodeInOut{
		Statements:    statements,
		Cases:         cases,
		Loop:          n.Loop,
		Description:   n.Description,
		FromObjectKey: n.FromObjectKey,
		ToObjectKey:   n.ToObjectKey,
		EventKey:      n.EventKey,
		ScenarioKey:   n.ScenarioKey,
		AttributeKey:  n.AttributeKey,
		IsDelete:      n.IsDelete,
	}
}
