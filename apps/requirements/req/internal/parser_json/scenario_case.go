package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// caseInOut represents a case in a switch node.
type caseInOut struct {
	Condition  string      `json:"condition" yaml:"condition"`
	Statements []nodeInOut `json:"statements" yaml:"statements"`
}

// ToRequirements converts the caseInOut to requirements.Case.
func (c caseInOut) ToRequirements() requirements.Case {
	nodeCase := requirements.Case{
		Condition:  c.Condition,
		Statements: nil,
	}
	for _, s := range c.Statements {
		nodeCase.Statements = append(nodeCase.Statements, s.ToRequirements())
	}

	return nodeCase
}

// FromRequirementsCase creates a caseInOut from requirements.Case.
func FromRequirementsCase(c requirements.Case) caseInOut {
	nodeCase := caseInOut{
		Condition:  c.Condition,
		Statements: nil,
	}

	for _, s := range c.Statements {
		nodeCase.Statements = append(nodeCase.Statements, FromRequirementsNode(s))
	}

	return nodeCase
}
