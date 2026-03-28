package coreerr

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CodesSuite struct {
	suite.Suite
}

func TestCodesSuite(t *testing.T) {
	suite.Run(t, new(CodesSuite))
}

// TestAllCodesUnique verifies that no two error code constants share the same string value.
func (suite *CodesSuite) TestAllCodesUnique() {
	codes := allCodeConstants()

	seen := make(map[Code]string) // code value -> constant name
	for name, code := range codes {
		if existing, found := seen[code]; found {
			suite.Failf("duplicate error code", "constants %s and %s both have value %q", existing, name, code)
		}
		seen[code] = name
	}
}

// TestAllCodesNonEmpty verifies that no error code constant is an empty string.
func (suite *CodesSuite) TestAllCodesNonEmpty() {
	codes := allCodeConstants()

	for name, code := range codes {
		suite.NotEmpty(string(code), "constant %s has empty code value", name)
	}
}

// allCodeConstants returns all exported Code constants in this package via reflection.
// This uses the fact that all codes are package-level variables of type Code.
// We maintain a manual list to ensure completeness.
func allCodeConstants() map[string]Code {
	return map[string]Code{
		// Key errors.
		"KeyTypeInvalid":          KeyTypeInvalid,
		"KeySubkeyRequired":       KeySubkeyRequired,
		"KeyParentkeyMustBeBlank": KeyParentkeyMustBeBlank,
		"KeyParentkeyRequired":    KeyParentkeyRequired,
		"KeyRootHasParent":        KeyRootHasParent,
		"KeyRootHasParentkey":     KeyRootHasParentkey,
		"KeyNoParent":             KeyNoParent,
		"KeyWrongParentType":      KeyWrongParentType,
		"KeyParentkeyMismatch":    KeyParentkeyMismatch,
		"KeyTypeUnknown":          KeyTypeUnknown,
		"KeyCassocParentUnknown":  KeyCassocParentUnknown,
		"KeyCassocModelHasParent": KeyCassocModelHasParent,

		// Model errors.
		"ModelKeyRequired":           ModelKeyRequired,
		"ModelNameRequired":          ModelNameRequired,
		"ModelInvariantTypeInvalid":  ModelInvariantTypeInvalid,
		"ModelInvariantDuplicateLet": ModelInvariantDuplicateLet,
		"ModelGfuncKeyMismatch":      ModelGfuncKeyMismatch,
		"ModelNsetKeyMismatch":       ModelNsetKeyMismatch,
		"ModelAgenSuperclassCount":   ModelAgenSuperclassCount,
		"ModelAgenSubclassCount":     ModelAgenSubclassCount,
		"ModelCassocOrphanParent":    ModelCassocOrphanParent,

		// ExpressionSpec errors.
		"ExprspecNotationRequired":  ExprspecNotationRequired,
		"ExprspecNotationInvalid":   ExprspecNotationInvalid,
		"ExprspecExpressionInvalid": ExprspecExpressionInvalid,

		// TypeSpec errors.
		"TypespecNotationRequired": TypespecNotationRequired,
		"TypespecNotationInvalid":  TypespecNotationInvalid,
		"TypespecExprtypeInvalid":  TypespecExprtypeInvalid,

		// Logic errors.
		"LogicKeyInvalid":            LogicKeyInvalid,
		"LogicTypeRequired":          LogicTypeRequired,
		"LogicTypeInvalid":           LogicTypeInvalid,
		"LogicDescRequired":          LogicDescRequired,
		"LogicTargetRequired":        LogicTargetRequired,
		"LogicTargetMustBeEmpty":     LogicTargetMustBeEmpty,
		"LogicTargetNoUnderscore":    LogicTargetNoUnderscore,
		"LogicSpecInvalid":           LogicSpecInvalid,
		"LogicTargetTypespecInvalid": LogicTargetTypespecInvalid,

		// GlobalFunction errors.
		"GfuncKeyInvalid":       GfuncKeyInvalid,
		"GfuncKeyTypeInvalid":   GfuncKeyTypeInvalid,
		"GfuncNameNoUnderscore": GfuncNameNoUnderscore,
		"GfuncLogicInvalid":     GfuncLogicInvalid,
		"GfuncLogicKeyMismatch": GfuncLogicKeyMismatch,
		"GfuncLogicTypeInvalid": GfuncLogicTypeInvalid,

		// NamedSet errors.
		"NsetKeyInvalid":      NsetKeyInvalid,
		"NsetKeyTypeInvalid":  NsetKeyTypeInvalid,
		"NsetNameRequired":    NsetNameRequired,
		"NsetSpecInvalid":     NsetSpecInvalid,
		"NsetTypespecInvalid": NsetTypespecInvalid,

		// Parameter errors.
		"ParamNameRequired":      ParamNameRequired,
		"ParamDatatypesRequired": ParamDatatypesRequired,

		// Action errors.
		"ActionKeyInvalid":               ActionKeyInvalid,
		"ActionKeyTypeInvalid":           ActionKeyTypeInvalid,
		"ActionNameRequired":             ActionNameRequired,
		"ActionNameInvalidChars":         ActionNameInvalidChars,
		"ActionRequiresTypeInvalid":      ActionRequiresTypeInvalid,
		"ActionRequiresDuplicateLet":     ActionRequiresDuplicateLet,
		"ActionGuaranteeTypeInvalid":     ActionGuaranteeTypeInvalid,
		"ActionGuaranteeDuplicateLet":    ActionGuaranteeDuplicateLet,
		"ActionGuaranteeDuplicateTarget": ActionGuaranteeDuplicateTarget,
		"ActionSafetyTypeInvalid":        ActionSafetyTypeInvalid,
		"ActionSafetyDuplicateLet":       ActionSafetyDuplicateLet,

		// Guard errors.
		"GuardKeyInvalid":       GuardKeyInvalid,
		"GuardKeyTypeInvalid":   GuardKeyTypeInvalid,
		"GuardNameRequired":     GuardNameRequired,
		"GuardNameInvalidChars": GuardNameInvalidChars,
		"GuardLogicInvalid":     GuardLogicInvalid,
		"GuardLogicKeyMismatch": GuardLogicKeyMismatch,
		"GuardLogicTypeInvalid": GuardLogicTypeInvalid,

		// Event errors.
		"EventKeyInvalid":       EventKeyInvalid,
		"EventKeyTypeInvalid":   EventKeyTypeInvalid,
		"EventNameRequired":     EventNameRequired,
		"EventNameInvalidChars": EventNameInvalidChars,

		// Query errors.
		"QueryKeyInvalid":               QueryKeyInvalid,
		"QueryKeyTypeInvalid":           QueryKeyTypeInvalid,
		"QueryNameRequired":             QueryNameRequired,
		"QueryNameInvalidChars":         QueryNameInvalidChars,
		"QueryRequiresTypeInvalid":      QueryRequiresTypeInvalid,
		"QueryRequiresDuplicateLet":     QueryRequiresDuplicateLet,
		"QueryGuaranteeTypeInvalid":     QueryGuaranteeTypeInvalid,
		"QueryGuaranteeDuplicateLet":    QueryGuaranteeDuplicateLet,
		"QueryGuaranteeDuplicateTarget": QueryGuaranteeDuplicateTarget,

		// State errors.
		"StateKeyInvalid":       StateKeyInvalid,
		"StateKeyTypeInvalid":   StateKeyTypeInvalid,
		"StateNameRequired":     StateNameRequired,
		"StateNameInvalidChars": StateNameInvalidChars,

		// StateAction errors.
		"StateactionKeyInvalid":       StateactionKeyInvalid,
		"StateactionKeyTypeInvalid":   StateactionKeyTypeInvalid,
		"StateactionWhenRequired":     StateactionWhenRequired,
		"StateactionWhenInvalid":      StateactionWhenInvalid,
		"StateactionActionkeyInvalid": StateactionActionkeyInvalid,
		"StateactionActionkeyType":    StateactionActionkeyType,
		"StateactionActionNotfound":   StateactionActionNotfound,

		// Transition errors.
		"TransitionKeyInvalid":          TransitionKeyInvalid,
		"TransitionKeyTypeInvalid":      TransitionKeyTypeInvalid,
		"TransitionNoState":             TransitionNoState,
		"TransitionFromstatekeyInvalid": TransitionFromstatekeyInvalid,
		"TransitionFromstatekeyType":    TransitionFromstatekeyType,
		"TransitionTostatekeyInvalid":   TransitionTostatekeyInvalid,
		"TransitionTostatekeyType":      TransitionTostatekeyType,
		"TransitionEventkeyInvalid":     TransitionEventkeyInvalid,
		"TransitionEventkeyType":        TransitionEventkeyType,
		"TransitionGuardkeyInvalid":     TransitionGuardkeyInvalid,
		"TransitionGuardkeyType":        TransitionGuardkeyType,
		"TransitionActionkeyInvalid":    TransitionActionkeyInvalid,
		"TransitionActionkeyType":       TransitionActionkeyType,
		"TransitionFromstateNotfound":   TransitionFromstateNotfound,
		"TransitionTostateNotfound":     TransitionTostateNotfound,
		"TransitionEventNotfound":       TransitionEventNotfound,
		"TransitionGuardNotfound":       TransitionGuardNotfound,
		"TransitionActionNotfound":      TransitionActionNotfound,

		// Class errors.
		"ClassKeyInvalid":             ClassKeyInvalid,
		"ClassKeyTypeInvalid":         ClassKeyTypeInvalid,
		"ClassNameRequired":           ClassNameRequired,
		"ClassActorkeyInvalid":        ClassActorkeyInvalid,
		"ClassActorkeyTypeInvalid":    ClassActorkeyTypeInvalid,
		"ClassSuperkeyInvalid":        ClassSuperkeyInvalid,
		"ClassSuperkeyTypeInvalid":    ClassSuperkeyTypeInvalid,
		"ClassSubkeyInvalid":          ClassSubkeyInvalid,
		"ClassSubkeyTypeInvalid":      ClassSubkeyTypeInvalid,
		"ClassSuperSubSame":           ClassSuperSubSame,
		"ClassActorNotfound":          ClassActorNotfound,
		"ClassSupergenNotfound":       ClassSupergenNotfound,
		"ClassSupergenWrongSubdomain": ClassSupergenWrongSubdomain,
		"ClassSubgenNotfound":         ClassSubgenNotfound,
		"ClassSubgenWrongSubdomain":   ClassSubgenWrongSubdomain,
		"ClassInvariantTypeInvalid":   ClassInvariantTypeInvalid,
		"ClassInvariantDuplicateLet":  ClassInvariantDuplicateLet,
		"ClassGuaranteeInvalidTarget": ClassGuaranteeInvalidTarget,

		// Attribute errors.
		"AttrKeyInvalid":            AttrKeyInvalid,
		"AttrKeyTypeInvalid":        AttrKeyTypeInvalid,
		"AttrNameRequired":          AttrNameRequired,
		"AttrDerivationTypeInvalid": AttrDerivationTypeInvalid,
		"AttrInvariantTypeInvalid":  AttrInvariantTypeInvalid,
		"AttrInvariantDuplicateLet": AttrInvariantDuplicateLet,

		// Association errors.
		"AssocKeyInvalid":         AssocKeyInvalid,
		"AssocKeyTypeInvalid":     AssocKeyTypeInvalid,
		"AssocNameRequired":       AssocNameRequired,
		"AssocFromkeyInvalid":     AssocFromkeyInvalid,
		"AssocFromkeyTypeInvalid": AssocFromkeyTypeInvalid,
		"AssocTokeyInvalid":       AssocTokeyInvalid,
		"AssocTokeyTypeInvalid":   AssocTokeyTypeInvalid,
		"AssocFromMultInvalid":    AssocFromMultInvalid,
		"AssocToMultInvalid":      AssocToMultInvalid,
		"AssocAssocclassInvalid":  AssocAssocclassInvalid,
		"AssocAssocclassType":     AssocAssocclassType,
		"AssocAssocclassSameFrom": AssocAssocclassSameFrom,
		"AssocAssocclassSameTo":   AssocAssocclassSameTo,
		"AssocFromNotfound":       AssocFromNotfound,
		"AssocToNotfound":         AssocToNotfound,
		"AssocAssocclassNotfound": AssocAssocclassNotfound,

		// ClassGeneralization errors.
		"CgenKeyInvalid":     CgenKeyInvalid,
		"CgenKeyTypeInvalid": CgenKeyTypeInvalid,
		"CgenNameRequired":   CgenNameRequired,

		// Actor errors.
		"ActorKeyInvalid":          ActorKeyInvalid,
		"ActorKeyTypeInvalid":      ActorKeyTypeInvalid,
		"ActorNameRequired":        ActorNameRequired,
		"ActorTypeRequired":        ActorTypeRequired,
		"ActorTypeInvalid":         ActorTypeInvalid,
		"ActorSuperkeyInvalid":     ActorSuperkeyInvalid,
		"ActorSuperkeyTypeInvalid": ActorSuperkeyTypeInvalid,
		"ActorSubkeyInvalid":       ActorSubkeyInvalid,
		"ActorSubkeyTypeInvalid":   ActorSubkeyTypeInvalid,
		"ActorSuperSubSame":        ActorSuperSubSame,
		"ActorSupergenNotfound":    ActorSupergenNotfound,
		"ActorSubgenNotfound":      ActorSubgenNotfound,

		// ActorGeneralization errors.
		"AgenKeyInvalid":     AgenKeyInvalid,
		"AgenKeyTypeInvalid": AgenKeyTypeInvalid,
		"AgenNameRequired":   AgenNameRequired,

		// Scenario errors.
		"ScenarioKeyInvalid":     ScenarioKeyInvalid,
		"ScenarioKeyTypeInvalid": ScenarioKeyTypeInvalid,
		"ScenarioNameRequired":   ScenarioNameRequired,

		// ScenarioObject errors.
		"SobjectKeyInvalid":          SobjectKeyInvalid,
		"SobjectKeyTypeInvalid":      SobjectKeyTypeInvalid,
		"SobjectNamestyleRequired":   SobjectNamestyleRequired,
		"SobjectNamestyleInvalid":    SobjectNamestyleInvalid,
		"SobjectNameRequired":        SobjectNameRequired,
		"SobjectNameMustBeBlank":     SobjectNameMustBeBlank,
		"SobjectClasskeyInvalid":     SobjectClasskeyInvalid,
		"SobjectClasskeyTypeInvalid": SobjectClasskeyTypeInvalid,
		"SobjectClassNotfound":       SobjectClassNotfound,

		// ScenarioStep errors.
		"SstepKeyInvalid":             SstepKeyInvalid,
		"SstepKeyTypeInvalid":         SstepKeyTypeInvalid,
		"SstepTypeUnknown":            SstepTypeUnknown,
		"SstepLeafTypeRequired":       SstepLeafTypeRequired,
		"SstepLeafTypeUnknown":        SstepLeafTypeUnknown,
		"SstepFromkeyInvalid":         SstepFromkeyInvalid,
		"SstepFromkeyTypeInvalid":     SstepFromkeyTypeInvalid,
		"SstepTokeyInvalid":           SstepTokeyInvalid,
		"SstepTokeyTypeInvalid":       SstepTokeyTypeInvalid,
		"SstepEventkeyInvalid":        SstepEventkeyInvalid,
		"SstepEventkeyTypeInvalid":    SstepEventkeyTypeInvalid,
		"SstepQuerykeyInvalid":        SstepQuerykeyInvalid,
		"SstepQuerykeyTypeInvalid":    SstepQuerykeyTypeInvalid,
		"SstepScenariokeyInvalid":     SstepScenariokeyInvalid,
		"SstepScenariokeyTypeInvalid": SstepScenariokeyTypeInvalid,
		"SstepEventFromRequired":      SstepEventFromRequired,
		"SstepEventToRequired":        SstepEventToRequired,
		"SstepEventKeyRequired":       SstepEventKeyRequired,
		"SstepEventQueryForbidden":    SstepEventQueryForbidden,
		"SstepQueryFromRequired":      SstepQueryFromRequired,
		"SstepQueryToRequired":        SstepQueryToRequired,
		"SstepQueryKeyRequired":       SstepQueryKeyRequired,
		"SstepQueryEventForbidden":    SstepQueryEventForbidden,
		"SstepScenarioFromRequired":   SstepScenarioFromRequired,
		"SstepScenarioToRequired":     SstepScenarioToRequired,
		"SstepScenarioKeyRequired":    SstepScenarioKeyRequired,
		"SstepScenarioEventForbidden": SstepScenarioEventForbidden,
		"SstepScenarioSelfRef":        SstepScenarioSelfRef,
		"SstepDeleteFromRequired":     SstepDeleteFromRequired,
		"SstepDeleteToForbidden":      SstepDeleteToForbidden,
		"SstepDeleteKeysForbidden":    SstepDeleteKeysForbidden,
		"SstepSequenceMinStatements":  SstepSequenceMinStatements,
		"SstepSwitchMinCases":         SstepSwitchMinCases,
		"SstepSwitchCaseType":         SstepSwitchCaseType,
		"SstepCaseConditionRequired":  SstepCaseConditionRequired,
		"SstepLoopConditionRequired":  SstepLoopConditionRequired,
		"SstepLoopMinStatements":      SstepLoopMinStatements,

		// UseCase errors.
		"UcKeyInvalid":          UcKeyInvalid,
		"UcKeyTypeInvalid":      UcKeyTypeInvalid,
		"UcNameRequired":        UcNameRequired,
		"UcLevelRequired":       UcLevelRequired,
		"UcLevelInvalid":        UcLevelInvalid,
		"UcSuperkeyInvalid":     UcSuperkeyInvalid,
		"UcSuperkeyTypeInvalid": UcSuperkeyTypeInvalid,
		"UcSubkeyInvalid":       UcSubkeyInvalid,
		"UcSubkeyTypeInvalid":   UcSubkeyTypeInvalid,
		"UcSuperSubSame":        UcSuperSubSame,
		"UcActorNotActorClass":  UcActorNotActorClass,
		"UcSupergenNotfound":    UcSupergenNotfound,
		"UcSubgenNotfound":      UcSubgenNotfound,

		// UseCaseGeneralization errors.
		"UcgenKeyInvalid":     UcgenKeyInvalid,
		"UcgenKeyTypeInvalid": UcgenKeyTypeInvalid,
		"UcgenNameRequired":   UcgenNameRequired,

		// UseCaseShared errors.
		"UshareSharetypeInvalid": UshareSharetypeInvalid,

		// Domain errors.
		"DomainKeyInvalid":            DomainKeyInvalid,
		"DomainKeyTypeInvalid":        DomainKeyTypeInvalid,
		"DomainNameRequired":          DomainNameRequired,
		"DomainSubdomainSingleKey":    DomainSubdomainSingleKey,
		"DomainSubdomainMultiDefault": DomainSubdomainMultiDefault,
		"DomainCassocOrphan":          DomainCassocOrphan,

		// DomainAssociation errors.
		"DassocKeyInvalid":         DassocKeyInvalid,
		"DassocKeyTypeInvalid":     DassocKeyTypeInvalid,
		"DassocProblemkeyInvalid":  DassocProblemkeyInvalid,
		"DassocProblemkeyType":     DassocProblemkeyType,
		"DassocSolutionkeyInvalid": DassocSolutionkeyInvalid,
		"DassocSolutionkeyType":    DassocSolutionkeyType,
		"DassocSameDomains":        DassocSameDomains,
		"DassocProblemNotfound":    DassocProblemNotfound,
		"DassocSolutionNotfound":   DassocSolutionNotfound,

		// Subdomain errors.
		"SubdomainKeyInvalid":             SubdomainKeyInvalid,
		"SubdomainKeyTypeInvalid":         SubdomainKeyTypeInvalid,
		"SubdomainNameRequired":           SubdomainNameRequired,
		"SubdomainCassocNoParent":         SubdomainCassocNoParent,
		"SubdomainCassocWrongParent":      SubdomainCassocWrongParent,
		"SubdomainCgenSuperclassCount":    SubdomainCgenSuperclassCount,
		"SubdomainCgenSubclassCount":      SubdomainCgenSubclassCount,
		"SubdomainUcgenSuperclassCount":   SubdomainUcgenSuperclassCount,
		"SubdomainUcgenSubclassCount":     SubdomainUcgenSubclassCount,
		"SubdomainUshareSealevelNotfound": SubdomainUshareSealevelNotfound,
		"SubdomainUshareMudlevelNotfound": SubdomainUshareMudlevelNotfound,
	}
}
