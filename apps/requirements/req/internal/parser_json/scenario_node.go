package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
)

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

// isEmpty returns true if the nodeInOut is a zero value (empty).
func (n nodeInOut) isEmpty() bool {
	return len(n.Statements) == 0 &&
		len(n.Cases) == 0 &&
		n.Loop == "" &&
		n.Description == "" &&
		n.FromObjectKey == "" &&
		n.ToObjectKey == "" &&
		n.EventKey == "" &&
		n.ScenarioKey == "" &&
		n.AttributeKey == "" &&
		!n.IsDelete
}

// ToRequirements converts the nodeInOut to model_scenario.Node.
func (n nodeInOut) ToRequirements() (model_scenario.Node, error) {
	node := model_scenario.Node{
		Loop:        n.Loop,
		Description: n.Description,
		IsDelete:    n.IsDelete,
	}

	// Parse all key fields (all are now *identity.Key).
	if n.FromObjectKey != "" {
		key, err := identity.ParseKey(n.FromObjectKey)
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.FromObjectKey = &key
	}
	if n.ToObjectKey != "" {
		key, err := identity.ParseKey(n.ToObjectKey)
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.ToObjectKey = &key
	}
	if n.EventKey != "" {
		key, err := identity.ParseKey(n.EventKey)
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.EventKey = &key
	}
	if n.ScenarioKey != "" {
		key, err := identity.ParseKey(n.ScenarioKey)
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.ScenarioKey = &key
	}
	if n.AttributeKey != "" {
		key, err := identity.ParseKey(n.AttributeKey)
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.AttributeKey = &key
	}

	for _, s := range n.Statements {
		stmt, err := s.ToRequirements()
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.Statements = append(node.Statements, stmt)
	}

	for _, c := range n.Cases {
		caseNode, err := c.ToRequirements()
		if err != nil {
			return model_scenario.Node{}, err
		}
		node.Cases = append(node.Cases, caseNode)
	}

	return node, nil
}

// FromRequirementsNode creates a nodeInOut from model_scenario.Node.
func FromRequirementsNode(n model_scenario.Node) nodeInOut {
	node := nodeInOut{
		Loop:        n.Loop,
		Description: n.Description,
		IsDelete:    n.IsDelete,
	}

	// Convert all pointer-type keys to strings.
	if n.FromObjectKey != nil {
		node.FromObjectKey = n.FromObjectKey.String()
	}
	if n.ToObjectKey != nil {
		node.ToObjectKey = n.ToObjectKey.String()
	}
	if n.EventKey != nil {
		node.EventKey = n.EventKey.String()
	}
	if n.ScenarioKey != nil {
		node.ScenarioKey = n.ScenarioKey.String()
	}
	if n.AttributeKey != nil {
		node.AttributeKey = n.AttributeKey.String()
	}

	for _, s := range n.Statements {
		node.Statements = append(node.Statements, FromRequirementsNode(s))
	}
	for _, c := range n.Cases {
		node.Cases = append(node.Cases, FromRequirementsCase(c))
	}
	return node
}
