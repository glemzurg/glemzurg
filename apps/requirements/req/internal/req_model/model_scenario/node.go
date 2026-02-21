package model_scenario

import (
	"encoding/json"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	// Node types.
	NODE_TYPE_LEAF     = "leaf"
	NODE_TYPE_SEQUENCE = "sequence"
	NODE_TYPE_SWITCH   = "switch"
	NODE_TYPE_CASE     = "case"
	NODE_TYPE_LOOP     = "loop"

	// Leaf types.
	LEAF_TYPE_EVENT    = "event"
	LEAF_TYPE_QUERY    = "query"
	LEAF_TYPE_SCENARIO = "scenario"
	LEAF_TYPE_DELETE   = "delete"
)

// Node represents a node in the scenario steps tree.
type Node struct {
	Key           identity.Key  `json:"key" yaml:"key"`
	NodeType      string        `json:"node_type" yaml:"node_type"`
	Statements    []Node        `json:"statements,omitempty" yaml:"statements,omitempty"`
	Condition     string        `json:"condition,omitempty" yaml:"condition,omitempty"`             // Used by loop and case nodes.
	Description   string        `json:"description,omitempty" yaml:"description,omitempty"`         // Leaf description.
	FromObjectKey *identity.Key `json:"from_object_key,omitempty" yaml:"from_object_key,omitempty"` // Source object.
	ToObjectKey   *identity.Key `json:"to_object_key,omitempty" yaml:"to_object_key,omitempty"`     // Target object.
	EventKey      *identity.Key `json:"event_key,omitempty" yaml:"event_key,omitempty"`
	QueryKey      *identity.Key `json:"query_key,omitempty" yaml:"query_key,omitempty"`
	ScenarioKey   *identity.Key `json:"scenario_key,omitempty" yaml:"scenario_key,omitempty"`
	IsDelete      bool          `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
}

// InferredLeafType returns the leaf type of the node based on its fields.
func (n *Node) InferredLeafType() string {
	if n.EventKey != nil {
		return LEAF_TYPE_EVENT
	}
	if n.ScenarioKey != nil {
		return LEAF_TYPE_SCENARIO
	}
	if n.QueryKey != nil {
		return LEAF_TYPE_QUERY
	}
	if n.IsDelete {
		return LEAF_TYPE_DELETE
	}
	panic("node is not a leaf")
}

// Validate validates the node and its sub-nodes.
func (n *Node) Validate() error {
	// Validate the key.
	if err := n.Key.Validate(); err != nil {
		return err
	}
	if n.Key.KeyType != identity.KEY_TYPE_SCENARIO_STEP {
		return errors.Errorf("Key: invalid key type '%s' for scenario step.", n.Key.KeyType)
	}
	switch n.NodeType {
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
		if len(n.Statements) == 0 {
			return errors.New("switch must have at least one case")
		}
		for _, stmt := range n.Statements {
			if stmt.NodeType != NODE_TYPE_CASE {
				return errors.New("switch children must all be case nodes")
			}
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case NODE_TYPE_CASE:
		if n.Condition == "" {
			return errors.New("case must have a condition")
		}
		for _, stmt := range n.Statements {
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case NODE_TYPE_LOOP:
		if n.Condition == "" {
			return errors.New("loop must have a condition")
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
			if n.EventKey != nil || n.ScenarioKey != nil || n.QueryKey != nil {
				return errors.New("delete leaf cannot have event_key, scenario_key, or query_key")
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
			if n.QueryKey != nil {
				nonEmptyKeys++
			}
			if nonEmptyKeys == 0 {
				return errors.New("leaf must have one of event_key, scenario_key, or query_key")
			}
			if nonEmptyKeys > 1 {
				return errors.New("leaf cannot have more than one of event_key, scenario_key, or query_key")
			}
		}
	default:
		return errors.Errorf("unknown node type '%s'", n.NodeType)
	}
	return nil
}

// ValidateWithParent validates the Node and its key's parent relationship.
func (n *Node) ValidateWithParent(parent *identity.Key) error {
	if err := n.Validate(); err != nil {
		return err
	}
	if err := n.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate children with the same parent (all steps are flat under the scenario).
	for i := range n.Statements {
		if err := n.Statements[i].ValidateWithParent(parent); err != nil {
			return err
		}
	}
	return nil
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

// MarshalJSON custom marshals the Node to only include non-empty fields.
// Uses value receiver so it works with both value and pointer types.
func (n Node) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	m["key"] = n.Key
	m["node_type"] = n.NodeType
	if len(n.Statements) > 0 {
		m["statements"] = n.Statements
	}
	if n.Condition != "" {
		m["condition"] = n.Condition
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
	if n.QueryKey != nil {
		m["query_key"] = n.QueryKey
	}
	if n.ScenarioKey != nil {
		m["scenario_key"] = n.ScenarioKey
	}
	if n.IsDelete {
		m["is_delete"] = n.IsDelete
	}
	return json.Marshal(m)
}

// MarshalYAML custom marshals the Node to only include non-empty fields.
// Uses value receiver so it works with both value and pointer types.
func (n Node) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{})
	m["key"] = n.Key.String()
	m["node_type"] = n.NodeType
	if len(n.Statements) > 0 {
		m["statements"] = n.Statements
	}
	if n.Condition != "" {
		m["condition"] = n.Condition
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
	if n.QueryKey != nil {
		m["query_key"] = n.QueryKey.String()
	}
	if n.ScenarioKey != nil {
		m["scenario_key"] = n.ScenarioKey.String()
	}
	if n.IsDelete {
		m["is_delete"] = n.IsDelete
	}
	return m, nil
}
