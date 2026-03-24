package model_scenario

import (
	"encoding/json"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
type Step struct { //nolint:recvcheck
	Key           identity.Key  `json:"key" yaml:"key"`
	StepType      string        `json:"step_type" yaml:"step_type"`
	LeafType      *string       `json:"leaf_type,omitempty" yaml:"leaf_type,omitempty"` // Only for leaf steps: event, query, scenario, delete.
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
//
//complexity:cyclo:warn=60,fail=60 Simple routing switch.
func (s *Step) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := s.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.SstepKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if s.Key.KeyType != identity.KEY_TYPE_SCENARIO_STEP {
		return coreerr.NewWithValues(ctx, coreerr.SstepKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for scenario step", s.Key.KeyType), "Key", s.Key.KeyType, identity.KEY_TYPE_SCENARIO_STEP)
	}
	switch s.StepType {
	case STEP_TYPE_SEQUENCE:
		if len(s.Statements) == 0 {
			return coreerr.New(ctx, coreerr.SstepSequenceMinStatements, "sequence must have at least one statement", "Statements")
		}
		for i := range s.Statements {
			childCtx := ctx.Child("statement", fmt.Sprintf("%d", i))
			if err := s.Statements[i].Validate(childCtx); err != nil {
				return err
			}
		}
	case STEP_TYPE_SWITCH:
		if len(s.Statements) == 0 {
			return coreerr.New(ctx, coreerr.SstepSwitchMinCases, "switch must have at least one case", "Statements")
		}
		for i := range s.Statements {
			if s.Statements[i].StepType != STEP_TYPE_CASE {
				return coreerr.NewWithValues(ctx, coreerr.SstepSwitchCaseType, "switch children must all be case steps", "Statements", s.Statements[i].StepType, STEP_TYPE_CASE)
			}
			childCtx := ctx.Child("case", fmt.Sprintf("%d", i))
			if err := s.Statements[i].Validate(childCtx); err != nil {
				return err
			}
		}
	case STEP_TYPE_CASE:
		if s.Condition == "" {
			return coreerr.New(ctx, coreerr.SstepCaseConditionRequired, "case must have a condition", "Condition")
		}
		for i := range s.Statements {
			childCtx := ctx.Child("statement", fmt.Sprintf("%d", i))
			if err := s.Statements[i].Validate(childCtx); err != nil {
				return err
			}
		}
	case STEP_TYPE_LOOP:
		if s.Condition == "" {
			return coreerr.New(ctx, coreerr.SstepLoopConditionRequired, "loop must have a condition", "Condition")
		}
		if len(s.Statements) == 0 {
			return coreerr.New(ctx, coreerr.SstepLoopMinStatements, "loop must have at least one statement", "Statements")
		}
		for i := range s.Statements {
			childCtx := ctx.Child("statement", fmt.Sprintf("%d", i))
			if err := s.Statements[i].Validate(childCtx); err != nil {
				return err
			}
		}
	case STEP_TYPE_LEAF:
		if s.LeafType == nil {
			return coreerr.New(ctx, coreerr.SstepLeafTypeRequired, "leaf must have a leaf_type", "LeafType")
		}
		switch *s.LeafType {
		case LEAF_TYPE_DELETE:
			if s.FromObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepDeleteFromRequired, "delete leaf must have a from_object_key", "FromObjectKey")
			}
			if s.ToObjectKey != nil {
				return coreerr.New(ctx, coreerr.SstepDeleteToForbidden, "delete leaf cannot have a to_object_key", "ToObjectKey")
			}
			if s.EventKey != nil || s.ScenarioKey != nil || s.QueryKey != nil {
				return coreerr.New(ctx, coreerr.SstepDeleteKeysForbidden, "delete leaf cannot have event_key, scenario_key, or query_key", "EventKey/ScenarioKey/QueryKey")
			}
		case LEAF_TYPE_EVENT:
			if s.FromObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepEventFromRequired, "event leaf must have a from_object_key", "FromObjectKey")
			}
			if s.ToObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepEventToRequired, "event leaf must have a to_object_key", "ToObjectKey")
			}
			if s.EventKey == nil {
				return coreerr.New(ctx, coreerr.SstepEventKeyRequired, "event leaf must have an event_key", "EventKey")
			}
			if s.ScenarioKey != nil || s.QueryKey != nil {
				return coreerr.New(ctx, coreerr.SstepEventQueryForbidden, "event leaf cannot have scenario_key or query_key", "ScenarioKey/QueryKey")
			}
		case LEAF_TYPE_QUERY:
			if s.FromObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepQueryFromRequired, "query leaf must have a from_object_key", "FromObjectKey")
			}
			if s.ToObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepQueryToRequired, "query leaf must have a to_object_key", "ToObjectKey")
			}
			if s.QueryKey == nil {
				return coreerr.New(ctx, coreerr.SstepQueryKeyRequired, "query leaf must have a query_key", "QueryKey")
			}
			if s.EventKey != nil || s.ScenarioKey != nil {
				return coreerr.New(ctx, coreerr.SstepQueryEventForbidden, "query leaf cannot have event_key or scenario_key", "EventKey/ScenarioKey")
			}
		case LEAF_TYPE_SCENARIO:
			if s.FromObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepScenarioFromRequired, "scenario leaf must have a from_object_key", "FromObjectKey")
			}
			if s.ToObjectKey == nil {
				return coreerr.New(ctx, coreerr.SstepScenarioToRequired, "scenario leaf must have a to_object_key", "ToObjectKey")
			}
			if s.ScenarioKey == nil {
				return coreerr.New(ctx, coreerr.SstepScenarioKeyRequired, "scenario leaf must have a scenario_key", "ScenarioKey")
			}
			if s.EventKey != nil || s.QueryKey != nil {
				return coreerr.New(ctx, coreerr.SstepScenarioEventForbidden, "scenario leaf cannot have event_key or query_key", "EventKey/QueryKey")
			}
		default:
			return coreerr.NewWithValues(ctx, coreerr.SstepLeafTypeUnknown, fmt.Sprintf("unknown leaf type '%s'", *s.LeafType), "LeafType", *s.LeafType, "one of: event, query, scenario, delete")
		}
		// Validate key types of all non-nil reference keys.
		if s.FromObjectKey != nil {
			if err := s.FromObjectKey.ValidateWithContext(ctx); err != nil {
				return coreerr.New(ctx, coreerr.SstepFromkeyInvalid, fmt.Sprintf("FromObjectKey: %s", err.Error()), "FromObjectKey")
			}
			if s.FromObjectKey.KeyType != identity.KEY_TYPE_SCENARIO_OBJECT {
				return coreerr.NewWithValues(ctx, coreerr.SstepFromkeyTypeInvalid, fmt.Sprintf("FromObjectKey: invalid key type '%s' for scenario object", s.FromObjectKey.KeyType), "FromObjectKey", s.FromObjectKey.KeyType, identity.KEY_TYPE_SCENARIO_OBJECT)
			}
		}
		if s.ToObjectKey != nil {
			if err := s.ToObjectKey.ValidateWithContext(ctx); err != nil {
				return coreerr.New(ctx, coreerr.SstepTokeyInvalid, fmt.Sprintf("ToObjectKey: %s", err.Error()), "ToObjectKey")
			}
			if s.ToObjectKey.KeyType != identity.KEY_TYPE_SCENARIO_OBJECT {
				return coreerr.NewWithValues(ctx, coreerr.SstepTokeyTypeInvalid, fmt.Sprintf("ToObjectKey: invalid key type '%s' for scenario object", s.ToObjectKey.KeyType), "ToObjectKey", s.ToObjectKey.KeyType, identity.KEY_TYPE_SCENARIO_OBJECT)
			}
		}
		if s.EventKey != nil {
			if err := s.EventKey.ValidateWithContext(ctx); err != nil {
				return coreerr.New(ctx, coreerr.SstepEventkeyInvalid, fmt.Sprintf("EventKey: %s", err.Error()), "EventKey")
			}
			if s.EventKey.KeyType != identity.KEY_TYPE_EVENT {
				return coreerr.NewWithValues(ctx, coreerr.SstepEventkeyTypeInvalid, fmt.Sprintf("EventKey: invalid key type '%s' for event", s.EventKey.KeyType), "EventKey", s.EventKey.KeyType, identity.KEY_TYPE_EVENT)
			}
		}
		if s.QueryKey != nil {
			if err := s.QueryKey.ValidateWithContext(ctx); err != nil {
				return coreerr.New(ctx, coreerr.SstepQuerykeyInvalid, fmt.Sprintf("QueryKey: %s", err.Error()), "QueryKey")
			}
			if s.QueryKey.KeyType != identity.KEY_TYPE_QUERY {
				return coreerr.NewWithValues(ctx, coreerr.SstepQuerykeyTypeInvalid, fmt.Sprintf("QueryKey: invalid key type '%s' for query", s.QueryKey.KeyType), "QueryKey", s.QueryKey.KeyType, identity.KEY_TYPE_QUERY)
			}
		}
		if s.ScenarioKey != nil {
			if err := s.ScenarioKey.ValidateWithContext(ctx); err != nil {
				return coreerr.New(ctx, coreerr.SstepScenariokeyInvalid, fmt.Sprintf("ScenarioKey: %s", err.Error()), "ScenarioKey")
			}
			if s.ScenarioKey.KeyType != identity.KEY_TYPE_SCENARIO {
				return coreerr.NewWithValues(ctx, coreerr.SstepScenariokeyTypeInvalid, fmt.Sprintf("ScenarioKey: invalid key type '%s' for scenario", s.ScenarioKey.KeyType), "ScenarioKey", s.ScenarioKey.KeyType, identity.KEY_TYPE_SCENARIO)
			}
		}
	default:
		return coreerr.NewWithValues(ctx, coreerr.SstepTypeUnknown, fmt.Sprintf("unknown step type '%s'", s.StepType), "StepType", s.StepType, "one of: leaf, sequence, switch, case, loop")
	}
	return nil
}

// ValidateWithParent validates the Step and its key's parent relationship.
func (s *Step) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	if err := s.Validate(ctx); err != nil {
		return err
	}
	if err := s.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// A scenario leaf cannot reference the scenario that contains it.
	if s.ScenarioKey != nil && parent != nil && *s.ScenarioKey == *parent {
		return coreerr.NewWithValues(ctx, coreerr.SstepScenarioSelfRef, "scenario leaf cannot reference its own scenario", "ScenarioKey", s.ScenarioKey.String(), "")
	}
	// Validate children with the same parent (all steps are flat under the scenario).
	for i := range s.Statements {
		childCtx := ctx.Child("statement", fmt.Sprintf("%d", i))
		if err := s.Statements[i].ValidateWithParent(childCtx, parent); err != nil {
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
	m := make(map[string]any)
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
func (s Step) MarshalYAML() (any, error) {
	m := make(map[string]any)
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
