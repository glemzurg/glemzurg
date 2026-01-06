package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/scenario"

// caseInOut represents a case in a switch node.
type caseInOut struct {
	Condition  string      `json:"condition" yaml:"condition"`
	Statements []nodeInOut `json:"statements" yaml:"statements"`
}

// ToRequirements converts the caseInOut to scenario.Case.
func (c caseInOut) ToRequirements() scenario.Case {
	nodeCase := scenario.Case{
		Condition:  c.Condition,
		Statements: nil,
	}
	for _, s := range c.Statements {
		nodeCase.Statements = append(nodeCase.Statements, s.ToRequirements())
	}

	return nodeCase
}

// FromRequirementsCase creates a caseInOut from scenario.Case.
func FromRequirementsCase(c scenario.Case) caseInOut {
	nodeCase := caseInOut{
		Condition:  c.Condition,
		Statements: nil,
	}

	for _, s := range c.Statements {
		nodeCase.Statements = append(nodeCase.Statements, FromRequirementsNode(s))
	}

	return nodeCase
}
