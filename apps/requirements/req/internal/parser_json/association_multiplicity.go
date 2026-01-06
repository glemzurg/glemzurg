package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
)

// multiplicityInOut is how two classes relate to each other.
type multiplicityInOut struct {
	LowerBound  uint `json:"lower_bound"`  // Zero is "any".
	HigherBound uint `json:"higher_bound"` // Zero is "any".
}

// ToRequirements converts the multiplicityInOut to class.Multiplicity.
func (m multiplicityInOut) ToRequirements() class.Multiplicity {
	return class.Multiplicity{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}

// FromRequirements creates a multiplicityInOut from class.Multiplicity.
func FromRequirementsMultiplicity(m class.Multiplicity) multiplicityInOut {
	return multiplicityInOut{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}
