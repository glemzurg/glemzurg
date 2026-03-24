package parser_ai

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/stretchr/testify/suite"
)

// MapValidationErrorSuite tests the mapValidationError function and its helpers.
type MapValidationErrorSuite struct {
	suite.Suite
}

func TestMapValidationErrorSuite(t *testing.T) {
	suite.Run(t, new(MapValidationErrorSuite))
}

// TestMappedCoreCodeWithPath verifies that known core error codes with a non-empty path
// produce the correct parser_ai error code and a fully populated CoreValidationDetail.
func (s *MapValidationErrorSuite) TestMappedCoreCodeWithPath() {
	tests := []struct {
		name       string
		coreCode   coreerr.Code
		parserCode int
	}{
		// Parameter errors.
		{"param name required", coreerr.ParamNameRequired, ErrConvParamNameRequired},
		{"param datatypes required", coreerr.ParamDatatypesRequired, ErrConvParamDatatypeRequired},

		// Logic type invalid.
		{"model invariant type invalid", coreerr.ModelInvariantTypeInvalid, ErrConvLogicTypeInvalid},
		{"class invariant type invalid", coreerr.ClassInvariantTypeInvalid, ErrConvLogicTypeInvalid},
		{"action requires type invalid", coreerr.ActionRequiresTypeInvalid, ErrConvLogicTypeInvalid},
		{"action guarantee type invalid", coreerr.ActionGuaranteeTypeInvalid, ErrConvLogicTypeInvalid},
		{"guard logic type invalid", coreerr.GuardLogicTypeInvalid, ErrConvLogicTypeInvalid},
		{"gfunc logic type invalid", coreerr.GfuncLogicTypeInvalid, ErrConvLogicTypeInvalid},

		// Duplicate let targets.
		{"model invariant duplicate let", coreerr.ModelInvariantDuplicateLet, ErrConvLogicDuplicateLet},
		{"action requires duplicate let", coreerr.ActionRequiresDuplicateLet, ErrConvLogicDuplicateLet},
		{"action guarantee duplicate let", coreerr.ActionGuaranteeDuplicateLet, ErrConvLogicDuplicateLet},

		// Duplicate guarantee targets.
		{"action guarantee duplicate target", coreerr.ActionGuaranteeDuplicateTarget, ErrConvLogicDuplicateTarget},
		{"query guarantee duplicate target", coreerr.QueryGuaranteeDuplicateTarget, ErrConvLogicDuplicateTarget},

		// Logic target rules.
		{"logic target required", coreerr.LogicTargetRequired, ErrConvLogicTargetRequired},
		{"logic target must be empty", coreerr.LogicTargetMustBeEmpty, ErrConvLogicTargetNotAllowed},
		{"logic target no underscore", coreerr.LogicTargetNoUnderscore, ErrConvLogicTargetNoUnderscore},

		// Guarantee invalid target.
		{"class guarantee invalid target", coreerr.ClassGuaranteeInvalidTarget, ErrConvGuaranteeInvalidTarget},

		// Cross-reference not found.
		{"class actor not found", coreerr.ClassActorNotfound, ErrConvReferenceNotFound},
		{"assoc from not found", coreerr.AssocFromNotfound, ErrConvReferenceNotFound},
		{"assoc to not found", coreerr.AssocToNotfound, ErrConvReferenceNotFound},
		{"transition event not found", coreerr.TransitionEventNotfound, ErrConvReferenceNotFound},
		{"transition guard not found", coreerr.TransitionGuardNotfound, ErrConvReferenceNotFound},

		// Generalization cardinality.
		{"model agen superclass count", coreerr.ModelAgenSuperclassCount, ErrConvGenCardinalityInvalid},
		{"subdomain cgen subclass count", coreerr.SubdomainCgenSubclassCount, ErrConvGenCardinalityInvalid},

		// Domain structural rules.
		{"domain subdomain single key", coreerr.DomainSubdomainSingleKey, ErrConvDomainStructureInvalid},
		{"domain subdomain multi default", coreerr.DomainSubdomainMultiDefault, ErrConvDomainStructureInvalid},

		// Domain association same domains.
		{"dassoc same domains", coreerr.DassocSameDomains, ErrConvDomainAssocSameDomains},

		// Association class same as endpoint.
		{"assoc assocclass same from", coreerr.AssocAssocclassSameFrom, ErrConvAssocClassSameAsEndpoint},
		{"assoc assocclass same to", coreerr.AssocAssocclassSameTo, ErrConvAssocClassSameAsEndpoint},

		// Use case references non-actor class.
		{"uc actor not actor class", coreerr.UcActorNotActorClass, ErrConvUseCaseActorNotActorClass},

		// Scenario step errors.
		{"sstep sequence min statements", coreerr.SstepSequenceMinStatements, ErrConvScenarioStepInvalid},
		{"sstep switch min cases", coreerr.SstepSwitchMinCases, ErrConvScenarioStepInvalid},
		{"sstep event from required", coreerr.SstepEventFromRequired, ErrConvScenarioStepInvalid},

		// Logic spec/expression errors.
		{"logic spec invalid", coreerr.LogicSpecInvalid, ErrConvLogicSpecInvalid},
		{"gfunc logic invalid", coreerr.GfuncLogicInvalid, ErrConvLogicSpecInvalid},
		{"guard logic invalid", coreerr.GuardLogicInvalid, ErrConvLogicSpecInvalid},
		{"nset spec invalid", coreerr.NsetSpecInvalid, ErrConvLogicSpecInvalid},

		// Expression AST errors.
		{"expr op invalid", coreerr.ExprOpInvalid, ErrConvLogicSpecInvalid},
		{"expr left required", coreerr.ExprLeftRequired, ErrConvLogicSpecInvalid},

		// Expression type errors.
		{"exprtype enum values required", coreerr.ExprtypeEnumValuesRequired, ErrConvLogicSpecInvalid},

		// DataType errors.
		{"dtype key required", coreerr.DtypeKeyRequired, ErrConvLogicSpecInvalid},
		{"dtype collection type invalid", coreerr.DtypeCollectiontypeInvalid, ErrConvLogicSpecInvalid},

		// Internal key errors.
		{"key type invalid", coreerr.KeyTypeInvalid, ErrConvInternalKeyError},
		{"class key invalid", coreerr.ClassKeyInvalid, ErrConvInternalKeyError},
		{"model key required", coreerr.ModelKeyRequired, ErrConvInternalKeyError},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx := coreerr.NewContext("model", "test").Child("class", "order")
			ve := coreerr.New(ctx, tc.coreCode, "test message", "test_field")
			pe := mapValidationError(ve)

			s.Require().NotNil(pe)
			s.Equal(tc.parserCode, pe.Code, "wrong parser code for core code %s", tc.coreCode)
			s.Equal("test message", pe.Message, "message should be the ValidationError message")

			// Core validation detail should be fully populated.
			s.Require().NotNil(pe.Context, "Context should be populated for mapped core errors")
			s.Equal(string(tc.coreCode), pe.Context.Code, "context code should be the core error code")
			s.Equal("test message", pe.Context.Message, "context message should match")
			s.Equal("test_field", pe.Context.Field, "context field should be propagated")
			s.Equal("model[test].class[order]", coreerr.FormatPath(pe.Context.Path), "context path should match")
		})
	}
}

// TestMappedCoreCodeWithRootPath verifies a mapped core code with a root-level path works.
func (s *MapValidationErrorSuite) TestMappedCoreCodeWithRootPath() {
	ctx := coreerr.NewContext("test", "")
	ve := coreerr.New(ctx, coreerr.ParamDatatypesRequired, "DataTypeRules is required", "DataTypeRules")
	pe := mapValidationError(ve)

	s.Require().NotNil(pe)
	s.Equal(ErrConvParamDatatypeRequired, pe.Code, "should map to specific code when path exists")
	s.Equal("DataTypeRules is required", pe.Message)
	s.Require().NotNil(pe.Context)
	s.Equal("DataTypeRules", pe.Context.Field)
	s.Equal("test", coreerr.FormatPath(pe.Context.Path))
}

// TestUnmappedCoreCodeFallsBackToGeneric verifies that an unmapped core code produces the catch-all error.
func (s *MapValidationErrorSuite) TestUnmappedCoreCodeFallsBackToGeneric() {
	ctx := coreerr.NewContext("test", "")
	ve := coreerr.New(ctx, "TOTALLY_UNKNOWN_CODE", "something went wrong", "some_field")
	pe := mapValidationError(ve)

	s.Require().NotNil(pe)
	s.Equal(ErrConvModelValidation, pe.Code, "unmapped code should fall back to ErrConvModelValidation")
	s.Contains(pe.Error(), "resulting model validation failed")
	s.Contains(pe.Hint, "core error code: TOTALLY_UNKNOWN_CODE")
}

// TestNonValidationErrorFallsBackToGeneric verifies that a plain error (not *ValidationError) uses the catch-all.
func (s *MapValidationErrorSuite) TestNonValidationErrorFallsBackToGeneric() {
	plainErr := fmt.Errorf("something broke")
	pe := mapValidationError(plainErr)

	s.Require().NotNil(pe)
	s.Equal(ErrConvModelValidation, pe.Code, "non-ValidationError should fall back to ErrConvModelValidation")
	s.Contains(pe.Error(), "resulting model validation failed")
	s.Contains(pe.Error(), "something broke")
	s.Nil(pe.Context, "non-validation errors should have nil Context")
}

// TestCoreValidationDetailPath verifies that the path from ValidationContext populates Context.Path.
func (s *MapValidationErrorSuite) TestCoreValidationDetailPath() {
	ctx := coreerr.NewContext("model", "test").Child("class", "order")
	ve := coreerr.New(ctx, coreerr.ClassActorNotfound, "actor not found", "actor_key")

	pe := mapValidationError(ve)

	s.Require().NotNil(pe)
	s.Equal(ErrConvReferenceNotFound, pe.Code)
	s.Equal("actor not found", pe.Message)
	s.Require().NotNil(pe.Context)
	s.Equal("actor_key", pe.Context.Field)
	s.Equal("model[test].class[order]", coreerr.FormatPath(pe.Context.Path))
}

// TestDeepPathContext verifies deeply nested paths produce correct structured context.
func (s *MapValidationErrorSuite) TestDeepPathContext() {
	ctx := coreerr.NewContext("model", "test").
		Child("domain", "sales").
		Child("subdomain", "default").
		Child("class", "order").
		Child("action", "create").
		Child("parameter", "0")
	ve := coreerr.New(ctx, coreerr.ParamDatatypesRequired, "DataTypeRules is required", "DataTypeRules")

	pe := mapValidationError(ve)

	s.Require().NotNil(pe)
	s.Equal(ErrConvParamDatatypeRequired, pe.Code)
	s.Equal("DataTypeRules is required", pe.Message)
	s.Require().NotNil(pe.Context)
	s.Equal("DataTypeRules", pe.Context.Field)
	s.Equal("model[test].domain[sales].subdomain[default].class[order].action[create].parameter[0]",
		coreerr.FormatPath(pe.Context.Path))
	s.Len(pe.Context.Path, 6, "path should have 6 segments")
}

// TestGotWantPassedThrough verifies that got/want fields are passed through in CoreValidationDetail.
func (s *MapValidationErrorSuite) TestGotWantPassedThrough() {
	tests := []struct {
		name         string
		got          string
		want         string
		expectedGot  string
		expectedWant string
	}{
		{"both got and want", "bad_value", "good_value", "bad_value", "good_value"},
		{"only got", "bad_value", "", "bad_value", ""},
		{"only want", "", "good_value", "", "good_value"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx := coreerr.NewContext("model", "test").Child("class", "order")
			ve := coreerr.NewWithValues(ctx, coreerr.ParamNameRequired, "param name required", "name", tc.got, tc.want)
			pe := mapValidationError(ve)

			s.Require().NotNil(pe)
			s.Equal(ErrConvParamNameRequired, pe.Code)
			s.Require().NotNil(pe.Context)
			s.Equal(tc.expectedGot, pe.Context.Got)
			s.Equal(tc.expectedWant, pe.Context.Want)
		})
	}
}

// TestFilePathIsBlankForMappedErrors verifies mapped core errors have no file path.
func (s *MapValidationErrorSuite) TestFilePathIsBlankForMappedErrors() {
	ctx := coreerr.NewContext("model", "test").Child("class", "order")
	ve := coreerr.New(ctx, coreerr.ClassActorNotfound, "actor not found", "actor_key")
	pe := mapValidationError(ve)

	s.Empty(pe.File, "mapped core errors should have blank file — context provides location")
}

// TestAllCoreCodesInMapHaveMatchingParserCode verifies every entry in coreToParserCode maps to a valid parser_ai code.
func (s *MapValidationErrorSuite) TestAllCoreCodesInMapHaveMatchingParserCode() {
	validParserCodes := map[int]bool{
		ErrConvParamDatatypeRequired:     true,
		ErrConvParamNameRequired:         true,
		ErrConvLogicTypeInvalid:          true,
		ErrConvLogicDuplicateLet:         true,
		ErrConvLogicDuplicateTarget:      true,
		ErrConvLogicTargetRequired:       true,
		ErrConvLogicTargetNotAllowed:     true,
		ErrConvLogicTargetNoUnderscore:   true,
		ErrConvReferenceNotFound:         true,
		ErrConvGenCardinalityInvalid:     true,
		ErrConvDomainStructureInvalid:    true,
		ErrConvScenarioStepInvalid:       true,
		ErrConvGuaranteeInvalidTarget:    true,
		ErrConvAssocClassSameAsEndpoint:  true,
		ErrConvInternalKeyError:          true,
		ErrConvUseCaseActorNotActorClass: true,
		ErrConvLogicSpecInvalid:          true,
		ErrConvDomainAssocSameDomains:    true,
	}

	for coreCode, parserCode := range coreToParserCode {
		s.True(validParserCodes[parserCode],
			"core code %s maps to unknown parser code %d", coreCode, parserCode)
	}
}
