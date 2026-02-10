package req_model

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// TlaDefinition represents a global TLA+ definition that can be referenced
// from other TLA+ expressions throughout the model.
//
// Definitions can be:
//   - A set for membership checks: x \in _SetOfValues
//   - A function for data transformation: _Max(x, y)
//   - A common boolean predicate: _HasAChild(x)
//
// All global definitions must have a leading underscore to distinguish them
// from class-scoped actions.
type TlaDefinition struct {
	Name       string   // The definition name (e.g., _Max, _SetOfValues). Must start with underscore.
	Comment    string   // Optional human-readable description of this definition.
	Parameters []string // The parameter names in TLA+ (e.g., ["x", "y"] for _Max(x, y)).
	Tla        string   // The TLA+ expression body.
}

// Validate validates the TlaDefinition struct.
func (d *TlaDefinition) Validate() error {
	return validation.ValidateStruct(d,
		validation.Field(&d.Name, validation.Required, validation.By(func(value interface{}) error {
			name := value.(string)
			if len(name) == 0 {
				return errors.New("name is required")
			}
			if name[0] != '_' {
				return errors.Errorf("global TLA definition name '%s' must start with underscore", name)
			}
			return nil
		})),
		validation.Field(&d.Tla, validation.Required),
	)
}
