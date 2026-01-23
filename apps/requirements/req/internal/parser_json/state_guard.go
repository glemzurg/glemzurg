package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

// guardInOut is a constraint on an event in a state machine.
type guardInOut struct {
	Key     string `json:"key"`
	Name    string `json:"name"`    // A simple unique name for a guard, for internal use.
	Details string `json:"details"` // How the details of the guard are represented, what shows in the uml.
}

// ToRequirements converts the guardInOut to model_state.Guard.
func (g guardInOut) ToRequirements() (model_state.Guard, error) {
	key, err := identity.ParseKey(g.Key)
	if err != nil {
		return model_state.Guard{}, err
	}

	return model_state.Guard{
		Key:     key,
		Name:    g.Name,
		Details: g.Details,
	}, nil
}

// FromRequirements creates a guardInOut from model_state.Guard.
func FromRequirementsGuard(g model_state.Guard) guardInOut {
	return guardInOut{
		Key:     g.Key.String(),
		Name:    g.Name,
		Details: g.Details,
	}
}
