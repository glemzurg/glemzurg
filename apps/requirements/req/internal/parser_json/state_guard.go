package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// guardInOut is a constraint on an event in a state machine.
type guardInOut struct {
	Key     string `json:"key"`
	Name    string `json:"name"`    // A simple unique name for a guard, for internal use.
	Details string `json:"details"` // How the details of the guard are represented, what shows in the uml.
}

// ToRequirements converts the guardInOut to requirements.Guard.
func (g guardInOut) ToRequirements() requirements.Guard {
	return requirements.Guard{
		Key:     g.Key,
		Name:    g.Name,
		Details: g.Details,
	}
}

// FromRequirements creates a guardInOut from requirements.Guard.
func FromRequirementsGuard(g requirements.Guard) guardInOut {
	return guardInOut{
		Key:     g.Key,
		Name:    g.Name,
		Details: g.Details,
	}
}
