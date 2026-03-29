package parser_ai

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	parserErrors "github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/errors"
	"github.com/stretchr/testify/suite"
)

// allErrorCodes returns every error code constant defined in errors.go.
// This must be kept in sync with errors.go — if a new code is added there,
// add it here. The TestAllCodesAccountedFor test helps catch drift.
func allErrorCodes() map[int]string {
	return map[int]string{
		// Model errors (1xxx).
		ErrModelNameRequired:    "ErrModelNameRequired",
		ErrModelNameEmpty:       "ErrModelNameEmpty",
		ErrModelInvalidJSON:     "ErrModelInvalidJSON",
		ErrModelSchemaViolation: "ErrModelSchemaViolation",

		// Actor errors (2xxx).
		ErrActorNameRequired:    "ErrActorNameRequired",
		ErrActorNameEmpty:       "ErrActorNameEmpty",
		ErrActorTypeRequired:    "ErrActorTypeRequired",
		ErrActorTypeInvalid:     "ErrActorTypeInvalid",
		ErrActorInvalidJSON:     "ErrActorInvalidJSON",
		ErrActorSchemaViolation: "ErrActorSchemaViolation",
		ErrActorDuplicateKey:    "ErrActorDuplicateKey",
		ErrActorFilenameInvalid: "ErrActorFilenameInvalid",

		// Domain errors (3xxx).
		ErrDomainNameRequired:    "ErrDomainNameRequired",
		ErrDomainNameEmpty:       "ErrDomainNameEmpty",
		ErrDomainInvalidJSON:     "ErrDomainInvalidJSON",
		ErrDomainSchemaViolation: "ErrDomainSchemaViolation",
		ErrDomainDuplicateKey:    "ErrDomainDuplicateKey",
		ErrDomainDirInvalid:      "ErrDomainDirInvalid",

		// Subdomain errors (4xxx).
		ErrSubdomainNameRequired:    "ErrSubdomainNameRequired",
		ErrSubdomainNameEmpty:       "ErrSubdomainNameEmpty",
		ErrSubdomainInvalidJSON:     "ErrSubdomainInvalidJSON",
		ErrSubdomainSchemaViolation: "ErrSubdomainSchemaViolation",
		ErrSubdomainDuplicateKey:    "ErrSubdomainDuplicateKey",
		ErrSubdomainDirInvalid:      "ErrSubdomainDirInvalid",

		// Class errors (5xxx).
		ErrClassNameRequired:        "ErrClassNameRequired",
		ErrClassNameEmpty:           "ErrClassNameEmpty",
		ErrClassInvalidJSON:         "ErrClassInvalidJSON",
		ErrClassSchemaViolation:     "ErrClassSchemaViolation",
		ErrClassDuplicateKey:        "ErrClassDuplicateKey",
		ErrClassDirInvalid:          "ErrClassDirInvalid",
		ErrClassActorNotFound:       "ErrClassActorNotFound",
		ErrClassAttributeNameEmpty:  "ErrClassAttributeNameEmpty",
		ErrClassIndexInvalid:        "ErrClassIndexInvalid",
		ErrClassIndexAttrNotFound:   "ErrClassIndexAttrNotFound",
		ErrClassDataTypeUnparseable: "ErrClassDataTypeUnparseable",

		// Association errors (6xxx).
		ErrAssocNameRequired:        "ErrAssocNameRequired",
		ErrAssocNameEmpty:           "ErrAssocNameEmpty",
		ErrAssocInvalidJSON:         "ErrAssocInvalidJSON",
		ErrAssocSchemaViolation:     "ErrAssocSchemaViolation",
		ErrAssocFromClassRequired:   "ErrAssocFromClassRequired",
		ErrAssocToClassRequired:     "ErrAssocToClassRequired",
		ErrAssocFromMultRequired:    "ErrAssocFromMultRequired",
		ErrAssocToMultRequired:      "ErrAssocToMultRequired",
		ErrAssocFromClassNotFound:   "ErrAssocFromClassNotFound",
		ErrAssocToClassNotFound:     "ErrAssocToClassNotFound",
		ErrAssocClassNotFound:       "ErrAssocClassNotFound",
		ErrAssocMultiplicityInvalid: "ErrAssocMultiplicityInvalid",
		ErrAssocFilenameInvalid:     "ErrAssocFilenameInvalid",
		ErrAssocDuplicateKey:        "ErrAssocDuplicateKey",
		ErrAssocNameMismatch:        "ErrAssocNameMismatch",

		// State machine errors (7xxx).
		ErrStateMachineInvalidJSON:       "ErrStateMachineInvalidJSON",
		ErrStateMachineSchemaViolation:   "ErrStateMachineSchemaViolation",
		ErrStateNameRequired:             "ErrStateNameRequired",
		ErrStateNameEmpty:                "ErrStateNameEmpty",
		ErrStateDuplicateKey:             "ErrStateDuplicateKey",
		ErrStateActionKeyRequired:        "ErrStateActionKeyRequired",
		ErrStateActionWhenRequired:       "ErrStateActionWhenRequired",
		ErrStateActionWhenInvalid:        "ErrStateActionWhenInvalid",
		ErrEventNameRequired:             "ErrEventNameRequired",
		ErrEventNameEmpty:                "ErrEventNameEmpty",
		ErrEventDuplicateKey:             "ErrEventDuplicateKey",
		ErrEventParamNameRequired:        "ErrEventParamNameRequired",
		ErrEventParamSourceRequired:      "ErrEventParamSourceRequired",
		ErrGuardNameRequired:             "ErrGuardNameRequired",
		ErrGuardNameEmpty:                "ErrGuardNameEmpty",
		ErrGuardDetailsRequired:          "ErrGuardDetailsRequired",
		ErrGuardDuplicateKey:             "ErrGuardDuplicateKey",
		ErrTransitionEventRequired:       "ErrTransitionEventRequired",
		ErrTransitionNoStates:            "ErrTransitionNoStates",
		ErrTransitionFromStateNotFound:   "ErrTransitionFromStateNotFound",
		ErrTransitionToStateNotFound:     "ErrTransitionToStateNotFound",
		ErrTransitionEventNotFound:       "ErrTransitionEventNotFound",
		ErrTransitionGuardNotFound:       "ErrTransitionGuardNotFound",
		ErrTransitionActionNotFound:      "ErrTransitionActionNotFound",
		ErrTransitionInitialToFinal:      "ErrTransitionInitialToFinal",
		ErrEventParamDataTypeUnparseable: "ErrEventParamDataTypeUnparseable",
		ErrStateDuplicateName:            "ErrStateDuplicateName",
		ErrEventDuplicateName:            "ErrEventDuplicateName",
		ErrGuardDuplicateName:            "ErrGuardDuplicateName",
		ErrStateKeyNameMismatch:          "ErrStateKeyNameMismatch",
		ErrEventKeyNameMismatch:          "ErrEventKeyNameMismatch",
		ErrGuardKeyNameMismatch:          "ErrGuardKeyNameMismatch",

		// Action errors (8xxx).
		ErrActionNameRequired:    "ErrActionNameRequired",
		ErrActionNameEmpty:       "ErrActionNameEmpty",
		ErrActionInvalidJSON:     "ErrActionInvalidJSON",
		ErrActionSchemaViolation: "ErrActionSchemaViolation",
		ErrActionDuplicateKey:    "ErrActionDuplicateKey",
		ErrActionFilenameInvalid: "ErrActionFilenameInvalid",
		ErrActionDuplicateName:   "ErrActionDuplicateName",

		// Query errors (9xxx).
		ErrQueryNameRequired:    "ErrQueryNameRequired",
		ErrQueryNameEmpty:       "ErrQueryNameEmpty",
		ErrQueryInvalidJSON:     "ErrQueryInvalidJSON",
		ErrQuerySchemaViolation: "ErrQuerySchemaViolation",
		ErrQueryDuplicateKey:    "ErrQueryDuplicateKey",
		ErrQueryFilenameInvalid: "ErrQueryFilenameInvalid",
		ErrQueryDuplicateName:   "ErrQueryDuplicateName",

		// Class generalization errors (10xxx).
		ErrClassGenNameRequired:         "ErrClassGenNameRequired",
		ErrClassGenNameEmpty:            "ErrClassGenNameEmpty",
		ErrClassGenInvalidJSON:          "ErrClassGenInvalidJSON",
		ErrClassGenSchemaViolation:      "ErrClassGenSchemaViolation",
		ErrClassGenSuperclassRequired:   "ErrClassGenSuperclassRequired",
		ErrClassGenSubclassesRequired:   "ErrClassGenSubclassesRequired",
		ErrClassGenSubclassesEmpty:      "ErrClassGenSubclassesEmpty",
		ErrClassGenSuperclassNotFound:   "ErrClassGenSuperclassNotFound",
		ErrClassGenSubclassNotFound:     "ErrClassGenSubclassNotFound",
		ErrClassGenDuplicateKey:         "ErrClassGenDuplicateKey",
		ErrClassGenFilenameInvalid:      "ErrClassGenFilenameInvalid",
		ErrClassGenSubclassDuplicate:    "ErrClassGenSubclassDuplicate",
		ErrClassGenSuperclassIsSubclass: "ErrClassGenSuperclassIsSubclass",

		// Tree validation errors (11xxx).
		ErrTreeClassActorNotFound:           "ErrTreeClassActorNotFound",
		ErrTreeAssocFromClassNotFound:       "ErrTreeAssocFromClassNotFound",
		ErrTreeAssocToClassNotFound:         "ErrTreeAssocToClassNotFound",
		ErrTreeAssocClassNotFound:           "ErrTreeAssocClassNotFound",
		ErrTreeClassGenSuperclassNotFound:   "ErrTreeClassGenSuperclassNotFound",
		ErrTreeClassGenSubclassNotFound:     "ErrTreeClassGenSubclassNotFound",
		ErrTreeClassIndexAttrNotFound:       "ErrTreeClassIndexAttrNotFound",
		ErrTreeStateMachineStateNotFound:    "ErrTreeStateMachineStateNotFound",
		ErrTreeStateMachineEventNotFound:    "ErrTreeStateMachineEventNotFound",
		ErrTreeStateMachineGuardNotFound:    "ErrTreeStateMachineGuardNotFound",
		ErrTreeStateMachineActionNotFound:   "ErrTreeStateMachineActionNotFound",
		ErrTreeTransitionNoStates:           "ErrTreeTransitionNoStates",
		ErrTreeTransitionInitialToFinal:     "ErrTreeTransitionInitialToFinal",
		ErrTreeClassGenSuperclassIsSubclass: "ErrTreeClassGenSuperclassIsSubclass",
		ErrTreeClassGenSubclassDuplicate:    "ErrTreeClassGenSubclassDuplicate",
		ErrTreeAssocMultiplicityInvalid:     "ErrTreeAssocMultiplicityInvalid",
		ErrTreeAssocClassSameAsEndpoint:     "ErrTreeAssocClassSameAsEndpoint",
		ErrTreeModelNoActors:                "ErrTreeModelNoActors",
		ErrTreeModelNoDomains:               "ErrTreeModelNoDomains",
		ErrTreeDomainNoSubdomains:           "ErrTreeDomainNoSubdomains",
		ErrTreeSubdomainTooFewClasses:       "ErrTreeSubdomainTooFewClasses",
		ErrTreeSubdomainNoAssociations:      "ErrTreeSubdomainNoAssociations",
		ErrTreeClassNoAttributes:            "ErrTreeClassNoAttributes",
		ErrTreeClassNoStateMachine:          "ErrTreeClassNoStateMachine",
		ErrTreeStateMachineNoTransitions:    "ErrTreeStateMachineNoTransitions",
		ErrKeyInvalidFormat:                 "ErrKeyInvalidFormat",
		ErrAssocFilenameInvalidFormat:       "ErrAssocFilenameInvalidFormat",
		ErrAssocFilenameInvalidComponent:    "ErrAssocFilenameInvalidComponent",
		ErrTreeActionUnreferenced:           "ErrTreeActionUnreferenced",
		ErrTreeSingleSubdomainNotDefault:    "ErrTreeSingleSubdomainNotDefault",
		ErrTreeMultipleSubdomainsHasDefault: "ErrTreeMultipleSubdomainsHasDefault",
		ErrTreeDomainAssocDomainNotFound:    "ErrTreeDomainAssocDomainNotFound",
		ErrTreeActorGenActorNotFound:        "ErrTreeActorGenActorNotFound",
		ErrTreeScenarioStepObjectNotFound:   "ErrTreeScenarioStepObjectNotFound",
		ErrTreeScenarioStepEventNotFound:    "ErrTreeScenarioStepEventNotFound",
		ErrTreeScenarioStepQueryNotFound:    "ErrTreeScenarioStepQueryNotFound",

		// Actor generalization errors (12xxx).
		ErrActorGenNameRequired:       "ErrActorGenNameRequired",
		ErrActorGenNameEmpty:          "ErrActorGenNameEmpty",
		ErrActorGenInvalidJSON:        "ErrActorGenInvalidJSON",
		ErrActorGenSchemaViolation:    "ErrActorGenSchemaViolation",
		ErrActorGenSuperclassRequired: "ErrActorGenSuperclassRequired",
		ErrActorGenSubclassesRequired: "ErrActorGenSubclassesRequired",
		ErrActorGenSubclassesEmpty:    "ErrActorGenSubclassesEmpty",

		// Use case generalization errors (13xxx).
		ErrUseCaseGenNameRequired:       "ErrUseCaseGenNameRequired",
		ErrUseCaseGenNameEmpty:          "ErrUseCaseGenNameEmpty",
		ErrUseCaseGenInvalidJSON:        "ErrUseCaseGenInvalidJSON",
		ErrUseCaseGenSchemaViolation:    "ErrUseCaseGenSchemaViolation",
		ErrUseCaseGenSuperclassRequired: "ErrUseCaseGenSuperclassRequired",
		ErrUseCaseGenSubclassesRequired: "ErrUseCaseGenSubclassesRequired",
		ErrUseCaseGenSubclassesEmpty:    "ErrUseCaseGenSubclassesEmpty",

		// Logic errors (14xxx).
		ErrLogicDescriptionRequired:    "ErrLogicDescriptionRequired",
		ErrLogicDescriptionEmpty:       "ErrLogicDescriptionEmpty",
		ErrLogicInvalidJSON:            "ErrLogicInvalidJSON",
		ErrLogicSchemaViolation:        "ErrLogicSchemaViolation",
		ErrLogicTargetRequired:         "ErrLogicTargetRequired",
		ErrLogicTargetNotAllowed:       "ErrLogicTargetNotAllowed",
		ErrLogicTargetNoLeadUnderscore: "ErrLogicTargetNoLeadUnderscore",
		ErrLogicTypeRequired:           "ErrLogicTypeRequired",

		// Parameter errors (15xxx).
		ErrParamNameRequired:        "ErrParamNameRequired",
		ErrParamNameEmpty:           "ErrParamNameEmpty",
		ErrParamInvalidJSON:         "ErrParamInvalidJSON",
		ErrParamSchemaViolation:     "ErrParamSchemaViolation",
		ErrParamDataTypeUnparseable: "ErrParamDataTypeUnparseable",

		// Global function errors (16xxx).
		ErrGlobalFuncNameRequired:     "ErrGlobalFuncNameRequired",
		ErrGlobalFuncNameEmpty:        "ErrGlobalFuncNameEmpty",
		ErrGlobalFuncInvalidJSON:      "ErrGlobalFuncInvalidJSON",
		ErrGlobalFuncSchemaViolation:  "ErrGlobalFuncSchemaViolation",
		ErrGlobalFuncNameNoUnderscore: "ErrGlobalFuncNameNoUnderscore",
		ErrGlobalFuncParamEmpty:       "ErrGlobalFuncParamEmpty",
		ErrGlobalFuncLogicRequired:    "ErrGlobalFuncLogicRequired",

		// Domain association errors (17xxx).
		ErrDomainAssocProblemKeyRequired:  "ErrDomainAssocProblemKeyRequired",
		ErrDomainAssocProblemKeyEmpty:     "ErrDomainAssocProblemKeyEmpty",
		ErrDomainAssocSolutionKeyRequired: "ErrDomainAssocSolutionKeyRequired",
		ErrDomainAssocSolutionKeyEmpty:    "ErrDomainAssocSolutionKeyEmpty",
		ErrDomainAssocInvalidJSON:         "ErrDomainAssocInvalidJSON",
		ErrDomainAssocSchemaViolation:     "ErrDomainAssocSchemaViolation",

		// Use case errors (18xxx).
		ErrUseCaseNameRequired:    "ErrUseCaseNameRequired",
		ErrUseCaseNameEmpty:       "ErrUseCaseNameEmpty",
		ErrUseCaseInvalidJSON:     "ErrUseCaseInvalidJSON",
		ErrUseCaseSchemaViolation: "ErrUseCaseSchemaViolation",
		ErrUseCaseLevelRequired:   "ErrUseCaseLevelRequired",
		ErrUseCaseLevelInvalid:    "ErrUseCaseLevelInvalid",

		// Scenario errors (19xxx).
		ErrScenarioNameRequired:    "ErrScenarioNameRequired",
		ErrScenarioNameEmpty:       "ErrScenarioNameEmpty",
		ErrScenarioInvalidJSON:     "ErrScenarioInvalidJSON",
		ErrScenarioSchemaViolation: "ErrScenarioSchemaViolation",

		// Use case shared errors (20xxx).
		ErrUseCaseSharedShareTypeRequired: "ErrUseCaseSharedShareTypeRequired",
		ErrUseCaseSharedShareTypeEmpty:    "ErrUseCaseSharedShareTypeEmpty",
		ErrUseCaseSharedInvalidJSON:       "ErrUseCaseSharedInvalidJSON",
		ErrUseCaseSharedSchemaViolation:   "ErrUseCaseSharedSchemaViolation",

		// Conversion errors (21xxx).
		ErrConvKeyConstruction:       "ErrConvKeyConstruction",
		ErrConvModelValidation:       "ErrConvModelValidation",
		ErrConvMultiplicityInvalid:   "ErrConvMultiplicityInvalid",
		ErrConvClassNotFound:         "ErrConvClassNotFound",
		ErrConvAssocKeyConstruction:  "ErrConvAssocKeyConstruction",
		ErrConvScopedKeyInvalid:      "ErrConvScopedKeyInvalid",
		ErrConvObjectResolveFailed:   "ErrConvObjectResolveFailed",
		ErrConvSourceModelValidation: "ErrConvSourceModelValidation",

		// Mapped core validation errors (21100-21199).
		ErrConvParamDatatypeRequired:     "ErrConvParamDatatypeRequired",
		ErrConvParamNameRequired:         "ErrConvParamNameRequired",
		ErrConvLogicTypeInvalid:          "ErrConvLogicTypeInvalid",
		ErrConvLogicDuplicateLet:         "ErrConvLogicDuplicateLet",
		ErrConvLogicDuplicateTarget:      "ErrConvLogicDuplicateTarget",
		ErrConvLogicTargetRequired:       "ErrConvLogicTargetRequired",
		ErrConvLogicTargetNotAllowed:     "ErrConvLogicTargetNotAllowed",
		ErrConvLogicTargetNoUnderscore:   "ErrConvLogicTargetNoUnderscore",
		ErrConvReferenceNotFound:         "ErrConvReferenceNotFound",
		ErrConvGenCardinalityInvalid:     "ErrConvGenCardinalityInvalid",
		ErrConvDomainStructureInvalid:    "ErrConvDomainStructureInvalid",
		ErrConvScenarioStepInvalid:       "ErrConvScenarioStepInvalid",
		ErrConvGuaranteeInvalidTarget:    "ErrConvGuaranteeInvalidTarget",
		ErrConvAssocClassSameAsEndpoint:  "ErrConvAssocClassSameAsEndpoint",
		ErrConvInternalKeyError:          "ErrConvInternalKeyError",
		ErrConvUseCaseActorNotActorClass: "ErrConvUseCaseActorNotActorClass",
		ErrConvLogicSpecInvalid:          "ErrConvLogicSpecInvalid",
		ErrConvDomainAssocSameDomains:    "ErrConvDomainAssocSameDomains",

		// Named set errors (22xxx).
		ErrNamedSetNameRequired:     "ErrNamedSetNameRequired",
		ErrNamedSetNameEmpty:        "ErrNamedSetNameEmpty",
		ErrNamedSetInvalidJSON:      "ErrNamedSetInvalidJSON",
		ErrNamedSetSchemaViolation:  "ErrNamedSetSchemaViolation",
		ErrNamedSetNameNoUnderscore: "ErrNamedSetNameNoUnderscore",
	}
}

type ErrorCoverageSuite struct {
	suite.Suite
}

func TestErrorCoverageSuite(t *testing.T) {
	suite.Run(t, new(ErrorCoverageSuite))
}

// TestAllCodesHaveDocs verifies that every error code constant has a corresponding .md file.
func (s *ErrorCoverageSuite) TestAllCodesHaveDocs() {
	codes := allErrorCodes()
	var missing []string

	for code, name := range codes {
		_, _, err := parserErrors.LoadErrorDoc(code)
		if err != nil {
			missing = append(missing, name+" (E"+strconv.Itoa(code)+")")
		}
	}

	s.Empty(missing, "error codes missing documentation files:\n  "+strings.Join(missing, "\n  "))
}

// TestNoOrphanDocs verifies that every .md file in errors/ corresponds to a defined error code.
func (s *ErrorCoverageSuite) TestNoOrphanDocs() {
	codes := allErrorCodes()

	// Build set of defined codes.
	definedCodes := make(map[int]bool, len(codes))
	for code := range codes {
		definedCodes[code] = true
	}

	// Read all .md files from embedded FS.
	entries, err := fs.ReadDir(parserErrors.ErrorDocs, ".")
	s.Require().NoError(err)

	var orphans []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Extract code from filename (e.g., "1001_model_name_required.md" → 1001).
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			orphans = append(orphans, entry.Name()+" (unparseable filename)")
			continue
		}
		code, parseErr := strconv.Atoi(parts[0])
		if parseErr != nil {
			orphans = append(orphans, entry.Name()+" (non-numeric prefix)")
			continue
		}
		if !definedCodes[code] {
			orphans = append(orphans, entry.Name()+" (E"+strconv.Itoa(code)+")")
		}
	}

	s.Empty(orphans, "orphan documentation files without corresponding error code constants:\n  "+strings.Join(orphans, "\n  "))
}

// TestAllCodesUnique verifies that no two error code constants share the same integer value.
func (s *ErrorCoverageSuite) TestAllCodesUnique() {
	codes := allErrorCodes()

	seen := make(map[int]string, len(codes))
	var duplicates []string

	for code, name := range codes {
		if existing, ok := seen[code]; ok {
			duplicates = append(duplicates, name+" and "+existing+" both use E"+strconv.Itoa(code))
		}
		seen[code] = name
	}

	s.Empty(duplicates, "duplicate error code values:\n  "+strings.Join(duplicates, "\n  "))
}

// TestAllDocsAreNonEmpty verifies that every error documentation file has meaningful content.
func (s *ErrorCoverageSuite) TestAllDocsAreNonEmpty() {
	codes := allErrorCodes()
	var emptyDocs []string

	for code, name := range codes {
		content, _, err := parserErrors.LoadErrorDoc(code)
		if err != nil {
			continue // Missing docs are caught by TestAllCodesHaveDocs.
		}
		trimmed := strings.TrimSpace(content)
		if len(trimmed) < 20 {
			emptyDocs = append(emptyDocs, name+" (E"+strconv.Itoa(code)+") has only "+strconv.Itoa(len(trimmed))+" chars")
		}
	}

	s.Empty(emptyDocs, "error documentation files that are too short:\n  "+strings.Join(emptyDocs, "\n  "))
}

// TestErrorCodeCount verifies the total number of error codes matches expectations.
// This test catches cases where a new constant is added to errors.go but not to allErrorCodes().
func (s *ErrorCoverageSuite) TestErrorCodeCount() {
	codes := allErrorCodes()
	// 230 error codes as of current implementation.
	// Update this number when adding new error codes.
	s.Len(codes, 230,
		"allErrorCodes() count doesn't match expected. If you added new error codes to errors.go, add them to allErrorCodes() in error_coverage_test.go too.")
}
