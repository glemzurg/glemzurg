package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// multiplicityInOut is how two classes relate to each other.
type multiplicityInOut struct {
	LowerBound  uint `json:"lower_bound"`  // Zero is "any".
	HigherBound uint `json:"higher_bound"` // Zero is "any".
}

// ToRequirements converts the multiplicityInOut to requirements.Multiplicity.
func (m multiplicityInOut) ToRequirements() requirements.Multiplicity {
	return requirements.Multiplicity{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}

// FromRequirements creates a multiplicityInOut from requirements.Multiplicity.
func FromRequirementsMultiplicity(m requirements.Multiplicity) multiplicityInOut {
	return multiplicityInOut{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}
