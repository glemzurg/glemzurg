package parser_ai

// Error codes for AI input validation errors.
// Each error type has a unique identifier for programmatic handling.
const (
	// Model errors (1xxx).
	ErrModelNameRequired    = 1001
	ErrModelNameEmpty       = 1002
	ErrModelInvalidJSON     = 1003
	ErrModelSchemaViolation = 1004

	// Actor errors (2xxx).
	ErrActorNameRequired    = 2001
	ErrActorNameEmpty       = 2002
	ErrActorTypeRequired    = 2003
	ErrActorTypeInvalid     = 2004
	ErrActorInvalidJSON     = 2005
	ErrActorSchemaViolation = 2006
	ErrActorDuplicateKey    = 2007
	ErrActorFilenameInvalid = 2008

	// Domain errors (3xxx).
	ErrDomainNameRequired    = 3001
	ErrDomainNameEmpty       = 3002
	ErrDomainInvalidJSON     = 3003
	ErrDomainSchemaViolation = 3004
	ErrDomainDuplicateKey    = 3005
	ErrDomainDirInvalid      = 3006

	// Subdomain errors (4xxx).
	ErrSubdomainNameRequired    = 4001
	ErrSubdomainNameEmpty       = 4002
	ErrSubdomainInvalidJSON     = 4003
	ErrSubdomainSchemaViolation = 4004
	ErrSubdomainDuplicateKey    = 4005
	ErrSubdomainDirInvalid      = 4006

	// Class errors (5xxx).
	ErrClassNameRequired       = 5001
	ErrClassNameEmpty          = 5002
	ErrClassInvalidJSON        = 5003
	ErrClassSchemaViolation    = 5004
	ErrClassDuplicateKey       = 5005
	ErrClassDirInvalid         = 5006
	ErrClassActorNotFound      = 5007
	ErrClassAttributeNameEmpty = 5008
	ErrClassIndexInvalid       = 5009
	ErrClassIndexAttrNotFound  = 5010

	// Association errors (6xxx).
	ErrAssocNameRequired                = 6001
	ErrAssocNameEmpty                   = 6002
	ErrAssocInvalidJSON                 = 6003
	ErrAssocSchemaViolation             = 6004
	ErrAssocFromClassRequired           = 6005
	ErrAssocToClassRequired             = 6006
	ErrAssocFromMultRequired            = 6007
	ErrAssocToMultRequired              = 6008
	ErrAssocFromClassNotFound           = 6009
	ErrAssocToClassNotFound             = 6010
	ErrAssocClassNotFound               = 6011
	ErrAssocMultiplicityInvalid         = 6012
	ErrAssocFilenameInvalid             = 6013
	ErrAssocDuplicateKey                = 6014
	ErrAssocUniquenessConstraintInvalid = 6016

	// State machine errors (7xxx).
	ErrStateMachineInvalidJSON     = 7001
	ErrStateMachineSchemaViolation = 7002
	ErrStateNameRequired           = 7003
	ErrStateNameEmpty              = 7004
	ErrStateDuplicateKey           = 7005
	ErrStateActionKeyRequired      = 7006
	ErrStateActionWhenRequired     = 7007
	ErrStateActionWhenInvalid      = 7008
	ErrEventNameRequired           = 7009
	ErrEventNameEmpty              = 7010
	ErrEventDuplicateKey           = 7011
	ErrEventParamNameRequired      = 7012
	ErrEventParamSourceRequired    = 7013
	ErrGuardNameRequired           = 7014
	ErrGuardNameEmpty              = 7015
	ErrGuardDetailsRequired        = 7016
	ErrGuardDuplicateKey           = 7017
	ErrTransitionEventRequired     = 7018
	ErrTransitionNoStates          = 7019
	ErrTransitionFromStateNotFound = 7020
	ErrTransitionToStateNotFound   = 7021
	ErrTransitionEventNotFound     = 7022
	ErrTransitionGuardNotFound     = 7023
	ErrTransitionActionNotFound    = 7024
	ErrTransitionInitialToFinal    = 7025

	// Action errors (8xxx).
	ErrActionNameRequired    = 8001
	ErrActionNameEmpty       = 8002
	ErrActionInvalidJSON     = 8003
	ErrActionSchemaViolation = 8004
	ErrActionDuplicateKey    = 8005
	ErrActionFilenameInvalid = 8006

	// Query errors (9xxx).
	ErrQueryNameRequired    = 9001
	ErrQueryNameEmpty       = 9002
	ErrQueryInvalidJSON     = 9003
	ErrQuerySchemaViolation = 9004
	ErrQueryDuplicateKey    = 9005
	ErrQueryFilenameInvalid = 9006

	// Class generalization errors (10xxx).
	ErrClassGenNameRequired         = 10001
	ErrClassGenNameEmpty            = 10002
	ErrClassGenInvalidJSON          = 10003
	ErrClassGenSchemaViolation      = 10004
	ErrClassGenSuperclassRequired   = 10005
	ErrClassGenSubclassesRequired   = 10006
	ErrClassGenSubclassesEmpty      = 10007
	ErrClassGenSuperclassNotFound   = 10008
	ErrClassGenSubclassNotFound     = 10009
	ErrClassGenDuplicateKey         = 10010
	ErrClassGenFilenameInvalid      = 10011
	ErrClassGenSubclassDuplicate    = 10012
	ErrClassGenSuperclassIsSubclass = 10013

	// Actor generalization errors (12xxx).
	ErrActorGenNameRequired       = 12001
	ErrActorGenNameEmpty          = 12002
	ErrActorGenInvalidJSON        = 12003
	ErrActorGenSchemaViolation    = 12004
	ErrActorGenSuperclassRequired = 12005
	ErrActorGenSubclassesRequired = 12006
	ErrActorGenSubclassesEmpty    = 12007

	// Use case generalization errors (13xxx).
	ErrUseCaseGenNameRequired       = 13001
	ErrUseCaseGenNameEmpty          = 13002
	ErrUseCaseGenInvalidJSON        = 13003
	ErrUseCaseGenSchemaViolation    = 13004
	ErrUseCaseGenSuperclassRequired = 13005
	ErrUseCaseGenSubclassesRequired = 13006
	ErrUseCaseGenSubclassesEmpty    = 13007

	// Logic errors (14xxx).
	ErrLogicDescriptionRequired    = 14001
	ErrLogicDescriptionEmpty       = 14002
	ErrLogicInvalidJSON            = 14003
	ErrLogicSchemaViolation        = 14004
	ErrLogicTargetRequired         = 14005
	ErrLogicTargetNotAllowed       = 14006
	ErrLogicTargetNoLeadUnderscore = 14007
	ErrLogicTypeRequired           = 14008
	ErrLogicDestroyEventRequired   = 14009
	ErrLogicDestroyEventNotAllowed = 14010
	ErrLogicAssocClassNotAllowed   = 14011
	ErrLogicAssocClassSpecRequired = 14014

	// Parameter errors (15xxx).
	ErrParamNameRequired    = 15001
	ErrParamNameEmpty       = 15002
	ErrParamInvalidJSON     = 15003
	ErrParamSchemaViolation = 15004

	// Global function errors (16xxx).
	ErrGlobalFuncNameRequired     = 16001
	ErrGlobalFuncNameEmpty        = 16002
	ErrGlobalFuncInvalidJSON      = 16003
	ErrGlobalFuncSchemaViolation  = 16004
	ErrGlobalFuncNameNoUnderscore = 16005
	ErrGlobalFuncParamEmpty       = 16006
	ErrGlobalFuncLogicRequired    = 16007

	// Domain association errors (17xxx).
	ErrDomainAssocProblemKeyRequired  = 17001
	ErrDomainAssocProblemKeyEmpty     = 17002
	ErrDomainAssocSolutionKeyRequired = 17003
	ErrDomainAssocSolutionKeyEmpty    = 17004
	ErrDomainAssocInvalidJSON         = 17005
	ErrDomainAssocSchemaViolation     = 17006

	// Subdomain association errors (171xx).
	ErrSubdomainAssocProblemKeyRequired  = 17101
	ErrSubdomainAssocProblemKeyEmpty     = 17102
	ErrSubdomainAssocSolutionKeyRequired = 17103
	ErrSubdomainAssocSolutionKeyEmpty    = 17104
	ErrSubdomainAssocInvalidJSON         = 17105
	ErrSubdomainAssocSchemaViolation     = 17106
	ErrSubdomainAssocSameSubdomains      = 17107

	// Key format errors (11026+) - filesystem wire-format keys must be well-formed snake_case.
	ErrKeyInvalidFormat              = 11026 // Key has invalid format (must be lowercase snake_case)
	ErrAssocFilenameInvalidFormat    = 11027 // Association filename has invalid format
	ErrAssocFilenameInvalidComponent = 11028 // Association filename has invalid component (must be snake_case)

	// Use case errors (18xxx).
	ErrUseCaseNameRequired    = 18001
	ErrUseCaseNameEmpty       = 18002
	ErrUseCaseInvalidJSON     = 18003
	ErrUseCaseSchemaViolation = 18004
	ErrUseCaseLevelRequired   = 18005
	ErrUseCaseLevelInvalid    = 18006

	// Scenario errors (19xxx).
	ErrScenarioNameRequired    = 19001
	ErrScenarioNameEmpty       = 19002
	ErrScenarioInvalidJSON     = 19003
	ErrScenarioSchemaViolation = 19004

	// Use case shared errors (20xxx).
	ErrUseCaseSharedShareTypeRequired = 20001
	ErrUseCaseSharedShareTypeEmpty    = 20002
	ErrUseCaseSharedInvalidJSON       = 20003
	ErrUseCaseSharedSchemaViolation   = 20004

	// Named set errors (22xxx).
	ErrNamedSetNameRequired     = 22001
	ErrNamedSetNameEmpty        = 22002
	ErrNamedSetInvalidJSON      = 22003
	ErrNamedSetSchemaViolation  = 22004
	ErrNamedSetNameNoUnderscore = 22005

	// Conversion errors (21xxx) - errors during inputModel to/from req_model conversion.
	ErrConvKeyConstruction       = 21001 // Identity key construction failed during conversion
	ErrConvModelValidation       = 21002 // Model validation failed after conversion (catch-all for unmapped core errors)
	ErrConvMultiplicityInvalid   = 21003 // Multiplicity parsing failed during conversion
	ErrConvClassNotFound         = 21004 // Class key not found during association conversion
	ErrConvAssocKeyConstruction  = 21005 // Class association key construction failed during conversion
	ErrConvScopedKeyInvalid      = 21006 // Scoped key format invalid during conversion (domain/subdomain/class)
	ErrConvObjectResolveFailed   = 21007 // Failed to resolve object reference during scenario conversion
	ErrConvSourceModelValidation = 21008 // Source model validation failed before ConvertFromModel

	// Mapped core validation errors (21100-21199) - specific core errors mapped to parser_ai errors.
	ErrConvParamDatatypeRequired         = 21100 // Parameter data_type_rules is empty
	ErrConvParamNameRequired             = 21101 // Parameter name is empty
	ErrConvLogicTypeInvalid              = 21102 // Logic has wrong type for its context (e.g., invariant must be assessment/let)
	ErrConvLogicDuplicateLet             = 21103 // Duplicate let target in logic list
	ErrConvLogicDuplicateTarget          = 21104 // Duplicate guarantee target in action/query
	ErrConvLogicTargetRequired           = 21105 // Logic target (attribute) is required but empty
	ErrConvLogicTargetNotAllowed         = 21106 // Logic target must be empty for this logic type
	ErrConvLogicTargetNoUnderscore       = 21107 // Logic target must not start with underscore
	ErrConvReferenceNotFound             = 21108 // Cross-reference to another entity not found
	ErrConvGenCardinalityInvalid         = 21109 // Generalization has wrong number of super/subclasses
	ErrConvDomainStructureInvalid        = 21110 // Domain structural rule violated (subdomain naming, orphan associations)
	ErrConvScenarioStepInvalid           = 21111 // Scenario step structural rule violated
	ErrConvGuaranteeInvalidTarget        = 21112 // Guarantee targets an attribute that doesn't exist
	ErrConvAssocClassSameAsEndpoint      = 21113 // Association class cannot be same as from or to class
	ErrConvInternalKeyError              = 21114 // Internal key validation error (should not normally occur)
	ErrConvUseCaseActorNotActorClass     = 21115 // Use case references class that is not an actor class
	ErrConvLogicSpecInvalid              = 21116 // Logic specification (TLA+ expression) failed validation
	ErrConvDomainAssocSameDomains        = 21117 // Domain association references same domain for problem and solution
	ErrConvSubdomainAssocSameSubdomains  = 21121 // Subdomain association references same subdomain for problem and solution
	ErrConvDomainSassocSingleSubdomain   = 21122 // Subdomain associations require at least two subdomains
	ErrConvTransitionInitialEventInvalid = 21118 // Initial transition event is not _new
	ErrConvTransitionFinalEventInvalid   = 21119 // Final transition event is not _destroy
	ErrConvAssocUniquenessInvalid        = 21120 // Association uniqueness tuple failed validation during conversion
)
