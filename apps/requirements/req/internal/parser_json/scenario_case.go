package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// caseInOut represents a case in a switch node.
type caseInOut struct {
	Condition  string      `json:"condition" yaml:"condition"`
	Statements []nodeInOut `json:"statements" yaml:"statements"`
}

// ToRequirements converts the caseInOut to requirements.Case.
func (c caseInOut) ToRequirements() requirements.Case {
	statements := make([]requirements.Node, len(c.Statements))
	for i, s := range c.Statements {
		statements[i] = s.ToRequirements()
	}
	return requirements.Case{
		Condition:  c.Condition,
		Statements: statements,
	}
}

// FromRequirementsCase creates a caseInOut from requirements.Case.
func FromRequirementsCase(c requirements.Case) caseInOut {
	statements := make([]nodeInOut, len(c.Statements))
	for i, s := range c.Statements {
		statements[i] = FromRequirementsNode(s)
	}
	return caseInOut{
		Condition:  c.Condition,
		Statements: statements,
	}
}
