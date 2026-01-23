package model_scenario

import (
	"encoding/json"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	// Node types.
	NODE_TYPE_LEAF     = "" // Leaf node has no type.
	NODE_TYPE_SEQUENCE = "sequence"
	NODE_TYPE_SWITCH   = "switch"
	NODE_TYPE_LOOP     = "loop"

	// Leaf types.
	LEAF_TYPE_EVENT     = "event"
	LEAF_TYPE_ATTRIBUTE = "attribute"
	LEAF_TYPE_SCENARIO  = "scenario"
	LEAF_TYPE_DELETE    = "delete"
)

// Node represents a node in the scenario steps tree.
type Node struct {
	Statements    []Node        `json:"statements,omitempty" yaml:"statements,omitempty"`
	Cases         []Case        `json:"cases,omitempty" yaml:"cases,omitempty"`
	Loop          string        `json:"loop,omitempty" yaml:"loop,omitempty"`               // Loop description.
	Description   string        `json:"description,omitempty" yaml:"description,omitempty"` // Leaf description.
	FromObjectKey *identity.Key `json:"from_object_key,omitempty" yaml:"from_object_key,omitempty"`
	ToObjectKey   *identity.Key `json:"to_object_key,omitempty" yaml:"to_object_key,omitempty"`
	EventKey      *identity.Key `json:"event_key,omitempty" yaml:"event_key,omitempty"`
	ScenarioKey   *identity.Key `json:"scenario_key,omitempty" yaml:"scenario_key,omitempty"`
	AttributeKey  *identity.Key `json:"attribute_key,omitempty" yaml:"attribute_key,omitempty"`
	IsDelete      bool          `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
}

// Inferredtype returns the type of the node based on its fields.
func (n *Node) Inferredtype() string {
	if n.Loop != "" {
		return NODE_TYPE_LOOP
	}
	if len(n.Cases) > 0 {
		return NODE_TYPE_SWITCH
	}
	if len(n.Statements) > 0 {
		return NODE_TYPE_SEQUENCE
	}
	return NODE_TYPE_LEAF
}

// InferredLeafType returns the leaf type of the node based on its fields.
func (n *Node) InferredLeafType() string {
	if n.EventKey != nil {
		return LEAF_TYPE_EVENT
	}
	if n.ScenarioKey != nil {
		return LEAF_TYPE_SCENARIO
	}
	if n.AttributeKey != nil {
		return LEAF_TYPE_ATTRIBUTE
	}
	if n.IsDelete {
		return LEAF_TYPE_DELETE
	}
	panic("node is not a leaf")
}

// Case represents a case in a switch node.
type Case struct {
	Condition  string `json:"condition" yaml:"condition"`
	Statements []Node `json:"statements" yaml:"statements"`
}

// Validate validates the node and its sub-nodes.
func (n *Node) Validate() error {
	switch n.Inferredtype() {
	case NODE_TYPE_SEQUENCE:
		if len(n.Statements) == 0 {
			return errors.New("sequence must have at least one statement")
		}
		for _, stmt := range n.Statements {
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case NODE_TYPE_SWITCH:
		if len(n.Cases) == 0 {
			return errors.New("switch must have at least one case")
		}
		for _, c := range n.Cases {
			if c.Condition == "" {
				return errors.New("switch case must have a conditional description")
			}
			for _, stmt := range c.Statements {
				if err := stmt.Validate(); err != nil {
					return err
				}
			}
		}
	case NODE_TYPE_LOOP:
		if n.Loop == "" {
			return errors.New("loop must have a loop description")
		}
		if len(n.Statements) == 0 {
			return errors.New("loop must have at least one statement")
		}
		for _, stmt := range n.Statements {
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case NODE_TYPE_LEAF:
		if n.IsDelete {
			if n.FromObjectKey == nil {
				return errors.New("delete leaf must have a from_object_key")
			}
			if n.ToObjectKey != nil {
				return errors.New("delete leaf cannot have a to_object_key")
			}
			if n.EventKey != nil || n.ScenarioKey != nil || n.AttributeKey != nil {
				return errors.New("delete leaf cannot have event_key, scenario_key, or attribute_key")
			}
		} else {
			if n.FromObjectKey == nil {
				return errors.New("leaf must have a from_object_key")
			}
			if n.ToObjectKey == nil {
				return errors.New("leaf must have a to_object_key")
			}
			nonEmptyKeys := 0
			if n.EventKey != nil {
				nonEmptyKeys++
			}
			if n.ScenarioKey != nil {
				nonEmptyKeys++
			}
			if n.AttributeKey != nil {
				nonEmptyKeys++
			}
			if nonEmptyKeys == 0 {
				return errors.New("leaf must have one of event_key, scenario_key, or attribute_key")
			}
			if nonEmptyKeys > 1 {
				return errors.New("leaf cannot have more than one of event_key, scenario_key, or attribute_key")
			}
		}
	}
	return nil
}

// ValidateWithParent validates the Node.
// Node has no key, so parent validation is not applicable.
func (n *Node) ValidateWithParent() error {
	return n.Validate()
}

// FromJSON parses the JSON string into the Node.
func (n *Node) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), n)
}

// ToJSON generates the JSON string from the Node.
func (n Node) ToJSON() (string, error) {
	data, err := json.Marshal(n)
	return string(data), err
}

// FromYAML parses the YAML string into the Node.
func (n *Node) FromYAML(yamlStr string) error {
	return yaml.Unmarshal([]byte(yamlStr), n)
}

// ToYAML generates the YAML string from the Node.
func (n Node) ToYAML() (string, error) {
	data, err := yaml.Marshal(n)
	return string(data), err
}

// MarshalJSON custom marshals the Node to include the inferred type.
// Uses value receiver so it works with both value and pointer types.
func (n Node) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if len(n.Statements) > 0 {
		m["statements"] = n.Statements
	}
	if len(n.Cases) > 0 {
		m["cases"] = n.Cases
	}
	if n.Loop != "" {
		m["loop"] = n.Loop
	}
	if n.Description != "" {
		m["description"] = n.Description
	}
	if n.FromObjectKey != nil {
		m["from_object_key"] = n.FromObjectKey
	}
	if n.ToObjectKey != nil {
		m["to_object_key"] = n.ToObjectKey
	}
	if n.EventKey != nil {
		m["event_key"] = n.EventKey
	}
	if n.AttributeKey != nil {
		m["attribute_key"] = n.AttributeKey
	}
	if n.ScenarioKey != nil {
		m["scenario_key"] = n.ScenarioKey
	}
	if n.IsDelete {
		m["is_delete"] = n.IsDelete
	}
	return json.Marshal(m)
}

// MarshalYAML custom marshals the Node to include the inferred type.
// Uses value receiver so it works with both value and pointer types.
func (n Node) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{})
	if len(n.Statements) > 0 {
		m["statements"] = n.Statements
	}
	if len(n.Cases) > 0 {
		m["cases"] = n.Cases
	}
	if n.Loop != "" {
		m["loop"] = n.Loop
	}
	if n.Description != "" {
		m["description"] = n.Description
	}
	if n.FromObjectKey != nil {
		m["from_object_key"] = n.FromObjectKey.String()
	}
	if n.ToObjectKey != nil {
		m["to_object_key"] = n.ToObjectKey.String()
	}
	if n.EventKey != nil {
		m["event_key"] = n.EventKey.String()
	}
	if n.AttributeKey != nil {
		m["attribute_key"] = n.AttributeKey.String()
	}
	if n.ScenarioKey != nil {
		m["scenario_key"] = n.ScenarioKey.String()
	}
	if n.IsDelete {
		m["is_delete"] = n.IsDelete
	}
	return m, nil
}
