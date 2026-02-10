package req_model

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TlaDefinitionTestSuite struct {
	suite.Suite
}

func TestTlaDefinitionSuite(t *testing.T) {
	suite.Run(t, new(TlaDefinitionTestSuite))
}

func (s *TlaDefinitionTestSuite) TestValidate_Success_Function() {
	def := TlaDefinition{
		Name:       "_Max",
		Parameters: []string{"x", "y"},
		Tla:        "IF x > y THEN x ELSE y",
	}
	s.NoError(def.Validate())
}

func (s *TlaDefinitionTestSuite) TestValidate_Success_WithComment() {
	def := TlaDefinition{
		Name:       "_Max",
		Comment:    "Returns the maximum of two values.",
		Parameters: []string{"x", "y"},
		Tla:        "IF x > y THEN x ELSE y",
	}
	s.NoError(def.Validate())
}

func (s *TlaDefinitionTestSuite) TestValidate_Success_Set() {
	def := TlaDefinition{
		Name:       "_ValidStatuses",
		Parameters: []string{},
		Tla:        `{"pending", "active", "complete"}`,
	}
	s.NoError(def.Validate())
}

func (s *TlaDefinitionTestSuite) TestValidate_Success_Predicate() {
	def := TlaDefinition{
		Name:       "_HasChildren",
		Parameters: []string{"parent"},
		Tla:        "Cardinality(parent.children) > 0",
	}
	s.NoError(def.Validate())
}

func (s *TlaDefinitionTestSuite) TestValidate_Success_NoParameters() {
	def := TlaDefinition{
		Name:       "_Constant",
		Parameters: nil,
		Tla:        "42",
	}
	s.NoError(def.Validate())
}

func (s *TlaDefinitionTestSuite) TestValidate_Error_MissingUnderscore() {
	def := TlaDefinition{
		Name:       "Max", // Missing leading underscore
		Parameters: []string{"x", "y"},
		Tla:        "IF x > y THEN x ELSE y",
	}
	err := def.Validate()
	s.Error(err)
	s.Contains(err.Error(), "must start with underscore")
}

func (s *TlaDefinitionTestSuite) TestValidate_Error_EmptyName() {
	def := TlaDefinition{
		Name:       "",
		Parameters: []string{"x"},
		Tla:        "x + 1",
	}
	err := def.Validate()
	s.Error(err)
}

func (s *TlaDefinitionTestSuite) TestValidate_Error_EmptyTla() {
	def := TlaDefinition{
		Name:       "_Something",
		Parameters: []string{},
		Tla:        "",
	}
	err := def.Validate()
	s.Error(err)
}
