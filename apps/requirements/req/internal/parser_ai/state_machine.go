package parser_ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// inputStateAction represents an action reference in a state.
type inputStateAction struct {
	ActionKey string `json:"action_key"`
	When      string `json:"when"` // "entry", "exit", or "do"
}

// inputState represents a state within a state machine.
type inputState struct {
	Name       string             `json:"name"`
	Details    string             `json:"details,omitempty"`
	UMLComment string             `json:"uml_comment,omitempty"`
	Actions    []inputStateAction `json:"actions,omitempty"`
}

// inputEventParameter represents a parameter for an event.
type inputEventParameter struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// inputEvent represents an event in a state machine.
type inputEvent struct {
	Name       string                `json:"name"`
	Details    string                `json:"details,omitempty"`
	Parameters []inputEventParameter `json:"parameters,omitempty"`
}

// inputGuard represents a guard condition in a state machine.
type inputGuard struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

// inputTransition represents a transition in a state machine.
type inputTransition struct {
	FromStateKey *string `json:"from_state_key,omitempty"`
	ToStateKey   *string `json:"to_state_key,omitempty"`
	EventKey     string  `json:"event_key"`
	GuardKey     *string `json:"guard_key,omitempty"`
	ActionKey    *string `json:"action_key,omitempty"`
	UMLComment   string  `json:"uml_comment,omitempty"`
}

// inputStateMachine represents a state_machine.json file.
type inputStateMachine struct {
	States      map[string]*inputState `json:"states,omitempty"`
	Events      map[string]*inputEvent `json:"events,omitempty"`
	Guards      map[string]*inputGuard `json:"guards,omitempty"`
	Transitions []inputTransition      `json:"transitions,omitempty"`
}

// stateMachineSchema is the compiled JSON schema for state machine files.
var stateMachineSchema *jsonschema.Schema

// stateMachineSchemaContent is the raw JSON schema content for error reporting.
var stateMachineSchemaContent string

func init() {
	compiler := jsonschema.NewCompiler()
	schemaBytes, err := json_schemas.Schemas.ReadFile("state_machine.schema.json")
	if err != nil {
		panic("failed to read state_machine.schema.json: " + err.Error())
	}
	stateMachineSchemaContent = string(schemaBytes)
	if err := compiler.AddResource("state_machine.schema.json", strings.NewReader(stateMachineSchemaContent)); err != nil {
		panic("failed to add state_machine schema resource: " + err.Error())
	}
	stateMachineSchema, err = compiler.Compile("state_machine.schema.json")
	if err != nil {
		panic("failed to compile state_machine.schema.json: " + err.Error())
	}
}

// parseStateMachine parses a state machine JSON file content into an inputStateMachine struct.
// The filename parameter is the path to the JSON file being parsed.
// It validates the input against the state machine schema and returns detailed errors if validation fails.
func parseStateMachine(content []byte, filename string) (*inputStateMachine, error) {
	var sm inputStateMachine

	// Parse JSON
	if err := json.Unmarshal(content, &sm); err != nil {
		return nil, NewParseError(
			ErrStateMachineInvalidJSON,
			"failed to parse state machine JSON: "+err.Error(),
			filename,
		)
	}

	// Validate against JSON schema
	var jsonData any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, NewParseError(
			ErrStateMachineInvalidJSON,
			"failed to parse state machine JSON for schema validation: "+err.Error(),
			filename,
		)
	}
	if err := stateMachineSchema.Validate(jsonData); err != nil {
		return nil, NewParseError(
			ErrStateMachineSchemaViolation,
			"state machine JSON does not match schema: "+err.Error(),
			filename,
		).WithSchema(stateMachineSchemaContent)
	}

	// Validate required fields and business rules
	if err := validateStateMachine(&sm, filename); err != nil {
		return nil, err
	}

	return &sm, nil
}

// validateStateMachine validates an inputStateMachine struct.
// The filename parameter is the path to the JSON file being parsed.
func validateStateMachine(sm *inputStateMachine, filename string) error {
	// Validate states
	for stateKey, state := range sm.States {
		// State name is required (schema enforces this)
		if state.Name == "" {
			return NewParseError(
				ErrStateNameRequired,
				fmt.Sprintf("state '%s' name is required, got ''", stateKey),
				filename,
			).WithField("states." + stateKey + ".name")
		}

		// State name cannot be only whitespace
		if strings.TrimSpace(state.Name) == "" {
			return NewParseError(
				ErrStateNameEmpty,
				fmt.Sprintf("state '%s' name cannot be empty or whitespace only, got '%s'", stateKey, state.Name),
				filename,
			).WithField("states." + stateKey + ".name")
		}

		// Validate state actions
		for i, action := range state.Actions {
			// Action key is required (schema enforces this)
			if action.ActionKey == "" {
				return NewParseError(
					ErrStateActionKeyRequired,
					fmt.Sprintf("state '%s' action[%d] action_key is required", stateKey, i),
					filename,
				).WithField(fmt.Sprintf("states.%s.actions[%d].action_key", stateKey, i))
			}

			// Action key cannot be only whitespace
			if strings.TrimSpace(action.ActionKey) == "" {
				return NewParseError(
					ErrStateActionKeyRequired,
					fmt.Sprintf("state '%s' action[%d] action_key cannot be whitespace only, got '%s'", stateKey, i, action.ActionKey),
					filename,
				).WithField(fmt.Sprintf("states.%s.actions[%d].action_key", stateKey, i))
			}

			// When is required (schema enforces this)
			if action.When == "" {
				return NewParseError(
					ErrStateActionWhenRequired,
					fmt.Sprintf("state '%s' action[%d] when is required", stateKey, i),
					filename,
				).WithField(fmt.Sprintf("states.%s.actions[%d].when", stateKey, i))
			}

			// When must be a valid value (schema enforces this via enum, but we double-check)
			if action.When != "entry" && action.When != "exit" && action.When != "do" {
				return NewParseError(
					ErrStateActionWhenInvalid,
					fmt.Sprintf("state '%s' action[%d] when must be 'entry', 'exit', or 'do', got '%s'", stateKey, i, action.When),
					filename,
				).WithField(fmt.Sprintf("states.%s.actions[%d].when", stateKey, i))
			}
		}
	}

	// Validate events
	for eventKey, event := range sm.Events {
		// Event name is required (schema enforces this)
		if event.Name == "" {
			return NewParseError(
				ErrEventNameRequired,
				fmt.Sprintf("event '%s' name is required, got ''", eventKey),
				filename,
			).WithField("events." + eventKey + ".name")
		}

		// Event name cannot be only whitespace
		if strings.TrimSpace(event.Name) == "" {
			return NewParseError(
				ErrEventNameEmpty,
				fmt.Sprintf("event '%s' name cannot be empty or whitespace only, got '%s'", eventKey, event.Name),
				filename,
			).WithField("events." + eventKey + ".name")
		}

		// Validate event parameters
		for i, param := range event.Parameters {
			// Parameter name is required (schema enforces this)
			if param.Name == "" {
				return NewParseError(
					ErrEventParamNameRequired,
					fmt.Sprintf("event '%s' parameter[%d] name is required", eventKey, i),
					filename,
				).WithField(fmt.Sprintf("events.%s.parameters[%d].name", eventKey, i))
			}

			// Parameter name cannot be only whitespace
			if strings.TrimSpace(param.Name) == "" {
				return NewParseError(
					ErrEventParamNameRequired,
					fmt.Sprintf("event '%s' parameter[%d] name cannot be whitespace only, got '%s'", eventKey, i, param.Name),
					filename,
				).WithField(fmt.Sprintf("events.%s.parameters[%d].name", eventKey, i))
			}

			// Parameter source is required (schema enforces this)
			if param.Source == "" {
				return NewParseError(
					ErrEventParamSourceRequired,
					fmt.Sprintf("event '%s' parameter[%d] source is required", eventKey, i),
					filename,
				).WithField(fmt.Sprintf("events.%s.parameters[%d].source", eventKey, i))
			}

			// Parameter source cannot be only whitespace
			if strings.TrimSpace(param.Source) == "" {
				return NewParseError(
					ErrEventParamSourceRequired,
					fmt.Sprintf("event '%s' parameter[%d] source cannot be whitespace only, got '%s'", eventKey, i, param.Source),
					filename,
				).WithField(fmt.Sprintf("events.%s.parameters[%d].source", eventKey, i))
			}
		}
	}

	// Validate guards
	for guardKey, guard := range sm.Guards {
		// Guard name is required (schema enforces this)
		if guard.Name == "" {
			return NewParseError(
				ErrGuardNameRequired,
				fmt.Sprintf("guard '%s' name is required, got ''", guardKey),
				filename,
			).WithField("guards." + guardKey + ".name")
		}

		// Guard name cannot be only whitespace
		if strings.TrimSpace(guard.Name) == "" {
			return NewParseError(
				ErrGuardNameEmpty,
				fmt.Sprintf("guard '%s' name cannot be empty or whitespace only, got '%s'", guardKey, guard.Name),
				filename,
			).WithField("guards." + guardKey + ".name")
		}

		// Guard details is required (schema enforces this)
		if guard.Details == "" {
			return NewParseError(
				ErrGuardDetailsRequired,
				fmt.Sprintf("guard '%s' details is required, got ''", guardKey),
				filename,
			).WithField("guards." + guardKey + ".details")
		}

		// Guard details cannot be only whitespace
		if strings.TrimSpace(guard.Details) == "" {
			return NewParseError(
				ErrGuardDetailsRequired,
				fmt.Sprintf("guard '%s' details cannot be empty or whitespace only, got '%s'", guardKey, guard.Details),
				filename,
			).WithField("guards." + guardKey + ".details")
		}
	}

	// Validate transitions
	for i, transition := range sm.Transitions {
		// Event key is required (schema enforces this)
		if transition.EventKey == "" {
			return NewParseError(
				ErrTransitionEventRequired,
				fmt.Sprintf("transition[%d] event_key is required", i),
				filename,
			).WithField(fmt.Sprintf("transitions[%d].event_key", i))
		}

		// Event key cannot be only whitespace
		if strings.TrimSpace(transition.EventKey) == "" {
			return NewParseError(
				ErrTransitionEventRequired,
				fmt.Sprintf("transition[%d] event_key cannot be whitespace only, got '%s'", i, transition.EventKey),
				filename,
			).WithField(fmt.Sprintf("transitions[%d].event_key", i))
		}
	}

	return nil
}
