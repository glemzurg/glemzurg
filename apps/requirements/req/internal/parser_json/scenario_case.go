package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"

// caseInOut represents a case in a switch node.
type caseInOut struct {
	Condition  string      `json:"condition" yaml:"condition"`
	Statements []nodeInOut `json:"statements" yaml:"statements"`
}

// ToRequirements converts the caseInOut to model_scenario.Case.
func (c caseInOut) ToRequirements() (model_scenario.Case, error) {
	nodeCase := model_scenario.Case{
		Condition:  c.Condition,
		Statements: nil,
	}
	for _, s := range c.Statements {
		stmt, err := s.ToRequirements()
		if err != nil {
			return model_scenario.Case{}, err
		}
		nodeCase.Statements = append(nodeCase.Statements, stmt)
	}

	return nodeCase, nil
}

// FromRequirementsCase creates a caseInOut from model_scenario.Case.
func FromRequirementsCase(c model_scenario.Case) caseInOut {
	nodeCase := caseInOut{
		Condition:  c.Condition,
		Statements: nil,
	}

	for _, s := range c.Statements {
		nodeCase.Statements = append(nodeCase.Statements, FromRequirementsNode(s))
	}

	return nodeCase
}
