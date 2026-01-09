package model_scenario

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
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
	FromObject *Object                `json:"-" yaml:"-"`
	ToObject   *Object                `json:"-" yaml:"-"`
	Event      *model_state.Event     `json:"-" yaml:"-"`
	Scenario   *Scenario              `json:"-" yaml:"-"`
	Attribute  *model_class.Attribute `json:"-" yaml:"-"`
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
	if n.EventKey != "" {
		return LEAF_TYPE_EVENT
	}
	if n.ScenarioKey != "" {
		return LEAF_TYPE_SCENARIO
	}
	if n.AttributeKey != "" {
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
			if n.FromObjectKey == "" {
				return errors.New("delete leaf must have a from_object_key")
			}
			if n.ToObjectKey != "" {
				return errors.New("delete leaf cannot have a to_object_key")
			}
			if n.EventKey != "" || n.ScenarioKey != "" || n.AttributeKey != "" {
				return errors.New("delete leaf cannot have event_key, scenario_key, or attribute_key")
			}
		} else {
			if n.FromObjectKey == "" {
				return errors.New("leaf must have a from_object_key")
			}
			if n.ToObjectKey == "" {
				return errors.New("leaf must have a to_object_key")
			}
			keys := []string{n.EventKey, n.ScenarioKey, n.AttributeKey}
			nonEmptyKeys := 0
			for _, key := range keys {
				if key != "" {
					nonEmptyKeys++
				}
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

// ScopeObjects prepends the object keys with the full path of the scenario to make them unique in the requirements.
func (n *Node) ScopeObjects(scenarioKey string) error {
	// Populate this node's references
	if n.FromObjectKey != "" {
		n.FromObjectKey = scenarioKey + "/object/" + n.FromObjectKey
	}
	if n.ToObjectKey != "" {
		n.ToObjectKey = scenarioKey + "/object/" + n.ToObjectKey
	}

	// Recursively populate references in statements
	if n.Statements != nil {
		for i := range n.Statements {
			if err := n.Statements[i].ScopeObjects(scenarioKey); err != nil {
				return err
			}
		}
	}

	// Recursively populate references in cases
	if n.Cases != nil {
		for i := range n.Cases {
			for j := range n.Cases[i].Statements {
				if err := n.Cases[i].Statements[j].ScopeObjects(scenarioKey); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// PopulateReferences populates the FromObject, ToObject, Event, and Scenario fields
// from the provided lookup maps. It recursively populates references in sub-nodes.
func (n *Node) PopulateReferences(objects map[string]Object, events map[string]model_state.Event, attributes map[string]model_class.Attribute, scenarios map[string]Scenario) error {
	// Populate this node's references
	if n.FromObjectKey != "" {
		if obj, exists := objects[n.FromObjectKey]; exists {
			n.FromObject = &obj
		} else {
			return errors.Errorf("from_object_key '%s' not found in objects", n.FromObjectKey)
		}
	}
	if n.ToObjectKey != "" {
		if obj, exists := objects[n.ToObjectKey]; exists {
			n.ToObject = &obj
		} else {
			return errors.Errorf("to_object_key '%s' not found in objects", n.ToObjectKey)
		}
	}
	if n.EventKey != "" {
		if evt, exists := events[n.EventKey]; exists {
			n.Event = &evt
		} else {
			return errors.Errorf("event_key '%s' not found in events", n.EventKey)
		}
	}
	if n.AttributeKey != "" {
		if attr, exists := attributes[n.AttributeKey]; exists {
			n.Attribute = &attr
		} else {
			return errors.Errorf("attribute_key '%s' not found in attributes", n.AttributeKey)
		}
	}
	if n.ScenarioKey != "" {
		if scen, exists := scenarios[n.ScenarioKey]; exists {
			n.Scenario = &scen
		} else {
			return errors.Errorf("scenario_key '%s' not found in scenarios", n.ScenarioKey)
		}
	}

	// Recursively populate references in statements
	if n.Statements != nil {
		for i := range n.Statements {
			if err := n.Statements[i].PopulateReferences(objects, events, attributes, scenarios); err != nil {
				return err
			}
		}
	}

	// Recursively populate references in cases
	if n.Cases != nil {
		for i := range n.Cases {
			for j := range n.Cases[i].Statements {
				if err := n.Cases[i].Statements[j].PopulateReferences(objects, events, attributes, scenarios); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// MarshalJSON custom marshals the Node to include the inferred type.
func (n *Node) MarshalJSON() ([]byte, error) {
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
	if n.FromObjectKey != "" {
		m["from_object_key"] = n.FromObjectKey
	}
	if n.ToObjectKey != "" {
		m["to_object_key"] = n.ToObjectKey
	}
	if n.EventKey != "" {
		m["event_key"] = n.EventKey
	}
	if n.AttributeKey != "" {
		m["attribute_key"] = n.AttributeKey
	}
	if n.ScenarioKey != "" {
		m["scenario_key"] = n.ScenarioKey
	}
	if n.IsDelete {
		m["is_delete"] = n.IsDelete
	}
	return json.Marshal(m)
}

// MarshalYAML custom marshals the Node to include the inferred type.
func (n *Node) MarshalYAML() (interface{}, error) {
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
	if n.FromObjectKey != "" {
		m["from_object_key"] = n.FromObjectKey
	}
	if n.ToObjectKey != "" {
		m["to_object_key"] = n.ToObjectKey
	}
	if n.EventKey != "" {
		m["event_key"] = n.EventKey
	}
	if n.AttributeKey != "" {
		m["attribute_key"] = n.AttributeKey
	}
	if n.ScenarioKey != "" {
		m["scenario_key"] = n.ScenarioKey
	}
	if n.IsDelete {
		m["is_delete"] = n.IsDelete
	}
	return m, nil
}
