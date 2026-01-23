package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

// multiplicityInOut is how two classes relate to each other.
type multiplicityInOut struct {
	LowerBound  uint `json:"lower_bound"`  // Zero is "any".
	HigherBound uint `json:"higher_bound"` // Zero is "any".
}

// ToRequirements converts the multiplicityInOut to model_class.Multiplicity.
func (m multiplicityInOut) ToRequirements() model_class.Multiplicity {
	return model_class.Multiplicity{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}

// FromRequirements creates a multiplicityInOut from model_class.Multiplicity.
func FromRequirementsMultiplicity(m model_class.Multiplicity) multiplicityInOut {
	return multiplicityInOut{
		LowerBound:  m.LowerBound,
		HigherBound: m.HigherBound,
	}
}
