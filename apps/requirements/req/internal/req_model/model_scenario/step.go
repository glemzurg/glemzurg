package model_scenario

import (
	"encoding/json"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	// Step types.
	STEP_TYPE_LEAF     = "leaf"
	STEP_TYPE_SEQUENCE = "sequence"
	STEP_TYPE_SWITCH   = "switch"
	STEP_TYPE_CASE     = "case"
	STEP_TYPE_LOOP     = "loop"

	// Leaf types.
	LEAF_TYPE_EVENT    = "event"
	LEAF_TYPE_QUERY    = "query"
	LEAF_TYPE_SCENARIO = "scenario"
	LEAF_TYPE_DELETE   = "delete"
)

// Step represents a step in the scenario steps tree.
type Step struct {
	Key           identity.Key  `json:"key" yaml:"key"`
	StepType      string        `json:"step_type" yaml:"step_type"`
	LeafType      *string       `json:"leaf_type,omitempty" yaml:"leaf_type,omitempty"`             // Only for leaf steps: event, query, scenario, delete.
	Statements    []Step        `json:"statements,omitempty" yaml:"statements,omitempty"`
	Condition     string        `json:"condition,omitempty" yaml:"condition,omitempty"`             // Used by loop and case steps.
	Description   string        `json:"description,omitempty" yaml:"description,omitempty"`         // Leaf description.
	FromObjectKey *identity.Key `json:"from_object_key,omitempty" yaml:"from_object_key,omitempty"` // Source object.
	ToObjectKey   *identity.Key `json:"to_object_key,omitempty" yaml:"to_object_key,omitempty"`     // Target object.
	EventKey      *identity.Key `json:"event_key,omitempty" yaml:"event_key,omitempty"`
	QueryKey      *identity.Key `json:"query_key,omitempty" yaml:"query_key,omitempty"`
	ScenarioKey   *identity.Key `json:"scenario_key,omitempty" yaml:"scenario_key,omitempty"`
}

// Validate validates the step and its sub-steps.
func (s *Step) Validate() error {
	// Validate the key.
	if err := s.Key.Validate(); err != nil {
		return err
	}
	if s.Key.KeyType != identity.KEY_TYPE_SCENARIO_STEP {
		return errors.Errorf("Key: invalid key type '%s' for scenario step.", s.Key.KeyType)
	}
	switch s.StepType {
	case STEP_TYPE_SEQUENCE:
		if len(s.Statements) == 0 {
			return errors.New("sequence must have at least one statement")
		}
		for _, stmt := range s.Statements {
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case STEP_TYPE_SWITCH:
		if len(s.Statements) == 0 {
			return errors.New("switch must have at least one case")
		}
		for _, stmt := range s.Statements {
			if stmt.StepType != STEP_TYPE_CASE {
				return errors.New("switch children must all be case steps")
			}
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case STEP_TYPE_CASE:
		if s.Condition == "" {
			return errors.New("case must have a condition")
		}
		for _, stmt := range s.Statements {
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case STEP_TYPE_LOOP:
		if s.Condition == "" {
			return errors.New("loop must have a condition")
		}
		if len(s.Statements) == 0 {
			return errors.New("loop must have at least one statement")
		}
		for _, stmt := range s.Statements {
			if err := stmt.Validate(); err != nil {
				return err
			}
		}
	case STEP_TYPE_LEAF:
		if s.LeafType == nil {
			return errors.New("leaf must have a leaf_type")
		}
		switch *s.LeafType {
		case LEAF_TYPE_DELETE:
			if s.FromObjectKey == nil {
				return errors.New("delete leaf must have a from_object_key")
			}
			if s.ToObjectKey != nil {
				return errors.New("delete leaf cannot have a to_object_key")
			}
			if s.EventKey != nil || s.ScenarioKey != nil || s.QueryKey != nil {
				return errors.New("delete leaf cannot have event_key, scenario_key, or query_key")
			}
		case LEAF_TYPE_EVENT:
			if s.FromObjectKey == nil {
				return errors.New("event leaf must have a from_object_key")
			}
			if s.ToObjectKey == nil {
				return errors.New("event leaf must have a to_object_key")
			}
			if s.EventKey == nil {
				return errors.New("event leaf must have an event_key")
			}
			if s.ScenarioKey != nil || s.QueryKey != nil {
				return errors.New("event leaf cannot have scenario_key or query_key")
			}
		case LEAF_TYPE_QUERY:
			if s.FromObjectKey == nil {
				return errors.New("query leaf must have a from_object_key")
			}
			if s.ToObjectKey == nil {
				return errors.New("query leaf must have a to_object_key")
			}
			if s.QueryKey == nil {
				return errors.New("query leaf must have a query_key")
			}
			if s.EventKey != nil || s.ScenarioKey != nil {
				return errors.New("query leaf cannot have event_key or scenario_key")
			}
		case LEAF_TYPE_SCENARIO:
			if s.FromObjectKey == nil {
				return errors.New("scenario leaf must have a from_object_key")
			}
			if s.ToObjectKey == nil {
				return errors.New("scenario leaf must have a to_object_key")
			}
			if s.ScenarioKey == nil {
				return errors.New("scenario leaf must have a scenario_key")
			}
			if s.EventKey != nil || s.QueryKey != nil {
				return errors.New("scenario leaf cannot have event_key or query_key")
			}
		default:
			return errors.Errorf("unknown leaf type '%s'", *s.LeafType)
		}
		// Validate key types of all non-nil reference keys.
		if s.FromObjectKey != nil {
			if err := s.FromObjectKey.Validate(); err != nil {
				return errors.Wrap(err, "FromObjectKey")
			}
			if s.FromObjectKey.KeyType != identity.KEY_TYPE_SCENARIO_OBJECT {
				return errors.Errorf("FromObjectKey: invalid key type '%s' for scenario object", s.FromObjectKey.KeyType)
			}
		}
		if s.ToObjectKey != nil {
			if err := s.ToObjectKey.Validate(); err != nil {
				return errors.Wrap(err, "ToObjectKey")
			}
			if s.ToObjectKey.KeyType != identity.KEY_TYPE_SCENARIO_OBJECT {
				return errors.Errorf("ToObjectKey: invalid key type '%s' for scenario object", s.ToObjectKey.KeyType)
			}
		}
		if s.EventKey != nil {
			if err := s.EventKey.Validate(); err != nil {
				return errors.Wrap(err, "EventKey")
			}
			if s.EventKey.KeyType != identity.KEY_TYPE_EVENT {
				return errors.Errorf("EventKey: invalid key type '%s' for event", s.EventKey.KeyType)
			}
		}
		if s.QueryKey != nil {
			if err := s.QueryKey.Validate(); err != nil {
				return errors.Wrap(err, "QueryKey")
			}
			if s.QueryKey.KeyType != identity.KEY_TYPE_QUERY {
				return errors.Errorf("QueryKey: invalid key type '%s' for query", s.QueryKey.KeyType)
			}
		}
		if s.ScenarioKey != nil {
			if err := s.ScenarioKey.Validate(); err != nil {
				return errors.Wrap(err, "ScenarioKey")
			}
			if s.ScenarioKey.KeyType != identity.KEY_TYPE_SCENARIO {
				return errors.Errorf("ScenarioKey: invalid key type '%s' for scenario", s.ScenarioKey.KeyType)
			}
		}
	default:
		return errors.Errorf("unknown step type '%s'", s.StepType)
	}
	return nil
}

// ValidateWithParent validates the Step and its key's parent relationship.
func (s *Step) ValidateWithParent(parent *identity.Key) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if err := s.Key.ValidateParent(parent); err != nil {
		return err
	}
	// A scenario leaf cannot reference the scenario that contains it.
	if s.ScenarioKey != nil && parent != nil && *s.ScenarioKey == *parent {
		return errors.New("scenario leaf cannot reference its own scenario")
	}
	// Validate children with the same parent (all steps are flat under the scenario).
	for i := range s.Statements {
		if err := s.Statements[i].ValidateWithParent(parent); err != nil {
			return err
		}
	}
	return nil
}

// FromJSON parses the JSON string into the Step.
func (s *Step) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), s)
}

// ToJSON generates the JSON string from the Step.
func (s Step) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	return string(data), err
}

// FromYAML parses the YAML string into the Step.
func (s *Step) FromYAML(yamlStr string) error {
	return yaml.Unmarshal([]byte(yamlStr), s)
}

// ToYAML generates the YAML string from the Step.
func (s Step) ToYAML() (string, error) {
	data, err := yaml.Marshal(s)
	return string(data), err
}

// MarshalJSON custom marshals the Step to only include non-empty fields.
// Uses value receiver so it works with both value and pointer types.
func (s Step) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	m["key"] = s.Key
	m["step_type"] = s.StepType
	if s.LeafType != nil {
		m["leaf_type"] = *s.LeafType
	}
	if len(s.Statements) > 0 {
		m["statements"] = s.Statements
	}
	if s.Condition != "" {
		m["condition"] = s.Condition
	}
	if s.Description != "" {
		m["description"] = s.Description
	}
	if s.FromObjectKey != nil {
		m["from_object_key"] = s.FromObjectKey
	}
	if s.ToObjectKey != nil {
		m["to_object_key"] = s.ToObjectKey
	}
	if s.EventKey != nil {
		m["event_key"] = s.EventKey
	}
	if s.QueryKey != nil {
		m["query_key"] = s.QueryKey
	}
	if s.ScenarioKey != nil {
		m["scenario_key"] = s.ScenarioKey
	}
	return json.Marshal(m)
}

// MarshalYAML custom marshals the Step to only include non-empty fields.
// Uses value receiver so it works with both value and pointer types.
func (s Step) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{})
	m["key"] = s.Key.String()
	m["step_type"] = s.StepType
	if s.LeafType != nil {
		m["leaf_type"] = *s.LeafType
	}
	if len(s.Statements) > 0 {
		m["statements"] = s.Statements
	}
	if s.Condition != "" {
		m["condition"] = s.Condition
	}
	if s.Description != "" {
		m["description"] = s.Description
	}
	if s.FromObjectKey != nil {
		m["from_object_key"] = s.FromObjectKey.String()
	}
	if s.ToObjectKey != nil {
		m["to_object_key"] = s.ToObjectKey.String()
	}
	if s.EventKey != nil {
		m["event_key"] = s.EventKey.String()
	}
	if s.QueryKey != nil {
		m["query_key"] = s.QueryKey.String()
	}
	if s.ScenarioKey != nil {
		m["scenario_key"] = s.ScenarioKey.String()
	}
	return m, nil
}
