package parser_ai

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

// mapValidationError converts a core ValidationError (from model.Validate()) into a specific
// ParseError with an appropriate error code, file path, field, and hint.
//
// This replaces the old catch-all ErrConvModelValidation (21002) with targeted error codes
// that give calling AIs actionable guidance on how to fix the problem.
func mapValidationError(err error) *ParseError {
	var ve *coreerr.ValidationError
	if !stderrors.As(err, &ve) {
		// Not a ValidationError — fall back to generic.
		return convErr(ErrConvModelValidation, fmt.Sprintf("resulting model validation failed: %s", err.Error()), "model.json")
	}

	code := ve.Code()
	parserCode, found := coreToParserCode[code]
	if !found {
		// Unmapped core error — use catch-all but include the core code for debugging.
		return convErr(ErrConvModelValidation,
			fmt.Sprintf("resulting model validation failed: %s", err.Error()),
			"model.json",
		).WithHint(fmt.Sprintf("core error code: %s", code))
	}

	return buildMappedParseError(err, ve, parserCode)
}

// buildMappedParseError creates a ParseError from a mapped ValidationError,
// populating field, got/want, and wrapping context.
func buildMappedParseError(err error, ve *coreerr.ValidationError, parserCode int) *ParseError {
	pe := NewParseError(parserCode, ve.Message(), "")

	if ve.Field() != "" {
		pe = pe.WithField(ve.Field())
	}

	if ve.Got() != "" || ve.Want() != "" {
		pe = pe.WithGotWant(ve.Got(), ve.Want())
	}

	context := extractWrappingContext(err, ve)
	if context != "" {
		pe = pe.WithContext(context)
	}

	return pe
}

// extractWrappingContext extracts the pkg/errors.Wrapf context from the error chain.
// Given a full error like "domain 'x': subdomain 'y': class 'z': [CODE] msg",
// it strips the inner ValidationError text and returns "domain 'x': subdomain 'y': class 'z'".
func extractWrappingContext(err error, ve *coreerr.ValidationError) string {
	full := err.Error()
	inner := ve.Error()

	// The wrapping chain prepends "context: context: ... innerError".
	if full == inner {
		return "" // No wrapping context.
	}

	context := strings.TrimSuffix(full, ": "+inner)
	if context == full {
		// Fallback: if the pattern doesn't match, try trimming just the inner part.
		context = strings.TrimSuffix(full, inner)
		context = strings.TrimRight(context, ": ")
	}
	return context
}

// coreToParserCode maps core validation error codes to parser_ai error codes.
// Every core error that can reach model.Validate() should have a mapping here.
// Unmapped codes fall through to the generic ErrConvModelValidation.
var coreToParserCode = map[coreerr.Code]int{
	// Parameter errors.
	coreerr.ParamNameRequired:      ErrConvParamNameRequired,
	coreerr.ParamDatatypesRequired: ErrConvParamDatatypeRequired,

	// Logic type invalid for context.
	coreerr.ModelInvariantTypeInvalid:  ErrConvLogicTypeInvalid,
	coreerr.ClassInvariantTypeInvalid:  ErrConvLogicTypeInvalid,
	coreerr.AttrInvariantTypeInvalid:   ErrConvLogicTypeInvalid,
	coreerr.AttrDerivationTypeInvalid:  ErrConvLogicTypeInvalid,
	coreerr.ActionRequiresTypeInvalid:  ErrConvLogicTypeInvalid,
	coreerr.ActionGuaranteeTypeInvalid: ErrConvLogicTypeInvalid,
	coreerr.ActionSafetyTypeInvalid:    ErrConvLogicTypeInvalid,
	coreerr.QueryRequiresTypeInvalid:   ErrConvLogicTypeInvalid,
	coreerr.QueryGuaranteeTypeInvalid:  ErrConvLogicTypeInvalid,
	coreerr.GuardLogicTypeInvalid:      ErrConvLogicTypeInvalid,
	coreerr.GfuncLogicTypeInvalid:      ErrConvLogicTypeInvalid,

	// Duplicate let targets.
	coreerr.ModelInvariantDuplicateLet:  ErrConvLogicDuplicateLet,
	coreerr.ClassInvariantDuplicateLet:  ErrConvLogicDuplicateLet,
	coreerr.AttrInvariantDuplicateLet:   ErrConvLogicDuplicateLet,
	coreerr.ActionRequiresDuplicateLet:  ErrConvLogicDuplicateLet,
	coreerr.ActionGuaranteeDuplicateLet: ErrConvLogicDuplicateLet,
	coreerr.ActionSafetyDuplicateLet:    ErrConvLogicDuplicateLet,
	coreerr.QueryRequiresDuplicateLet:   ErrConvLogicDuplicateLet,
	coreerr.QueryGuaranteeDuplicateLet:  ErrConvLogicDuplicateLet,

	// Duplicate guarantee/action targets.
	coreerr.ActionGuaranteeDuplicateTarget: ErrConvLogicDuplicateTarget,
	coreerr.QueryGuaranteeDuplicateTarget:  ErrConvLogicDuplicateTarget,

	// Logic target required/forbidden/no-underscore.
	coreerr.LogicTargetRequired:     ErrConvLogicTargetRequired,
	coreerr.LogicTargetMustBeEmpty:  ErrConvLogicTargetNotAllowed,
	coreerr.LogicTargetNoUnderscore: ErrConvLogicTargetNoUnderscore,

	// Guarantee targets an attribute that doesn't exist.
	coreerr.ClassGuaranteeInvalidTarget: ErrConvGuaranteeInvalidTarget,

	// Cross-reference not found (re-validated in core after tree validation).
	coreerr.ClassActorNotfound:              ErrConvReferenceNotFound,
	coreerr.ClassSupergenNotfound:           ErrConvReferenceNotFound,
	coreerr.ClassSupergenWrongSubdomain:     ErrConvReferenceNotFound,
	coreerr.ClassSubgenNotfound:             ErrConvReferenceNotFound,
	coreerr.ClassSubgenWrongSubdomain:       ErrConvReferenceNotFound,
	coreerr.ActorSupergenNotfound:           ErrConvReferenceNotFound,
	coreerr.ActorSubgenNotfound:             ErrConvReferenceNotFound,
	coreerr.UcSupergenNotfound:              ErrConvReferenceNotFound,
	coreerr.UcSubgenNotfound:                ErrConvReferenceNotFound,
	coreerr.AssocFromNotfound:               ErrConvReferenceNotFound,
	coreerr.AssocToNotfound:                 ErrConvReferenceNotFound,
	coreerr.AssocAssocclassNotfound:         ErrConvReferenceNotFound,
	coreerr.DassocProblemNotfound:           ErrConvReferenceNotFound,
	coreerr.DassocSolutionNotfound:          ErrConvReferenceNotFound,
	coreerr.SobjectClassNotfound:            ErrConvReferenceNotFound,
	coreerr.SubdomainUshareSealevelNotfound: ErrConvReferenceNotFound,
	coreerr.SubdomainUshareMudlevelNotfound: ErrConvReferenceNotFound,
	coreerr.TransitionFromstateNotfound:     ErrConvReferenceNotFound,
	coreerr.TransitionTostateNotfound:       ErrConvReferenceNotFound,
	coreerr.TransitionEventNotfound:         ErrConvReferenceNotFound,
	coreerr.TransitionGuardNotfound:         ErrConvReferenceNotFound,
	coreerr.TransitionActionNotfound:        ErrConvReferenceNotFound,
	coreerr.StateactionActionNotfound:       ErrConvReferenceNotFound,
	coreerr.ModelCassocOrphanParent:         ErrConvReferenceNotFound,
	coreerr.DomainCassocOrphan:              ErrConvReferenceNotFound,
	coreerr.SubdomainCassocNoParent:         ErrConvReferenceNotFound,
	coreerr.SubdomainCassocWrongParent:      ErrConvReferenceNotFound,

	// Generalization cardinality errors.
	coreerr.ModelAgenSuperclassCount:      ErrConvGenCardinalityInvalid,
	coreerr.ModelAgenSubclassCount:        ErrConvGenCardinalityInvalid,
	coreerr.SubdomainCgenSuperclassCount:  ErrConvGenCardinalityInvalid,
	coreerr.SubdomainCgenSubclassCount:    ErrConvGenCardinalityInvalid,
	coreerr.SubdomainUcgenSuperclassCount: ErrConvGenCardinalityInvalid,
	coreerr.SubdomainUcgenSubclassCount:   ErrConvGenCardinalityInvalid,

	// Domain structural rules.
	coreerr.DomainSubdomainSingleKey:    ErrConvDomainStructureInvalid,
	coreerr.DomainSubdomainMultiDefault: ErrConvDomainStructureInvalid,

	// Domain association same domains.
	coreerr.DassocSameDomains: ErrConvDomainAssocSameDomains,

	// Association class same as endpoint.
	coreerr.AssocAssocclassSameFrom: ErrConvAssocClassSameAsEndpoint,
	coreerr.AssocAssocclassSameTo:   ErrConvAssocClassSameAsEndpoint,

	// Use case references non-actor class.
	coreerr.UcActorNotActorClass: ErrConvUseCaseActorNotActorClass,

	// Scenario step structural rules.
	coreerr.SstepSequenceMinStatements: ErrConvScenarioStepInvalid,
	coreerr.SstepSwitchMinCases:        ErrConvScenarioStepInvalid,
	coreerr.SstepSwitchCaseType:        ErrConvScenarioStepInvalid,
	coreerr.SstepCaseConditionRequired: ErrConvScenarioStepInvalid,
	coreerr.SstepLoopConditionRequired: ErrConvScenarioStepInvalid,
	coreerr.SstepLoopMinStatements:     ErrConvScenarioStepInvalid,
	coreerr.SstepScenarioSelfRef:       ErrConvScenarioStepInvalid,

	// Scenario step leaf requirements.
	coreerr.SstepEventFromRequired:      ErrConvScenarioStepInvalid,
	coreerr.SstepEventToRequired:        ErrConvScenarioStepInvalid,
	coreerr.SstepEventKeyRequired:       ErrConvScenarioStepInvalid,
	coreerr.SstepEventQueryForbidden:    ErrConvScenarioStepInvalid,
	coreerr.SstepQueryFromRequired:      ErrConvScenarioStepInvalid,
	coreerr.SstepQueryToRequired:        ErrConvScenarioStepInvalid,
	coreerr.SstepQueryKeyRequired:       ErrConvScenarioStepInvalid,
	coreerr.SstepQueryEventForbidden:    ErrConvScenarioStepInvalid,
	coreerr.SstepScenarioFromRequired:   ErrConvScenarioStepInvalid,
	coreerr.SstepScenarioToRequired:     ErrConvScenarioStepInvalid,
	coreerr.SstepScenarioKeyRequired:    ErrConvScenarioStepInvalid,
	coreerr.SstepScenarioEventForbidden: ErrConvScenarioStepInvalid,
	coreerr.SstepDeleteFromRequired:     ErrConvScenarioStepInvalid,
	coreerr.SstepDeleteToForbidden:      ErrConvScenarioStepInvalid,
	coreerr.SstepDeleteKeysForbidden:    ErrConvScenarioStepInvalid,

	// Logic spec/expression validation.
	coreerr.LogicSpecInvalid:           ErrConvLogicSpecInvalid,
	coreerr.LogicTargetTypespecInvalid: ErrConvLogicSpecInvalid,
	coreerr.GfuncLogicInvalid:          ErrConvLogicSpecInvalid,
	coreerr.GuardLogicInvalid:          ErrConvLogicSpecInvalid,
	coreerr.NsetSpecInvalid:            ErrConvLogicSpecInvalid,
	coreerr.NsetTypespecInvalid:        ErrConvLogicSpecInvalid,

	// Internal key errors — should not normally occur if converter works correctly.
	coreerr.KeyTypeInvalid:            ErrConvInternalKeyError,
	coreerr.KeySubkeyRequired:         ErrConvInternalKeyError,
	coreerr.KeyParentkeyMustBeBlank:   ErrConvInternalKeyError,
	coreerr.KeyParentkeyRequired:      ErrConvInternalKeyError,
	coreerr.KeyRootHasParent:          ErrConvInternalKeyError,
	coreerr.KeyRootHasParentkey:       ErrConvInternalKeyError,
	coreerr.KeyNoParent:               ErrConvInternalKeyError,
	coreerr.KeyWrongParentType:        ErrConvInternalKeyError,
	coreerr.KeyParentkeyMismatch:      ErrConvInternalKeyError,
	coreerr.KeyTypeUnknown:            ErrConvInternalKeyError,
	coreerr.KeyCassocParentUnknown:    ErrConvInternalKeyError,
	coreerr.KeyCassocModelHasParent:   ErrConvInternalKeyError,
	coreerr.ModelGfuncKeyMismatch:     ErrConvInternalKeyError,
	coreerr.ModelNsetKeyMismatch:      ErrConvInternalKeyError,
	coreerr.GfuncKeyInvalid:           ErrConvInternalKeyError,
	coreerr.GfuncKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.GfuncLogicKeyMismatch:     ErrConvInternalKeyError,
	coreerr.NsetKeyInvalid:            ErrConvInternalKeyError,
	coreerr.NsetKeyTypeInvalid:        ErrConvInternalKeyError,
	coreerr.GuardKeyInvalid:           ErrConvInternalKeyError,
	coreerr.GuardKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.GuardLogicKeyMismatch:     ErrConvInternalKeyError,
	coreerr.LogicKeyInvalid:           ErrConvInternalKeyError,
	coreerr.ActionKeyInvalid:          ErrConvInternalKeyError,
	coreerr.ActionKeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.QueryKeyInvalid:           ErrConvInternalKeyError,
	coreerr.QueryKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.EventKeyInvalid:           ErrConvInternalKeyError,
	coreerr.EventKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.StateKeyInvalid:           ErrConvInternalKeyError,
	coreerr.StateKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.StateactionKeyInvalid:     ErrConvInternalKeyError,
	coreerr.StateactionKeyTypeInvalid: ErrConvInternalKeyError,
	coreerr.TransitionKeyInvalid:      ErrConvInternalKeyError,
	coreerr.TransitionKeyTypeInvalid:  ErrConvInternalKeyError,
	coreerr.ClassKeyInvalid:           ErrConvInternalKeyError,
	coreerr.ClassKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.AssocKeyInvalid:           ErrConvInternalKeyError,
	coreerr.AssocKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.CgenKeyInvalid:            ErrConvInternalKeyError,
	coreerr.CgenKeyTypeInvalid:        ErrConvInternalKeyError,
	coreerr.ActorKeyInvalid:           ErrConvInternalKeyError,
	coreerr.ActorKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.AgenKeyInvalid:            ErrConvInternalKeyError,
	coreerr.AgenKeyTypeInvalid:        ErrConvInternalKeyError,
	coreerr.DomainKeyInvalid:          ErrConvInternalKeyError,
	coreerr.DomainKeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.SubdomainKeyInvalid:       ErrConvInternalKeyError,
	coreerr.SubdomainKeyTypeInvalid:   ErrConvInternalKeyError,
	coreerr.DassocKeyInvalid:          ErrConvInternalKeyError,
	coreerr.DassocKeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.ScenarioKeyInvalid:        ErrConvInternalKeyError,
	coreerr.ScenarioKeyTypeInvalid:    ErrConvInternalKeyError,
	coreerr.UcKeyInvalid:              ErrConvInternalKeyError,
	coreerr.UcKeyTypeInvalid:          ErrConvInternalKeyError,
	coreerr.UcgenKeyInvalid:           ErrConvInternalKeyError,
	coreerr.UcgenKeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.AttrKeyInvalid:            ErrConvInternalKeyError,
	coreerr.AttrKeyTypeInvalid:        ErrConvInternalKeyError,
	coreerr.SobjectKeyInvalid:         ErrConvInternalKeyError,
	coreerr.SobjectKeyTypeInvalid:     ErrConvInternalKeyError,
	coreerr.SstepKeyInvalid:           ErrConvInternalKeyError,
	coreerr.SstepKeyTypeInvalid:       ErrConvInternalKeyError,

	// Entity-level key type validations (caught in ValidateWithParent).
	coreerr.AssocFromkeyInvalid:           ErrConvInternalKeyError,
	coreerr.AssocFromkeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.AssocTokeyInvalid:             ErrConvInternalKeyError,
	coreerr.AssocTokeyTypeInvalid:         ErrConvInternalKeyError,
	coreerr.AssocAssocclassInvalid:        ErrConvInternalKeyError,
	coreerr.AssocAssocclassType:           ErrConvInternalKeyError,
	coreerr.DassocProblemkeyInvalid:       ErrConvInternalKeyError,
	coreerr.DassocProblemkeyType:          ErrConvInternalKeyError,
	coreerr.DassocSolutionkeyInvalid:      ErrConvInternalKeyError,
	coreerr.DassocSolutionkeyType:         ErrConvInternalKeyError,
	coreerr.ClassActorkeyInvalid:          ErrConvInternalKeyError,
	coreerr.ClassActorkeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.ClassSuperkeyInvalid:          ErrConvInternalKeyError,
	coreerr.ClassSuperkeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.ClassSubkeyInvalid:            ErrConvInternalKeyError,
	coreerr.ClassSubkeyTypeInvalid:        ErrConvInternalKeyError,
	coreerr.ClassSuperSubSame:             ErrConvInternalKeyError,
	coreerr.ActorSuperkeyInvalid:          ErrConvInternalKeyError,
	coreerr.ActorSuperkeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.ActorSubkeyInvalid:            ErrConvInternalKeyError,
	coreerr.ActorSubkeyTypeInvalid:        ErrConvInternalKeyError,
	coreerr.ActorSuperSubSame:             ErrConvInternalKeyError,
	coreerr.UcSuperkeyInvalid:             ErrConvInternalKeyError,
	coreerr.UcSuperkeyTypeInvalid:         ErrConvInternalKeyError,
	coreerr.UcSubkeyInvalid:               ErrConvInternalKeyError,
	coreerr.UcSubkeyTypeInvalid:           ErrConvInternalKeyError,
	coreerr.UcSuperSubSame:                ErrConvInternalKeyError,
	coreerr.SobjectClasskeyInvalid:        ErrConvInternalKeyError,
	coreerr.SobjectClasskeyTypeInvalid:    ErrConvInternalKeyError,
	coreerr.StateactionWhenRequired:       ErrConvInternalKeyError,
	coreerr.StateactionWhenInvalid:        ErrConvInternalKeyError,
	coreerr.StateactionActionkeyInvalid:   ErrConvInternalKeyError,
	coreerr.StateactionActionkeyType:      ErrConvInternalKeyError,
	coreerr.TransitionNoState:             ErrConvInternalKeyError,
	coreerr.TransitionFromstatekeyInvalid: ErrConvInternalKeyError,
	coreerr.TransitionFromstatekeyType:    ErrConvInternalKeyError,
	coreerr.TransitionTostatekeyInvalid:   ErrConvInternalKeyError,
	coreerr.TransitionTostatekeyType:      ErrConvInternalKeyError,
	coreerr.TransitionEventkeyInvalid:     ErrConvInternalKeyError,
	coreerr.TransitionEventkeyType:        ErrConvInternalKeyError,
	coreerr.TransitionGuardkeyInvalid:     ErrConvInternalKeyError,
	coreerr.TransitionGuardkeyType:        ErrConvInternalKeyError,
	coreerr.TransitionActionkeyInvalid:    ErrConvInternalKeyError,
	coreerr.TransitionActionkeyType:       ErrConvInternalKeyError,
	coreerr.SstepFromkeyInvalid:           ErrConvInternalKeyError,
	coreerr.SstepFromkeyTypeInvalid:       ErrConvInternalKeyError,
	coreerr.SstepTokeyInvalid:             ErrConvInternalKeyError,
	coreerr.SstepTokeyTypeInvalid:         ErrConvInternalKeyError,
	coreerr.SstepEventkeyInvalid:          ErrConvInternalKeyError,
	coreerr.SstepEventkeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.SstepQuerykeyInvalid:          ErrConvInternalKeyError,
	coreerr.SstepQuerykeyTypeInvalid:      ErrConvInternalKeyError,
	coreerr.SstepScenariokeyInvalid:       ErrConvInternalKeyError,
	coreerr.SstepScenariokeyTypeInvalid:   ErrConvInternalKeyError,
	coreerr.SstepLeafTypeRequired:         ErrConvInternalKeyError,
	coreerr.SstepLeafTypeUnknown:          ErrConvInternalKeyError,
	coreerr.SstepTypeUnknown:              ErrConvInternalKeyError,

	// Schema-caught errors that should never reach here, but map them for completeness.
	coreerr.ModelKeyRequired:          ErrConvInternalKeyError,
	coreerr.ModelNameRequired:         ErrConvInternalKeyError,
	coreerr.ClassNameRequired:         ErrConvInternalKeyError,
	coreerr.ActorNameRequired:         ErrConvInternalKeyError,
	coreerr.ActorTypeRequired:         ErrConvInternalKeyError,
	coreerr.ActorTypeInvalid:          ErrConvInternalKeyError,
	coreerr.DomainNameRequired:        ErrConvInternalKeyError,
	coreerr.SubdomainNameRequired:     ErrConvInternalKeyError,
	coreerr.AssocNameRequired:         ErrConvInternalKeyError,
	coreerr.AssocFromMultInvalid:      ErrConvInternalKeyError,
	coreerr.AssocToMultInvalid:        ErrConvInternalKeyError,
	coreerr.CgenNameRequired:          ErrConvInternalKeyError,
	coreerr.AgenNameRequired:          ErrConvInternalKeyError,
	coreerr.UcgenNameRequired:         ErrConvInternalKeyError,
	coreerr.EventNameRequired:         ErrConvInternalKeyError,
	coreerr.StateNameRequired:         ErrConvInternalKeyError,
	coreerr.GuardNameRequired:         ErrConvInternalKeyError,
	coreerr.ActionNameRequired:        ErrConvInternalKeyError,
	coreerr.QueryNameRequired:         ErrConvInternalKeyError,
	coreerr.AttrNameRequired:          ErrConvInternalKeyError,
	coreerr.ScenarioNameRequired:      ErrConvInternalKeyError,
	coreerr.UcNameRequired:            ErrConvInternalKeyError,
	coreerr.UcLevelRequired:           ErrConvInternalKeyError,
	coreerr.UcLevelInvalid:            ErrConvInternalKeyError,
	coreerr.GfuncNameRequired:         ErrConvInternalKeyError,
	coreerr.GfuncNameNoUnderscore:     ErrConvInternalKeyError,
	coreerr.NsetNameRequired:          ErrConvInternalKeyError,
	coreerr.SobjectNamestyleRequired:  ErrConvInternalKeyError,
	coreerr.SobjectNamestyleInvalid:   ErrConvInternalKeyError,
	coreerr.SobjectNameRequired:       ErrConvInternalKeyError,
	coreerr.SobjectNameMustBeBlank:    ErrConvInternalKeyError,
	coreerr.UshareSharetypeInvalid:    ErrConvInternalKeyError,
	coreerr.LogicTypeRequired:         ErrConvInternalKeyError,
	coreerr.LogicTypeInvalid:          ErrConvInternalKeyError,
	coreerr.LogicDescRequired:         ErrConvInternalKeyError,
	coreerr.ExprspecNotationRequired:  ErrConvInternalKeyError,
	coreerr.ExprspecNotationInvalid:   ErrConvInternalKeyError,
	coreerr.ExprspecExpressionInvalid: ErrConvInternalKeyError,
	coreerr.TypespecNotationRequired:  ErrConvInternalKeyError,
	coreerr.TypespecNotationInvalid:   ErrConvInternalKeyError,
	coreerr.TypespecExprtypeInvalid:   ErrConvInternalKeyError,

	// Expression AST validation errors — internal to TLA+ expression tree.
	coreerr.ExprIntValueRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprRatValueRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprSetElemRequired:       ErrConvLogicSpecInvalid,
	coreerr.ExprSetElemInvalid:        ErrConvLogicSpecInvalid,
	coreerr.ExprTupleElemRequired:     ErrConvLogicSpecInvalid,
	coreerr.ExprTupleElemNil:          ErrConvLogicSpecInvalid,
	coreerr.ExprTupleElemInvalid:      ErrConvLogicSpecInvalid,
	coreerr.ExprRecordFieldRequired:   ErrConvLogicSpecInvalid,
	coreerr.ExprRecordNameRequired:    ErrConvLogicSpecInvalid,
	coreerr.ExprRecordValueRequired:   ErrConvLogicSpecInvalid,
	coreerr.ExprRecordValueInvalid:    ErrConvLogicSpecInvalid,
	coreerr.ExprSetconstKindRequired:  ErrConvLogicSpecInvalid,
	coreerr.ExprSetconstKindInvalid:   ErrConvLogicSpecInvalid,
	coreerr.ExprAttrkeyInvalid:        ErrConvLogicSpecInvalid,
	coreerr.ExprLocalvarNameRequired:  ErrConvLogicSpecInvalid,
	coreerr.ExprPriorfieldRequired:    ErrConvLogicSpecInvalid,
	coreerr.ExprNextstateExprRequired: ErrConvLogicSpecInvalid,
	coreerr.ExprOpRequired:            ErrConvLogicSpecInvalid,
	coreerr.ExprOpInvalid:             ErrConvLogicSpecInvalid,
	coreerr.ExprLeftRequired:          ErrConvLogicSpecInvalid,
	coreerr.ExprLeftInvalid:           ErrConvLogicSpecInvalid,
	coreerr.ExprRightRequired:         ErrConvLogicSpecInvalid,
	coreerr.ExprRightInvalid:          ErrConvLogicSpecInvalid,
	coreerr.ExprElementRequired:       ErrConvLogicSpecInvalid,
	coreerr.ExprElementInvalid:        ErrConvLogicSpecInvalid,
	coreerr.ExprSetRequired:           ErrConvLogicSpecInvalid,
	coreerr.ExprSetInvalid:            ErrConvLogicSpecInvalid,
	coreerr.ExprExprRequired:          ErrConvLogicSpecInvalid,
	coreerr.ExprFieldRequired:         ErrConvLogicSpecInvalid,
	coreerr.ExprBaseRequired:          ErrConvLogicSpecInvalid,
	coreerr.ExprBaseInvalid:           ErrConvLogicSpecInvalid,
	coreerr.ExprTupleRequired:         ErrConvLogicSpecInvalid,
	coreerr.ExprTupleInvalid:          ErrConvLogicSpecInvalid,
	coreerr.ExprIndexRequired:         ErrConvLogicSpecInvalid,
	coreerr.ExprIndexInvalid:          ErrConvLogicSpecInvalid,
	coreerr.ExprAlterationsRequired:   ErrConvLogicSpecInvalid,
	coreerr.ExprAltFieldRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprAltValueRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprAltValueInvalid:       ErrConvLogicSpecInvalid,
	coreerr.ExprStrRequired:           ErrConvLogicSpecInvalid,
	coreerr.ExprStrInvalid:            ErrConvLogicSpecInvalid,
	coreerr.ExprOperandsMinTwo:        ErrConvLogicSpecInvalid,
	coreerr.ExprOperandRequired:       ErrConvLogicSpecInvalid,
	coreerr.ExprOperandInvalid:        ErrConvLogicSpecInvalid,
	coreerr.ExprConditionRequired:     ErrConvLogicSpecInvalid,
	coreerr.ExprConditionInvalid:      ErrConvLogicSpecInvalid,
	coreerr.ExprThenRequired:          ErrConvLogicSpecInvalid,
	coreerr.ExprThenInvalid:           ErrConvLogicSpecInvalid,
	coreerr.ExprElseRequired:          ErrConvLogicSpecInvalid,
	coreerr.ExprElseInvalid:           ErrConvLogicSpecInvalid,
	coreerr.ExprBranchesRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprBranchCondRequired:    ErrConvLogicSpecInvalid,
	coreerr.ExprBranchCondInvalid:     ErrConvLogicSpecInvalid,
	coreerr.ExprBranchResultRequired:  ErrConvLogicSpecInvalid,
	coreerr.ExprBranchResultInvalid:   ErrConvLogicSpecInvalid,
	coreerr.ExprOtherwiseInvalid:      ErrConvLogicSpecInvalid,
	coreerr.ExprQuantKindRequired:     ErrConvLogicSpecInvalid,
	coreerr.ExprQuantKindInvalid:      ErrConvLogicSpecInvalid,
	coreerr.ExprVariableRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprDomainRequired:        ErrConvLogicSpecInvalid,
	coreerr.ExprDomainInvalid:         ErrConvLogicSpecInvalid,
	coreerr.ExprPredicateRequired:     ErrConvLogicSpecInvalid,
	coreerr.ExprPredicateInvalid:      ErrConvLogicSpecInvalid,
	coreerr.ExprStartRequired:         ErrConvLogicSpecInvalid,
	coreerr.ExprStartInvalid:          ErrConvLogicSpecInvalid,
	coreerr.ExprEndRequired:           ErrConvLogicSpecInvalid,
	coreerr.ExprEndInvalid:            ErrConvLogicSpecInvalid,
	coreerr.ExprActionkeyInvalid:      ErrConvLogicSpecInvalid,
	coreerr.ExprArgRequired:           ErrConvLogicSpecInvalid,
	coreerr.ExprArgInvalid:            ErrConvLogicSpecInvalid,
	coreerr.ExprFunctionkeyInvalid:    ErrConvLogicSpecInvalid,
	coreerr.ExprModuleRequired:        ErrConvLogicSpecInvalid,
	coreerr.ExprFunctionRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprSetkeyInvalid:         ErrConvLogicSpecInvalid,

	// Expression type validation errors — internal to TLA+ type system.
	coreerr.ExprtypeEnumValuesRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprtypeSetElementRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprtypeSetElementInvalid:       ErrConvLogicSpecInvalid,
	coreerr.ExprtypeSequenceElementRequired: ErrConvLogicSpecInvalid,
	coreerr.ExprtypeSequenceElementInvalid:  ErrConvLogicSpecInvalid,
	coreerr.ExprtypeBagElementRequired:      ErrConvLogicSpecInvalid,
	coreerr.ExprtypeBagElementInvalid:       ErrConvLogicSpecInvalid,
	coreerr.ExprtypeTupleElementsRequired:   ErrConvLogicSpecInvalid,
	coreerr.ExprtypeTupleElementNil:         ErrConvLogicSpecInvalid,
	coreerr.ExprtypeTupleElementInvalid:     ErrConvLogicSpecInvalid,
	coreerr.ExprtypeRecordFieldsRequired:    ErrConvLogicSpecInvalid,
	coreerr.ExprtypeRecordFieldNameRequired: ErrConvLogicSpecInvalid,
	coreerr.ExprtypeRecordFieldTypeRequired: ErrConvLogicSpecInvalid,
	coreerr.ExprtypeRecordFieldTypeInvalid:  ErrConvLogicSpecInvalid,
	coreerr.ExprtypeFunctionReturnRequired:  ErrConvLogicSpecInvalid,
	coreerr.ExprtypeFunctionReturnInvalid:   ErrConvLogicSpecInvalid,
	coreerr.ExprtypeFunctionParamNil:        ErrConvLogicSpecInvalid,
	coreerr.ExprtypeFunctionParamInvalid:    ErrConvLogicSpecInvalid,
	coreerr.ExprtypeObjectClasskeyInvalid:   ErrConvLogicSpecInvalid,

	// DataType validation errors — internal to data type parsing/validation.
	coreerr.DtypeKeyRequired:                  ErrConvLogicSpecInvalid,
	coreerr.DtypeCollectiontypeRequired:       ErrConvLogicSpecInvalid,
	coreerr.DtypeCollectiontypeInvalid:        ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicRequired:               ErrConvLogicSpecInvalid,
	coreerr.DtypeRecordfieldsRequired:         ErrConvLogicSpecInvalid,
	coreerr.DtypeColluniqRequired:             ErrConvLogicSpecInvalid,
	coreerr.DtypeColluniqMustBeBlank:          ErrConvLogicSpecInvalid,
	coreerr.DtypeCollminMustBeBlank:           ErrConvLogicSpecInvalid,
	coreerr.DtypeCollmaxMustBeBlank:           ErrConvLogicSpecInvalid,
	coreerr.DtypeCollminTooSmall:              ErrConvLogicSpecInvalid,
	coreerr.DtypeCollmaxTooSmall:              ErrConvLogicSpecInvalid,
	coreerr.DtypeCollmaxLessThanMin:           ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicConstrainttypeRequired: ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicConstrainttypeInvalid:  ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicRefRequired:            ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicRefMustBeBlank:         ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicObjkeyRequired:         ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicObjkeyMustBeBlank:      ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicEnumsRequired:          ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicEnumsMustBeBlank:       ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicEnumordRequired:        ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicEnumordMustBeBlank:     ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicSpanRequired:           ErrConvLogicSpecInvalid,
	coreerr.DtypeAtomicSpanMustBeBlank:        ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanLowertypeRequired:        ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanLowertypeInvalid:         ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanHighertypeRequired:       ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanHighertypeInvalid:        ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanUnitsRequired:            ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanPrecisionRequired:        ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanLowervalRequired:         ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanHighervalRequired:        ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanLowerdenomRequired:       ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanLowerdenomInvalid:        ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanHigherdenomRequired:      ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanHigherdenomInvalid:       ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanPrecisionInvalid:         ErrConvLogicSpecInvalid,
	coreerr.DtypeSpanPrecisionNotPow10:        ErrConvLogicSpecInvalid,
	coreerr.DtypeEnumValueRequired:            ErrConvLogicSpecInvalid,
	coreerr.DtypeFieldNameRequired:            ErrConvLogicSpecInvalid,
	coreerr.DtypeFieldDatatypeRequired:        ErrConvLogicSpecInvalid,
	coreerr.DtypeFieldNameInvalid:             ErrConvLogicSpecInvalid,
}
