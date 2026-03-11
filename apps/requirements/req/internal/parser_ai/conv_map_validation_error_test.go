package parser_ai

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

// MapValidationErrorSuite tests the mapValidationError function and its helpers.
type MapValidationErrorSuite struct {
	suite.Suite
}

func TestMapValidationErrorSuite(t *testing.T) {
	suite.Run(t, new(MapValidationErrorSuite))
}

// TestMappedCoreCode verifies that known core error codes produce the correct parser_ai error code.
func (s *MapValidationErrorSuite) TestMappedCoreCode() {
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
			ve := coreerr.New(tc.coreCode, "test message", "test_field")
			pe := mapValidationError(ve)

			s.Require().NotNil(pe)
			s.Equal(tc.parserCode, pe.Code, "wrong parser code for core code %s", tc.coreCode)
			s.Equal("test message", pe.Message, "message should be the ValidationError message")
			s.Equal("test_field", pe.Field, "field should be propagated")
		})
	}
}

// TestUnmappedCoreCodeFallsBackToGeneric verifies that an unmapped core code produces the catch-all error.
func (s *MapValidationErrorSuite) TestUnmappedCoreCodeFallsBackToGeneric() {
	ve := coreerr.New("TOTALLY_UNKNOWN_CODE", "something went wrong", "some_field")
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
}

// TestWrappedValidationErrorExtractsContext verifies that errors.Wrapf context is extracted.
func (s *MapValidationErrorSuite) TestWrappedValidationErrorExtractsContext() {
	ve := coreerr.New(coreerr.ClassActorNotfound, "actor not found", "actor_key")
	wrapped := errors.Wrapf(ve, "class 'order'")

	pe := mapValidationError(wrapped)

	s.Require().NotNil(pe)
	s.Equal(ErrConvReferenceNotFound, pe.Code, "should extract ValidationError through wrapping")
	s.Equal("actor not found", pe.Message, "message should be the inner ValidationError message")
	s.Equal("actor_key", pe.Field, "field should be propagated from inner ValidationError")
	s.Equal("class 'order'", pe.Context, "context should be extracted from wrapping chain")
}

// TestMultiLevelWrappingContext verifies deeply wrapped errors extract the full context chain.
func (s *MapValidationErrorSuite) TestMultiLevelWrappingContext() {
	ve := coreerr.New(coreerr.ParamDatatypesRequired, "DataTypeRules is required", "DataTypeRules")
	wrapped := errors.Wrapf(ve, "parameter 'amount'")
	wrapped = errors.Wrapf(wrapped, "action 'create'")
	wrapped = errors.Wrapf(wrapped, "class 'order'")

	pe := mapValidationError(wrapped)

	s.Require().NotNil(pe)
	s.Equal(ErrConvParamDatatypeRequired, pe.Code)
	s.Equal("DataTypeRules is required", pe.Message)
	s.Equal("DataTypeRules", pe.Field)
	s.Equal("class 'order': action 'create': parameter 'amount'", pe.Context)
}

// TestGotWantPassedThrough verifies that got/want fields are passed through as structured data.
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
		{"neither", "", "", "", ""},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ve := coreerr.NewWithValues(coreerr.ParamNameRequired, "param name required", "name", tc.got, tc.want)
			pe := mapValidationError(ve)

			s.Require().NotNil(pe)
			s.Equal(tc.expectedGot, pe.Got)
			s.Equal(tc.expectedWant, pe.Want)
		})
	}
}

// TestFieldNotSetWhenEmpty verifies that an empty field on the ValidationError produces an empty field on ParseError.
func (s *MapValidationErrorSuite) TestFieldNotSetWhenEmpty() {
	ve := coreerr.New(coreerr.DassocSameDomains, "domain association same domains", "")
	pe := mapValidationError(ve)

	s.Require().NotNil(pe)
	s.Equal(ErrConvDomainAssocSameDomains, pe.Code)
	s.Empty(pe.Field, "field should be empty when ValidationError has no field")
}

// TestFilePathIsModelJSON verifies all mapped errors use model.json as the file path.
func (s *MapValidationErrorSuite) TestFilePathIsModelJSON() {
	ve := coreerr.New(coreerr.ClassActorNotfound, "actor not found", "actor_key")
	pe := mapValidationError(ve)

	s.Equal("model.json", pe.File, "all validation errors should reference model.json")
}

// TestNoContextWhenUnwrapped verifies no context is set for unwrapped ValidationErrors.
func (s *MapValidationErrorSuite) TestNoContextWhenUnwrapped() {
	ve := coreerr.New(coreerr.ParamNameRequired, "name is required", "name")
	pe := mapValidationError(ve)

	s.Require().NotNil(pe)
	s.Empty(pe.Context, "unwrapped error should have no context")
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

// TestExtractWrappingContext tests the context extraction helper directly.
func (s *MapValidationErrorSuite) TestExtractWrappingContext() {
	ve := coreerr.New("TEST_CODE", "inner message", "field")

	// No wrapping — empty context.
	s.Empty(extractWrappingContext(ve, ve))

	// Single wrap.
	w1 := errors.Wrap(ve, "level 1")
	s.Equal("level 1", extractWrappingContext(w1, ve))

	// Double wrap.
	w2 := errors.Wrap(w1, "level 2")
	s.Equal("level 2: level 1", extractWrappingContext(w2, ve))
}
