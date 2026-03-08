package coreerr

// Error codes for core model validation failures.
// Each code uniquely identifies a specific validation check.
// Codes follow the pattern {ENTITY}_{FIELD}_{VIOLATION}.

const (
	// ---------------------------------------------------------------
	// Key errors — identity key validation.

	KeyTypeInvalid          Code = "KEY_TYPE_INVALID"            // KeyType is not in the allowed set.
	KeySubkeyRequired       Code = "KEY_SUBKEY_REQUIRED"         // SubKey is empty.
	KeyParentkeyMustBeBlank Code = "KEY_PARENTKEY_MUST_BE_BLANK" // Root-level key type has a non-blank ParentKey.
	KeyParentkeyRequired    Code = "KEY_PARENTKEY_REQUIRED"      // Non-root key type has a blank ParentKey.
	KeyRootHasParent        Code = "KEY_ROOT_HAS_PARENT"         // Root-level key type was given a parent.
	KeyRootHasParentkey     Code = "KEY_ROOT_HAS_PARENTKEY"      // Root-level key has non-empty ParentKey.
	KeyNoParent             Code = "KEY_NO_PARENT"               // Key type requires a parent but got nil.
	KeyWrongParentType      Code = "KEY_WRONG_PARENT_TYPE"       // Parent key has wrong KeyType.
	KeyParentkeyMismatch    Code = "KEY_PARENTKEY_MISMATCH"      // ParentKey doesn't match expected parent string.
	KeyTypeUnknown          Code = "KEY_TYPE_UNKNOWN"            // Unknown key type in ValidateParent.
	KeyCassocParentUnknown  Code = "KEY_CASSOC_PARENT_UNKNOWN"   // Class association parent type cannot be determined.
	KeyCassocModelHasParent Code = "KEY_CASSOC_MODEL_HAS_PARENT" // Model-level class association should not have a parent.

	// ---------------------------------------------------------------
	// Model errors — top-level model validation.

	ModelKeyRequired           Code = "MODEL_KEY_REQUIRED"            // Model Key is empty.
	ModelNameRequired          Code = "MODEL_NAME_REQUIRED"           // Model Name is empty.
	ModelInvariantTypeInvalid  Code = "MODEL_INVARIANT_TYPE_INVALID"  // Model invariant has wrong logic type.
	ModelInvariantDuplicateLet Code = "MODEL_INVARIANT_DUPLICATE_LET" // Model invariant has duplicate let target.
	ModelGfuncKeyMismatch      Code = "MODEL_GFUNC_KEY_MISMATCH"      // Global function map key != function key.
	ModelNsetKeyMismatch       Code = "MODEL_NSET_KEY_MISMATCH"       // Named set map key != named set key.
	ModelAgenSuperclassCount   Code = "MODEL_AGEN_SUPERCLASS_COUNT"   // Actor generalization doesn't have exactly one superclass.
	ModelAgenSubclassCount     Code = "MODEL_AGEN_SUBCLASS_COUNT"     // Actor generalization doesn't have enough subclasses.
	ModelCassocOrphanParent    Code = "MODEL_CASSOC_ORPHAN_PARENT"    // Model-level class association has a parent key.

	// ---------------------------------------------------------------
	// ExpressionSpec errors.

	ExprspecNotationRequired  Code = "EXPRSPEC_NOTATION_REQUIRED"  // Notation field is empty.
	ExprspecNotationInvalid   Code = "EXPRSPEC_NOTATION_INVALID"   // Notation is not a valid value.
	ExprspecExpressionInvalid Code = "EXPRSPEC_EXPRESSION_INVALID" // Expression failed validation.

	// ---------------------------------------------------------------
	// TypeSpec errors.

	TypespecNotationRequired Code = "TYPESPEC_NOTATION_REQUIRED" // Notation field is empty.
	TypespecNotationInvalid  Code = "TYPESPEC_NOTATION_INVALID"  // Notation is not a valid value.
	TypespecExprtypeInvalid  Code = "TYPESPEC_EXPRTYPE_INVALID"  // ExpressionType failed validation.

	// ---------------------------------------------------------------
	// Logic errors.

	LogicKeyInvalid            Code = "LOGIC_KEY_INVALID"             // Logic key failed validation.
	LogicTypeRequired          Code = "LOGIC_TYPE_REQUIRED"           // Logic Type is empty.
	LogicTypeInvalid           Code = "LOGIC_TYPE_INVALID"            // Logic Type is not a valid value.
	LogicDescRequired          Code = "LOGIC_DESC_REQUIRED"           // Logic Description is empty.
	LogicTargetRequired        Code = "LOGIC_TARGET_REQUIRED"         // Logic Type requires a non-empty Target.
	LogicTargetMustBeEmpty     Code = "LOGIC_TARGET_MUST_BE_EMPTY"    // Logic Type requires an empty Target.
	LogicTargetNoUnderscore    Code = "LOGIC_TARGET_NO_UNDERSCORE"    // Query/let Target cannot start with underscore.
	LogicSpecInvalid           Code = "LOGIC_SPEC_INVALID"            // Logic Spec failed validation.
	LogicTargetTypespecInvalid Code = "LOGIC_TARGET_TYPESPEC_INVALID" // Logic TargetTypeSpec failed validation.

	// ---------------------------------------------------------------
	// GlobalFunction errors.

	GfuncKeyInvalid       Code = "GFUNC_KEY_INVALID"        // GlobalFunction key failed validation.
	GfuncKeyTypeInvalid   Code = "GFUNC_KEY_TYPE_INVALID"   // Key is not KEY_TYPE_GLOBAL_FUNCTION.
	GfuncNameRequired     Code = "GFUNC_NAME_REQUIRED"      // GlobalFunction Name is empty.
	GfuncNameNoUnderscore Code = "GFUNC_NAME_NO_UNDERSCORE" // Name doesn't start with underscore.
	GfuncLogicInvalid     Code = "GFUNC_LOGIC_INVALID"      // Nested logic failed validation.
	GfuncLogicKeyMismatch Code = "GFUNC_LOGIC_KEY_MISMATCH" // Logic key doesn't match function key.
	GfuncLogicTypeInvalid Code = "GFUNC_LOGIC_TYPE_INVALID" // Logic type must be 'value'.

	// ---------------------------------------------------------------
	// NamedSet errors.

	NsetKeyInvalid      Code = "NSET_KEY_INVALID"      // NamedSet key failed validation.
	NsetKeyTypeInvalid  Code = "NSET_KEY_TYPE_INVALID" // Key is not KEY_TYPE_NAMED_SET.
	NsetNameRequired    Code = "NSET_NAME_REQUIRED"    // NamedSet Name is empty.
	NsetSpecInvalid     Code = "NSET_SPEC_INVALID"     // NamedSet Spec failed validation.
	NsetTypespecInvalid Code = "NSET_TYPESPEC_INVALID" // NamedSet TypeSpec failed validation.

	// ---------------------------------------------------------------
	// Parameter errors.

	ParamNameRequired      Code = "PARAM_NAME_REQUIRED"      // Parameter Name is empty.
	ParamDatatypesRequired Code = "PARAM_DATATYPES_REQUIRED" // Parameter DataTypes is empty.

	// ---------------------------------------------------------------
	// Action errors.

	ActionKeyInvalid               Code = "ACTION_KEY_INVALID"                // Action key failed validation.
	ActionKeyTypeInvalid           Code = "ACTION_KEY_TYPE_INVALID"           // Key is not KEY_TYPE_ACTION.
	ActionNameRequired             Code = "ACTION_NAME_REQUIRED"              // Action Name is empty.
	ActionRequiresTypeInvalid      Code = "ACTION_REQUIRES_TYPE_INVALID"      // Requires logic type must be assessment or let.
	ActionRequiresDuplicateLet     Code = "ACTION_REQUIRES_DUPLICATE_LET"     // Duplicate let target in Requires.
	ActionGuaranteeTypeInvalid     Code = "ACTION_GUARANTEE_TYPE_INVALID"     // Guarantee logic type must be state_change or let.
	ActionGuaranteeDuplicateLet    Code = "ACTION_GUARANTEE_DUPLICATE_LET"    // Duplicate let target in Guarantees.
	ActionGuaranteeDuplicateTarget Code = "ACTION_GUARANTEE_DUPLICATE_TARGET" // Duplicate target in Guarantees.
	ActionSafetyTypeInvalid        Code = "ACTION_SAFETY_TYPE_INVALID"        // SafetyRule logic type must be safety_rule or let.
	ActionSafetyDuplicateLet       Code = "ACTION_SAFETY_DUPLICATE_LET"       // Duplicate let target in SafetyRules.

	// ---------------------------------------------------------------
	// Guard errors.

	GuardKeyInvalid       Code = "GUARD_KEY_INVALID"        // Guard key failed validation.
	GuardKeyTypeInvalid   Code = "GUARD_KEY_TYPE_INVALID"   // Key is not KEY_TYPE_GUARD.
	GuardNameRequired     Code = "GUARD_NAME_REQUIRED"      // Guard Name is empty.
	GuardLogicInvalid     Code = "GUARD_LOGIC_INVALID"      // Guard Logic failed validation.
	GuardLogicKeyMismatch Code = "GUARD_LOGIC_KEY_MISMATCH" // Logic key doesn't match guard key.
	GuardLogicTypeInvalid Code = "GUARD_LOGIC_TYPE_INVALID" // Logic type must be 'assessment'.

	// ---------------------------------------------------------------
	// Event errors.

	EventKeyInvalid     Code = "EVENT_KEY_INVALID"      // Event key failed validation.
	EventKeyTypeInvalid Code = "EVENT_KEY_TYPE_INVALID" // Key is not KEY_TYPE_EVENT.
	EventNameRequired   Code = "EVENT_NAME_REQUIRED"    // Event Name is empty.

	// ---------------------------------------------------------------
	// Query errors.

	QueryKeyInvalid               Code = "QUERY_KEY_INVALID"                // Query key failed validation.
	QueryKeyTypeInvalid           Code = "QUERY_KEY_TYPE_INVALID"           // Key is not KEY_TYPE_QUERY.
	QueryNameRequired             Code = "QUERY_NAME_REQUIRED"              // Query Name is empty.
	QueryRequiresTypeInvalid      Code = "QUERY_REQUIRES_TYPE_INVALID"      // Requires logic type must be assessment or let.
	QueryRequiresDuplicateLet     Code = "QUERY_REQUIRES_DUPLICATE_LET"     // Duplicate let target in Requires.
	QueryGuaranteeTypeInvalid     Code = "QUERY_GUARANTEE_TYPE_INVALID"     // Guarantee logic type must be query or let.
	QueryGuaranteeDuplicateLet    Code = "QUERY_GUARANTEE_DUPLICATE_LET"    // Duplicate let target in Guarantees.
	QueryGuaranteeDuplicateTarget Code = "QUERY_GUARANTEE_DUPLICATE_TARGET" // Duplicate target in Guarantees.

	// ---------------------------------------------------------------
	// State errors.

	StateKeyInvalid     Code = "STATE_KEY_INVALID"      // State key failed validation.
	StateKeyTypeInvalid Code = "STATE_KEY_TYPE_INVALID" // Key is not KEY_TYPE_STATE.
	StateNameRequired   Code = "STATE_NAME_REQUIRED"    // State Name is empty.

	// ---------------------------------------------------------------
	// StateAction errors (action references within states).

	StateactionKeyInvalid       Code = "STATEACTION_KEY_INVALID"       // StateAction key failed validation.
	StateactionKeyTypeInvalid   Code = "STATEACTION_KEY_TYPE_INVALID"  // Key is not KEY_TYPE_STATE_ACTION.
	StateactionWhenRequired     Code = "STATEACTION_WHEN_REQUIRED"     // When field is empty.
	StateactionWhenInvalid      Code = "STATEACTION_WHEN_INVALID"      // When field is not a valid value.
	StateactionActionkeyInvalid Code = "STATEACTION_ACTIONKEY_INVALID" // ActionKey failed validation.
	StateactionActionkeyType    Code = "STATEACTION_ACTIONKEY_TYPE"    // ActionKey is not KEY_TYPE_ACTION.
	StateactionActionNotfound   Code = "STATEACTION_ACTION_NOTFOUND"   // ActionKey references non-existent action.

	// ---------------------------------------------------------------
	// Transition errors.

	TransitionKeyInvalid          Code = "TRANSITION_KEY_INVALID"          // Transition key failed validation.
	TransitionKeyTypeInvalid      Code = "TRANSITION_KEY_TYPE_INVALID"     // Key is not KEY_TYPE_TRANSITION.
	TransitionNoState             Code = "TRANSITION_NO_STATE"             // Neither FromStateKey nor ToStateKey set.
	TransitionFromstatekeyInvalid Code = "TRANSITION_FROMSTATEKEY_INVALID" // FromStateKey failed validation.
	TransitionFromstatekeyType    Code = "TRANSITION_FROMSTATEKEY_TYPE"    // FromStateKey is not KEY_TYPE_STATE.
	TransitionTostatekeyInvalid   Code = "TRANSITION_TOSTATEKEY_INVALID"   // ToStateKey failed validation.
	TransitionTostatekeyType      Code = "TRANSITION_TOSTATEKEY_TYPE"      // ToStateKey is not KEY_TYPE_STATE.
	TransitionEventkeyInvalid     Code = "TRANSITION_EVENTKEY_INVALID"     // EventKey failed validation.
	TransitionEventkeyType        Code = "TRANSITION_EVENTKEY_TYPE"        // EventKey is not KEY_TYPE_EVENT.
	TransitionGuardkeyInvalid     Code = "TRANSITION_GUARDKEY_INVALID"     // GuardKey failed validation.
	TransitionGuardkeyType        Code = "TRANSITION_GUARDKEY_TYPE"        // GuardKey is not KEY_TYPE_GUARD.
	TransitionActionkeyInvalid    Code = "TRANSITION_ACTIONKEY_INVALID"    // ActionKey failed validation.
	TransitionActionkeyType       Code = "TRANSITION_ACTIONKEY_TYPE"       // ActionKey is not KEY_TYPE_ACTION.
	TransitionFromstateNotfound   Code = "TRANSITION_FROMSTATE_NOTFOUND"   // FromStateKey references non-existent state.
	TransitionTostateNotfound     Code = "TRANSITION_TOSTATE_NOTFOUND"     // ToStateKey references non-existent state.
	TransitionEventNotfound       Code = "TRANSITION_EVENT_NOTFOUND"       // EventKey references non-existent event.
	TransitionGuardNotfound       Code = "TRANSITION_GUARD_NOTFOUND"       // GuardKey references non-existent guard.
	TransitionActionNotfound      Code = "TRANSITION_ACTION_NOTFOUND"      // ActionKey references non-existent action.

	// ---------------------------------------------------------------
	// Class errors.

	ClassKeyInvalid             Code = "CLASS_KEY_INVALID"              // Class key failed validation.
	ClassKeyTypeInvalid         Code = "CLASS_KEY_TYPE_INVALID"         // Key is not KEY_TYPE_CLASS.
	ClassNameRequired           Code = "CLASS_NAME_REQUIRED"            // Class Name is empty.
	ClassActorkeyInvalid        Code = "CLASS_ACTORKEY_INVALID"         // ActorKey failed validation.
	ClassActorkeyTypeInvalid    Code = "CLASS_ACTORKEY_TYPE_INVALID"    // ActorKey is not KEY_TYPE_ACTOR.
	ClassSuperkeyInvalid        Code = "CLASS_SUPERKEY_INVALID"         // SuperclassOfKey failed validation.
	ClassSuperkeyTypeInvalid    Code = "CLASS_SUPERKEY_TYPE_INVALID"    // SuperclassOfKey is not KEY_TYPE_CLASS_GENERALIZATION.
	ClassSubkeyInvalid          Code = "CLASS_SUBKEY_INVALID"           // SubclassOfKey failed validation.
	ClassSubkeyTypeInvalid      Code = "CLASS_SUBKEY_TYPE_INVALID"      // SubclassOfKey is not KEY_TYPE_CLASS_GENERALIZATION.
	ClassSuperSubSame           Code = "CLASS_SUPER_SUB_SAME"           // SuperclassOfKey and SubclassOfKey are the same.
	ClassActorNotfound          Code = "CLASS_ACTOR_NOTFOUND"           // ActorKey references non-existent actor.
	ClassSupergenNotfound       Code = "CLASS_SUPERGEN_NOTFOUND"        // SuperclassOfKey references non-existent generalization.
	ClassSupergenWrongSubdomain Code = "CLASS_SUPERGEN_WRONG_SUBDOMAIN" // SuperclassOfKey generalization not in same subdomain.
	ClassSubgenNotfound         Code = "CLASS_SUBGEN_NOTFOUND"          // SubclassOfKey references non-existent generalization.
	ClassSubgenWrongSubdomain   Code = "CLASS_SUBGEN_WRONG_SUBDOMAIN"   // SubclassOfKey generalization not in same subdomain.
	ClassInvariantTypeInvalid   Code = "CLASS_INVARIANT_TYPE_INVALID"   // Class invariant has wrong logic type.
	ClassInvariantDuplicateLet  Code = "CLASS_INVARIANT_DUPLICATE_LET"  // Class invariant has duplicate let target.
	ClassGuaranteeInvalidTarget Code = "CLASS_GUARANTEE_INVALID_TARGET" // Guarantee targets non-existent attribute.

	// ---------------------------------------------------------------
	// Attribute errors.

	AttrKeyInvalid            Code = "ATTR_KEY_INVALID"             // Attribute key failed validation.
	AttrKeyTypeInvalid        Code = "ATTR_KEY_TYPE_INVALID"        // Key is not KEY_TYPE_ATTRIBUTE.
	AttrNameRequired          Code = "ATTR_NAME_REQUIRED"           // Attribute Name is empty.
	AttrDerivationTypeInvalid Code = "ATTR_DERIVATION_TYPE_INVALID" // DerivationPolicy logic type is invalid.
	AttrInvariantTypeInvalid  Code = "ATTR_INVARIANT_TYPE_INVALID"  // Attribute invariant has wrong logic type.
	AttrInvariantDuplicateLet Code = "ATTR_INVARIANT_DUPLICATE_LET" // Attribute invariant has duplicate let target.

	// ---------------------------------------------------------------
	// Association errors.

	AssocKeyInvalid         Code = "ASSOC_KEY_INVALID"          // Association key failed validation.
	AssocKeyTypeInvalid     Code = "ASSOC_KEY_TYPE_INVALID"     // Key is not KEY_TYPE_CLASS_ASSOCIATION.
	AssocNameRequired       Code = "ASSOC_NAME_REQUIRED"        // Association Name is empty.
	AssocFromkeyInvalid     Code = "ASSOC_FROMKEY_INVALID"      // FromClassKey failed validation.
	AssocFromkeyTypeInvalid Code = "ASSOC_FROMKEY_TYPE_INVALID" // FromClassKey is not KEY_TYPE_CLASS.
	AssocTokeyInvalid       Code = "ASSOC_TOKEY_INVALID"        // ToClassKey failed validation.
	AssocTokeyTypeInvalid   Code = "ASSOC_TOKEY_TYPE_INVALID"   // ToClassKey is not KEY_TYPE_CLASS.
	AssocFromMultInvalid    Code = "ASSOC_FROM_MULT_INVALID"    // FromMultiplicity failed validation.
	AssocToMultInvalid      Code = "ASSOC_TO_MULT_INVALID"      // ToMultiplicity failed validation.
	AssocAssocclassInvalid  Code = "ASSOC_ASSOCCLASS_INVALID"   // AssociationClassKey failed validation.
	AssocAssocclassType     Code = "ASSOC_ASSOCCLASS_TYPE"      // AssociationClassKey is not KEY_TYPE_CLASS.
	AssocAssocclassSameFrom Code = "ASSOC_ASSOCCLASS_SAME_FROM" // AssociationClassKey same as FromClassKey.
	AssocAssocclassSameTo   Code = "ASSOC_ASSOCCLASS_SAME_TO"   // AssociationClassKey same as ToClassKey.
	AssocFromNotfound       Code = "ASSOC_FROM_NOTFOUND"        // FromClassKey references non-existent class.
	AssocToNotfound         Code = "ASSOC_TO_NOTFOUND"          // ToClassKey references non-existent class.
	AssocAssocclassNotfound Code = "ASSOC_ASSOCCLASS_NOTFOUND"  // AssociationClassKey references non-existent class.

	// ---------------------------------------------------------------
	// ClassGeneralization errors.

	CgenKeyInvalid     Code = "CGEN_KEY_INVALID"      // ClassGeneralization key failed validation.
	CgenKeyTypeInvalid Code = "CGEN_KEY_TYPE_INVALID" // Key is not KEY_TYPE_CLASS_GENERALIZATION.
	CgenNameRequired   Code = "CGEN_NAME_REQUIRED"    // ClassGeneralization Name is empty.

	// ---------------------------------------------------------------
	// Actor errors.

	ActorKeyInvalid          Code = "ACTOR_KEY_INVALID"           // Actor key failed validation.
	ActorKeyTypeInvalid      Code = "ACTOR_KEY_TYPE_INVALID"      // Key is not KEY_TYPE_ACTOR.
	ActorNameRequired        Code = "ACTOR_NAME_REQUIRED"         // Actor Name is empty.
	ActorTypeRequired        Code = "ACTOR_TYPE_REQUIRED"         // Actor Type is empty.
	ActorTypeInvalid         Code = "ACTOR_TYPE_INVALID"          // Actor Type is not a valid value.
	ActorSuperkeyInvalid     Code = "ACTOR_SUPERKEY_INVALID"      // SuperclassOfKey failed validation.
	ActorSuperkeyTypeInvalid Code = "ACTOR_SUPERKEY_TYPE_INVALID" // SuperclassOfKey is not KEY_TYPE_ACTOR_GENERALIZATION.
	ActorSubkeyInvalid       Code = "ACTOR_SUBKEY_INVALID"        // SubclassOfKey failed validation.
	ActorSubkeyTypeInvalid   Code = "ACTOR_SUBKEY_TYPE_INVALID"   // SubclassOfKey is not KEY_TYPE_ACTOR_GENERALIZATION.
	ActorSuperSubSame        Code = "ACTOR_SUPER_SUB_SAME"        // SuperclassOfKey and SubclassOfKey are the same.
	ActorSupergenNotfound    Code = "ACTOR_SUPERGEN_NOTFOUND"     // SuperclassOfKey references non-existent generalization.
	ActorSubgenNotfound      Code = "ACTOR_SUBGEN_NOTFOUND"       // SubclassOfKey references non-existent generalization.

	// ---------------------------------------------------------------
	// ActorGeneralization errors.

	AgenKeyInvalid     Code = "AGEN_KEY_INVALID"      // ActorGeneralization key failed validation.
	AgenKeyTypeInvalid Code = "AGEN_KEY_TYPE_INVALID" // Key is not KEY_TYPE_ACTOR_GENERALIZATION.
	AgenNameRequired   Code = "AGEN_NAME_REQUIRED"    // ActorGeneralization Name is empty.

	// ---------------------------------------------------------------
	// Scenario errors.

	ScenarioKeyInvalid     Code = "SCENARIO_KEY_INVALID"      // Scenario key failed validation.
	ScenarioKeyTypeInvalid Code = "SCENARIO_KEY_TYPE_INVALID" // Key is not KEY_TYPE_SCENARIO.
	ScenarioNameRequired   Code = "SCENARIO_NAME_REQUIRED"    // Scenario Name is empty.

	// ---------------------------------------------------------------
	// ScenarioObject errors.

	SobjectKeyInvalid          Code = "SOBJECT_KEY_INVALID"           // ScenarioObject key failed validation.
	SobjectKeyTypeInvalid      Code = "SOBJECT_KEY_TYPE_INVALID"      // Key is not KEY_TYPE_SCENARIO_OBJECT.
	SobjectNamestyleRequired   Code = "SOBJECT_NAMESTYLE_REQUIRED"    // NameStyle is empty.
	SobjectNamestyleInvalid    Code = "SOBJECT_NAMESTYLE_INVALID"     // NameStyle is not a valid value.
	SobjectNameRequired        Code = "SOBJECT_NAME_REQUIRED"         // Name is required when NameStyle is "named".
	SobjectNameMustBeBlank     Code = "SOBJECT_NAME_MUST_BE_BLANK"    // Name must be blank when NameStyle is "class_name".
	SobjectClasskeyInvalid     Code = "SOBJECT_CLASSKEY_INVALID"      // ClassKey failed validation.
	SobjectClasskeyTypeInvalid Code = "SOBJECT_CLASSKEY_TYPE_INVALID" // ClassKey is not KEY_TYPE_CLASS.
	SobjectClassNotfound       Code = "SOBJECT_CLASS_NOTFOUND"        // ClassKey references non-existent class.

	// ---------------------------------------------------------------
	// ScenarioStep errors.

	SstepKeyInvalid             Code = "SSTEP_KEY_INVALID"              // ScenarioStep key failed validation.
	SstepKeyTypeInvalid         Code = "SSTEP_KEY_TYPE_INVALID"         // Key is not KEY_TYPE_SCENARIO_STEP.
	SstepTypeUnknown            Code = "SSTEP_TYPE_UNKNOWN"             // StepType is not a valid value.
	SstepLeafTypeRequired       Code = "SSTEP_LEAF_TYPE_REQUIRED"       // Leaf step missing LeafType.
	SstepLeafTypeUnknown        Code = "SSTEP_LEAF_TYPE_UNKNOWN"        // LeafType is not a valid value.
	SstepFromkeyInvalid         Code = "SSTEP_FROMKEY_INVALID"          // FromObjectKey failed validation.
	SstepFromkeyTypeInvalid     Code = "SSTEP_FROMKEY_TYPE_INVALID"     // FromObjectKey is not KEY_TYPE_SCENARIO_OBJECT.
	SstepTokeyInvalid           Code = "SSTEP_TOKEY_INVALID"            // ToObjectKey failed validation.
	SstepTokeyTypeInvalid       Code = "SSTEP_TOKEY_TYPE_INVALID"       // ToObjectKey is not KEY_TYPE_SCENARIO_OBJECT.
	SstepEventkeyInvalid        Code = "SSTEP_EVENTKEY_INVALID"         // EventKey failed validation.
	SstepEventkeyTypeInvalid    Code = "SSTEP_EVENTKEY_TYPE_INVALID"    // EventKey is not KEY_TYPE_EVENT.
	SstepQuerykeyInvalid        Code = "SSTEP_QUERYKEY_INVALID"         // QueryKey failed validation.
	SstepQuerykeyTypeInvalid    Code = "SSTEP_QUERYKEY_TYPE_INVALID"    // QueryKey is not KEY_TYPE_QUERY.
	SstepScenariokeyInvalid     Code = "SSTEP_SCENARIOKEY_INVALID"      // ScenarioKey failed validation.
	SstepScenariokeyTypeInvalid Code = "SSTEP_SCENARIOKEY_TYPE_INVALID" // ScenarioKey is not KEY_TYPE_SCENARIO.
	SstepEventFromRequired      Code = "SSTEP_EVENT_FROM_REQUIRED"      // Event leaf missing FromObjectKey.
	SstepEventToRequired        Code = "SSTEP_EVENT_TO_REQUIRED"        // Event leaf missing ToObjectKey.
	SstepEventKeyRequired       Code = "SSTEP_EVENT_KEY_REQUIRED"       // Event leaf missing EventKey.
	SstepEventQueryForbidden    Code = "SSTEP_EVENT_QUERY_FORBIDDEN"    // Event leaf has QueryKey set.
	SstepQueryFromRequired      Code = "SSTEP_QUERY_FROM_REQUIRED"      // Query leaf missing FromObjectKey.
	SstepQueryToRequired        Code = "SSTEP_QUERY_TO_REQUIRED"        // Query leaf missing ToObjectKey.
	SstepQueryKeyRequired       Code = "SSTEP_QUERY_KEY_REQUIRED"       // Query leaf missing QueryKey.
	SstepQueryEventForbidden    Code = "SSTEP_QUERY_EVENT_FORBIDDEN"    // Query leaf has EventKey set.
	SstepScenarioFromRequired   Code = "SSTEP_SCENARIO_FROM_REQUIRED"   // Scenario leaf missing FromObjectKey.
	SstepScenarioToRequired     Code = "SSTEP_SCENARIO_TO_REQUIRED"     // Scenario leaf missing ToObjectKey.
	SstepScenarioKeyRequired    Code = "SSTEP_SCENARIO_KEY_REQUIRED"    // Scenario leaf missing ScenarioKey.
	SstepScenarioEventForbidden Code = "SSTEP_SCENARIO_EVENT_FORBIDDEN" // Scenario leaf has EventKey set.
	SstepScenarioSelfRef        Code = "SSTEP_SCENARIO_SELF_REF"        // Scenario leaf references its own scenario.
	SstepDeleteFromRequired     Code = "SSTEP_DELETE_FROM_REQUIRED"     // Delete leaf missing FromObjectKey.
	SstepDeleteToForbidden      Code = "SSTEP_DELETE_TO_FORBIDDEN"      // Delete leaf has ToObjectKey set.
	SstepDeleteKeysForbidden    Code = "SSTEP_DELETE_KEYS_FORBIDDEN"    // Delete leaf has EventKey/QueryKey/ScenarioKey set.
	SstepSequenceMinStatements  Code = "SSTEP_SEQUENCE_MIN_STATEMENTS"  // Sequence step needs >=2 statements.
	SstepSwitchMinCases         Code = "SSTEP_SWITCH_MIN_CASES"         // Switch step needs >=1 case.
	SstepSwitchCaseType         Code = "SSTEP_SWITCH_CASE_TYPE"         // Switch case must be a STEP_TYPE_CASE.
	SstepCaseConditionRequired  Code = "SSTEP_CASE_CONDITION_REQUIRED"  // Case step missing Condition.
	SstepLoopConditionRequired  Code = "SSTEP_LOOP_CONDITION_REQUIRED"  // Loop step missing Condition.
	SstepLoopMinStatements      Code = "SSTEP_LOOP_MIN_STATEMENTS"      // Loop step needs >=1 statement.

	// ---------------------------------------------------------------
	// UseCase errors.

	UcKeyInvalid          Code = "UC_KEY_INVALID"           // UseCase key failed validation.
	UcKeyTypeInvalid      Code = "UC_KEY_TYPE_INVALID"      // Key is not KEY_TYPE_USE_CASE.
	UcNameRequired        Code = "UC_NAME_REQUIRED"         // UseCase Name is empty.
	UcLevelRequired       Code = "UC_LEVEL_REQUIRED"        // UseCase Level is empty.
	UcLevelInvalid        Code = "UC_LEVEL_INVALID"         // UseCase Level is not a valid value.
	UcSuperkeyInvalid     Code = "UC_SUPERKEY_INVALID"      // SuperclassOfKey failed validation.
	UcSuperkeyTypeInvalid Code = "UC_SUPERKEY_TYPE_INVALID" // SuperclassOfKey is not KEY_TYPE_USE_CASE_GENERALIZATION.
	UcSubkeyInvalid       Code = "UC_SUBKEY_INVALID"        // SubclassOfKey failed validation.
	UcSubkeyTypeInvalid   Code = "UC_SUBKEY_TYPE_INVALID"   // SubclassOfKey is not KEY_TYPE_USE_CASE_GENERALIZATION.
	UcSuperSubSame        Code = "UC_SUPER_SUB_SAME"        // SuperclassOfKey and SubclassOfKey are the same.
	UcActorNotActorClass  Code = "UC_ACTOR_NOT_ACTOR_CLASS" // UseCase's class is not an actor class.
	UcSupergenNotfound    Code = "UC_SUPERGEN_NOTFOUND"     // SuperclassOfKey references non-existent generalization.
	UcSubgenNotfound      Code = "UC_SUBGEN_NOTFOUND"       // SubclassOfKey references non-existent generalization.

	// ---------------------------------------------------------------
	// UseCaseGeneralization errors.

	UcgenKeyInvalid     Code = "UCGEN_KEY_INVALID"      // UseCaseGeneralization key failed validation.
	UcgenKeyTypeInvalid Code = "UCGEN_KEY_TYPE_INVALID" // Key is not KEY_TYPE_USE_CASE_GENERALIZATION.
	UcgenNameRequired   Code = "UCGEN_NAME_REQUIRED"    // UseCaseGeneralization Name is empty.

	// ---------------------------------------------------------------
	// UseCaseShared errors.

	UshareSharetypeInvalid Code = "USHARE_SHARETYPE_INVALID" // ShareType is not a valid value.

	// ---------------------------------------------------------------
	// Domain errors.

	DomainKeyInvalid            Code = "DOMAIN_KEY_INVALID"             // Domain key failed validation.
	DomainKeyTypeInvalid        Code = "DOMAIN_KEY_TYPE_INVALID"        // Key is not KEY_TYPE_DOMAIN.
	DomainNameRequired          Code = "DOMAIN_NAME_REQUIRED"           // Domain Name is empty.
	DomainSubdomainSingleKey    Code = "DOMAIN_SUBDOMAIN_SINGLE_KEY"    // Single subdomain not named "default".
	DomainSubdomainMultiDefault Code = "DOMAIN_SUBDOMAIN_MULTI_DEFAULT" // Multiple subdomains include one named "default".
	DomainCassocOrphan          Code = "DOMAIN_CASSOC_ORPHAN"           // Domain class association not connected to any subdomain class.

	// ---------------------------------------------------------------
	// DomainAssociation errors.

	DassocKeyInvalid         Code = "DASSOC_KEY_INVALID"         // DomainAssociation key failed validation.
	DassocKeyTypeInvalid     Code = "DASSOC_KEY_TYPE_INVALID"    // Key is not KEY_TYPE_DOMAIN_ASSOCIATION.
	DassocProblemkeyInvalid  Code = "DASSOC_PROBLEMKEY_INVALID"  // ProblemDomainKey failed validation.
	DassocProblemkeyType     Code = "DASSOC_PROBLEMKEY_TYPE"     // ProblemDomainKey is not KEY_TYPE_DOMAIN.
	DassocSolutionkeyInvalid Code = "DASSOC_SOLUTIONKEY_INVALID" // SolutionDomainKey failed validation.
	DassocSolutionkeyType    Code = "DASSOC_SOLUTIONKEY_TYPE"    // SolutionDomainKey is not KEY_TYPE_DOMAIN.
	DassocSameDomains        Code = "DASSOC_SAME_DOMAINS"        // ProblemDomainKey and SolutionDomainKey are the same.
	DassocProblemNotfound    Code = "DASSOC_PROBLEM_NOTFOUND"    // ProblemDomainKey references non-existent domain.
	DassocSolutionNotfound   Code = "DASSOC_SOLUTION_NOTFOUND"   // SolutionDomainKey references non-existent domain.

	// ---------------------------------------------------------------
	// Subdomain errors.

	SubdomainKeyInvalid             Code = "SUBDOMAIN_KEY_INVALID"              // Subdomain key failed validation.
	SubdomainKeyTypeInvalid         Code = "SUBDOMAIN_KEY_TYPE_INVALID"         // Key is not KEY_TYPE_SUBDOMAIN.
	SubdomainNameRequired           Code = "SUBDOMAIN_NAME_REQUIRED"            // Subdomain Name is empty.
	SubdomainCassocNoParent         Code = "SUBDOMAIN_CASSOC_NO_PARENT"         // Subdomain class association is not parented.
	SubdomainCassocWrongParent      Code = "SUBDOMAIN_CASSOC_WRONG_PARENT"      // Subdomain class association has wrong parent.
	SubdomainCgenSuperclassCount    Code = "SUBDOMAIN_CGEN_SUPERCLASS_COUNT"    // ClassGeneralization doesn't have exactly one superclass.
	SubdomainCgenSubclassCount      Code = "SUBDOMAIN_CGEN_SUBCLASS_COUNT"      // ClassGeneralization doesn't have enough subclasses.
	SubdomainUcgenSuperclassCount   Code = "SUBDOMAIN_UCGEN_SUPERCLASS_COUNT"   // UCGeneralization doesn't have exactly one superclass.
	SubdomainUcgenSubclassCount     Code = "SUBDOMAIN_UCGEN_SUBCLASS_COUNT"     // UCGeneralization doesn't have enough subclasses.
	SubdomainUshareSealevelNotfound Code = "SUBDOMAIN_USHARE_SEALEVEL_NOTFOUND" // UseCaseShared references non-existent sea-level use case.
	SubdomainUshareMudlevelNotfound Code = "SUBDOMAIN_USHARE_MUDLEVEL_NOTFOUND" // UseCaseShared references non-existent mud-level use case.

	// ---------------------------------------------------------------
	// Expression errors — model expression node validation.

	ExprIntValueRequired      Code = "EXPR_INT_VALUE_REQUIRED"      // IntLiteral Value is nil.
	ExprRatValueRequired      Code = "EXPR_RAT_VALUE_REQUIRED"      // RationalLiteral Value is nil.
	ExprSetElemRequired       Code = "EXPR_SET_ELEM_REQUIRED"       // SetLiteral element is nil.
	ExprSetElemInvalid        Code = "EXPR_SET_ELEM_INVALID"        // SetLiteral element failed validation.
	ExprTupleElemRequired     Code = "EXPR_TUPLE_ELEM_REQUIRED"     // TupleLiteral Elements is empty.
	ExprTupleElemNil          Code = "EXPR_TUPLE_ELEM_NIL"          // TupleLiteral element is nil.
	ExprTupleElemInvalid      Code = "EXPR_TUPLE_ELEM_INVALID"      // TupleLiteral element failed validation.
	ExprRecordFieldRequired   Code = "EXPR_RECORD_FIELD_REQUIRED"   // RecordLiteral Fields is empty.
	ExprRecordNameRequired    Code = "EXPR_RECORD_NAME_REQUIRED"    // RecordLiteral field Name is empty.
	ExprRecordValueRequired   Code = "EXPR_RECORD_VALUE_REQUIRED"   // RecordLiteral field Value is nil.
	ExprRecordValueInvalid    Code = "EXPR_RECORD_VALUE_INVALID"    // RecordLiteral field Value failed validation.
	ExprSetconstKindRequired  Code = "EXPR_SETCONST_KIND_REQUIRED"  // SetConstant Kind is empty.
	ExprSetconstKindInvalid   Code = "EXPR_SETCONST_KIND_INVALID"   // SetConstant Kind is not a valid value.
	ExprAttrkeyInvalid        Code = "EXPR_ATTRKEY_INVALID"         // AttributeRef AttributeKey failed validation.
	ExprLocalvarNameRequired  Code = "EXPR_LOCALVAR_NAME_REQUIRED"  // LocalVar Name is empty.
	ExprPriorfieldRequired    Code = "EXPR_PRIORFIELD_REQUIRED"     // PriorFieldValue Field is empty.
	ExprNextstateExprRequired Code = "EXPR_NEXTSTATE_EXPR_REQUIRED" // NextState Expr is nil.
	ExprOpRequired            Code = "EXPR_OP_REQUIRED"             // Binary operator Op is empty.
	ExprOpInvalid             Code = "EXPR_OP_INVALID"              // Binary operator Op is not a valid value.
	ExprLeftRequired          Code = "EXPR_LEFT_REQUIRED"           // Binary operator Left is nil.
	ExprLeftInvalid           Code = "EXPR_LEFT_INVALID"            // Binary operator Left failed validation.
	ExprRightRequired         Code = "EXPR_RIGHT_REQUIRED"          // Binary operator Right is nil.
	ExprRightInvalid          Code = "EXPR_RIGHT_INVALID"           // Binary operator Right failed validation.
	ExprElementRequired       Code = "EXPR_ELEMENT_REQUIRED"        // Membership Element is nil.
	ExprElementInvalid        Code = "EXPR_ELEMENT_INVALID"         // Membership Element failed validation.
	ExprSetRequired           Code = "EXPR_SET_REQUIRED"            // Membership/SetFilter Set is nil.
	ExprSetInvalid            Code = "EXPR_SET_INVALID"             // Membership/SetFilter Set failed validation.
	ExprExprRequired          Code = "EXPR_EXPR_REQUIRED"           // Negate/Not Expr is nil.
	ExprFieldRequired         Code = "EXPR_FIELD_REQUIRED"          // FieldAccess Field is empty.
	ExprBaseRequired          Code = "EXPR_BASE_REQUIRED"           // FieldAccess/RecordUpdate Base is nil.
	ExprBaseInvalid           Code = "EXPR_BASE_INVALID"            // FieldAccess/RecordUpdate Base failed validation.
	ExprTupleRequired         Code = "EXPR_TUPLE_REQUIRED"          // TupleIndex Tuple is nil.
	ExprTupleInvalid          Code = "EXPR_TUPLE_INVALID"           // TupleIndex Tuple failed validation.
	ExprIndexRequired         Code = "EXPR_INDEX_REQUIRED"          // TupleIndex/StringIndex Index is nil.
	ExprIndexInvalid          Code = "EXPR_INDEX_INVALID"           // TupleIndex/StringIndex Index failed validation.
	ExprAlterationsRequired   Code = "EXPR_ALTERATIONS_REQUIRED"    // RecordUpdate Alterations is empty.
	ExprAltFieldRequired      Code = "EXPR_ALT_FIELD_REQUIRED"      // RecordUpdate alteration Field is empty.
	ExprAltValueRequired      Code = "EXPR_ALT_VALUE_REQUIRED"      // RecordUpdate alteration Value is nil.
	ExprAltValueInvalid       Code = "EXPR_ALT_VALUE_INVALID"       // RecordUpdate alteration Value failed validation.
	ExprStrRequired           Code = "EXPR_STR_REQUIRED"            // StringIndex Str is nil.
	ExprStrInvalid            Code = "EXPR_STR_INVALID"             // StringIndex Str failed validation.
	ExprOperandsMinTwo        Code = "EXPR_OPERANDS_MIN_TWO"        // StringConcat/TupleConcat Operands has fewer than 2 elements.
	ExprOperandRequired       Code = "EXPR_OPERAND_REQUIRED"        // StringConcat/TupleConcat operand is nil.
	ExprOperandInvalid        Code = "EXPR_OPERAND_INVALID"         // StringConcat/TupleConcat operand failed validation.
	ExprConditionRequired     Code = "EXPR_CONDITION_REQUIRED"      // IfThenElse Condition is nil.
	ExprConditionInvalid      Code = "EXPR_CONDITION_INVALID"       // IfThenElse Condition failed validation.
	ExprThenRequired          Code = "EXPR_THEN_REQUIRED"           // IfThenElse Then is nil.
	ExprThenInvalid           Code = "EXPR_THEN_INVALID"            // IfThenElse Then failed validation.
	ExprElseRequired          Code = "EXPR_ELSE_REQUIRED"           // IfThenElse Else is nil.
	ExprElseInvalid           Code = "EXPR_ELSE_INVALID"            // IfThenElse Else failed validation.
	ExprBranchesRequired      Code = "EXPR_BRANCHES_REQUIRED"       // Case Branches is empty.
	ExprBranchCondRequired    Code = "EXPR_BRANCH_COND_REQUIRED"    // Case branch Condition is nil.
	ExprBranchCondInvalid     Code = "EXPR_BRANCH_COND_INVALID"     // Case branch Condition failed validation.
	ExprBranchResultRequired  Code = "EXPR_BRANCH_RESULT_REQUIRED"  // Case branch Result is nil.
	ExprBranchResultInvalid   Code = "EXPR_BRANCH_RESULT_INVALID"   // Case branch Result failed validation.
	ExprOtherwiseInvalid      Code = "EXPR_OTHERWISE_INVALID"       // Case Otherwise failed validation.
	ExprQuantKindRequired     Code = "EXPR_QUANT_KIND_REQUIRED"     // Quantifier Kind is empty.
	ExprQuantKindInvalid      Code = "EXPR_QUANT_KIND_INVALID"      // Quantifier Kind is not a valid value.
	ExprVariableRequired      Code = "EXPR_VARIABLE_REQUIRED"       // Quantifier/SetFilter Variable is empty.
	ExprDomainRequired        Code = "EXPR_DOMAIN_REQUIRED"         // Quantifier Domain is nil.
	ExprDomainInvalid         Code = "EXPR_DOMAIN_INVALID"          // Quantifier Domain failed validation.
	ExprPredicateRequired     Code = "EXPR_PREDICATE_REQUIRED"      // Quantifier/SetFilter Predicate is nil.
	ExprPredicateInvalid      Code = "EXPR_PREDICATE_INVALID"       // Quantifier/SetFilter Predicate failed validation.
	ExprStartRequired         Code = "EXPR_START_REQUIRED"          // SetRange Start is nil.
	ExprStartInvalid          Code = "EXPR_START_INVALID"           // SetRange Start failed validation.
	ExprEndRequired           Code = "EXPR_END_REQUIRED"            // SetRange End is nil.
	ExprEndInvalid            Code = "EXPR_END_INVALID"             // SetRange End failed validation.
	ExprActionkeyInvalid      Code = "EXPR_ACTIONKEY_INVALID"       // ActionCall ActionKey failed validation.
	ExprArgRequired           Code = "EXPR_ARG_REQUIRED"            // Call argument is nil.
	ExprArgInvalid            Code = "EXPR_ARG_INVALID"             // Call argument failed validation.
	ExprFunctionkeyInvalid    Code = "EXPR_FUNCTIONKEY_INVALID"     // GlobalCall FunctionKey failed validation.
	ExprModuleRequired        Code = "EXPR_MODULE_REQUIRED"         // BuiltinCall Module is empty.
	ExprFunctionRequired      Code = "EXPR_FUNCTION_REQUIRED"       // BuiltinCall Function is empty.
	ExprSetkeyInvalid         Code = "EXPR_SETKEY_INVALID"          // NamedSetRef SetKey failed validation.

	// ---------------------------------------------------------------
	// DataType errors — data type validation.

	DtypeKeyRequired            Code = "DTYPE_KEY_REQUIRED"            // DataType Key is empty.
	DtypeCollectiontypeRequired Code = "DTYPE_COLLECTIONTYPE_REQUIRED" // DataType CollectionType is empty.
	DtypeCollectiontypeInvalid  Code = "DTYPE_COLLECTIONTYPE_INVALID"  // DataType CollectionType is not a valid value.
	DtypeAtomicRequired         Code = "DTYPE_ATOMIC_REQUIRED"         // DataType Atomic is nil for atomic collection type.
	DtypeRecordfieldsRequired   Code = "DTYPE_RECORDFIELDS_REQUIRED"   // DataType RecordFields is empty for record collection type.
	DtypeColluniqRequired       Code = "DTYPE_COLLUNIQ_REQUIRED"       // DataType CollectionUnique is nil for collection type.
	DtypeColluniqMustBeBlank    Code = "DTYPE_COLLUNIQ_MUST_BE_BLANK"  // DataType CollectionUnique must be nil for non-collection type.
	DtypeCollminMustBeBlank     Code = "DTYPE_COLLMIN_MUST_BE_BLANK"   // DataType CollectionMin must be nil for non-collection type.
	DtypeCollmaxMustBeBlank     Code = "DTYPE_COLLMAX_MUST_BE_BLANK"   // DataType CollectionMax must be nil for non-collection type.
	DtypeCollminTooSmall        Code = "DTYPE_COLLMIN_TOO_SMALL"       // DataType CollectionMin must be >= 1.
	DtypeCollmaxTooSmall        Code = "DTYPE_COLLMAX_TOO_SMALL"       // DataType CollectionMax must be >= 1.
	DtypeCollmaxLessThanMin     Code = "DTYPE_COLLMAX_LESS_THAN_MIN"   // DataType CollectionMax must be >= CollectionMin.

	// ---------------------------------------------------------------
	// Atomic errors — atomic data type validation.

	DtypeAtomicConstrainttypeRequired Code = "DTYPE_ATOMIC_CONSTRAINTTYPE_REQUIRED" // Atomic ConstraintType is empty.
	DtypeAtomicConstrainttypeInvalid  Code = "DTYPE_ATOMIC_CONSTRAINTTYPE_INVALID"  // Atomic ConstraintType is not a valid value.
	DtypeAtomicRefRequired            Code = "DTYPE_ATOMIC_REF_REQUIRED"            // Atomic Reference must not be nil/empty for reference types.
	DtypeAtomicRefMustBeBlank         Code = "DTYPE_ATOMIC_REF_MUST_BE_BLANK"       // Atomic Reference must be nil for non-reference types.
	DtypeAtomicObjkeyRequired         Code = "DTYPE_ATOMIC_OBJKEY_REQUIRED"         // Atomic ObjectClassKey must not be nil/empty for object types.
	DtypeAtomicObjkeyMustBeBlank      Code = "DTYPE_ATOMIC_OBJKEY_MUST_BE_BLANK"    // Atomic ObjectClassKey must be nil for non-object types.
	DtypeAtomicEnumsRequired          Code = "DTYPE_ATOMIC_ENUMS_REQUIRED"          // Atomic Enums cannot be empty for enumeration types.
	DtypeAtomicEnumsMustBeBlank       Code = "DTYPE_ATOMIC_ENUMS_MUST_BE_BLANK"     // Atomic Enums must be empty for non-enumeration types.
	DtypeAtomicEnumordRequired        Code = "DTYPE_ATOMIC_ENUMORD_REQUIRED"        // Atomic EnumOrdered must not be nil for enumeration types.
	DtypeAtomicEnumordMustBeBlank     Code = "DTYPE_ATOMIC_ENUMORD_MUST_BE_BLANK"   // Atomic EnumOrdered must be nil for non-enumeration types.
	DtypeAtomicSpanRequired           Code = "DTYPE_ATOMIC_SPAN_REQUIRED"           // Atomic Span must not be nil for span types.
	DtypeAtomicSpanMustBeBlank        Code = "DTYPE_ATOMIC_SPAN_MUST_BE_BLANK"      // Atomic Span must be nil for non-span types.

	// ---------------------------------------------------------------
	// AtomicSpan errors — atomic span validation.

	DtypeSpanLowertypeRequired   Code = "DTYPE_SPAN_LOWERTYPE_REQUIRED"   // AtomicSpan LowerType is empty.
	DtypeSpanLowertypeInvalid    Code = "DTYPE_SPAN_LOWERTYPE_INVALID"    // AtomicSpan LowerType is not a valid value.
	DtypeSpanHighertypeRequired  Code = "DTYPE_SPAN_HIGHERTYPE_REQUIRED"  // AtomicSpan HigherType is empty.
	DtypeSpanHighertypeInvalid   Code = "DTYPE_SPAN_HIGHERTYPE_INVALID"   // AtomicSpan HigherType is not a valid value.
	DtypeSpanUnitsRequired       Code = "DTYPE_SPAN_UNITS_REQUIRED"       // AtomicSpan Units is empty.
	DtypeSpanPrecisionRequired   Code = "DTYPE_SPAN_PRECISION_REQUIRED"   // AtomicSpan Precision is zero.
	DtypeSpanLowervalRequired    Code = "DTYPE_SPAN_LOWERVAL_REQUIRED"    // AtomicSpan LowerValue must not be nil for constrained lower bound.
	DtypeSpanHighervalRequired   Code = "DTYPE_SPAN_HIGHERVAL_REQUIRED"   // AtomicSpan HigherValue must not be nil for constrained higher bound.
	DtypeSpanLowerdenomRequired  Code = "DTYPE_SPAN_LOWERDENOM_REQUIRED"  // AtomicSpan LowerDenominator must not be nil for constrained lower bound.
	DtypeSpanLowerdenomInvalid   Code = "DTYPE_SPAN_LOWERDENOM_INVALID"   // AtomicSpan LowerDenominator must be >= 1.
	DtypeSpanHigherdenomRequired Code = "DTYPE_SPAN_HIGHERDENOM_REQUIRED" // AtomicSpan HigherDenominator must not be nil for constrained higher bound.
	DtypeSpanHigherdenomInvalid  Code = "DTYPE_SPAN_HIGHERDENOM_INVALID"  // AtomicSpan HigherDenominator must be >= 1.
	DtypeSpanPrecisionInvalid    Code = "DTYPE_SPAN_PRECISION_INVALID"    // AtomicSpan Precision must be > 0 and <= 1.
	DtypeSpanPrecisionNotPow10   Code = "DTYPE_SPAN_PRECISION_NOT_POW10"  // AtomicSpan Precision must be a power of 10.

	// ---------------------------------------------------------------
	// AtomicEnum errors — atomic enum validation.

	DtypeEnumValueRequired Code = "DTYPE_ENUM_VALUE_REQUIRED" // AtomicEnum Value is empty.

	// ---------------------------------------------------------------
	// Field errors — record field validation.

	DtypeFieldNameRequired     Code = "DTYPE_FIELD_NAME_REQUIRED"     // Field Name is empty.
	DtypeFieldDatatypeRequired Code = "DTYPE_FIELD_DATATYPE_REQUIRED" // Field FieldDataType is nil.
	DtypeFieldNameInvalid      Code = "DTYPE_FIELD_NAME_INVALID"      // Field Name is not a valid lowercase identifier.

	// ---------------------------------------------------------------
	// ExpressionType errors — structural type validation.

	ExprtypeEnumValuesRequired      Code = "EXPRTYPE_ENUM_VALUES_REQUIRED"       // EnumType Values is empty or nil.
	ExprtypeSetElementRequired      Code = "EXPRTYPE_SET_ELEMENT_REQUIRED"       // SetType ElementType is nil.
	ExprtypeSetElementInvalid       Code = "EXPRTYPE_SET_ELEMENT_INVALID"        // SetType ElementType failed validation.
	ExprtypeSequenceElementRequired Code = "EXPRTYPE_SEQUENCE_ELEMENT_REQUIRED"  // SequenceType ElementType is nil.
	ExprtypeSequenceElementInvalid  Code = "EXPRTYPE_SEQUENCE_ELEMENT_INVALID"   // SequenceType ElementType failed validation.
	ExprtypeBagElementRequired      Code = "EXPRTYPE_BAG_ELEMENT_REQUIRED"       // BagType ElementType is nil.
	ExprtypeBagElementInvalid       Code = "EXPRTYPE_BAG_ELEMENT_INVALID"        // BagType ElementType failed validation.
	ExprtypeTupleElementsRequired   Code = "EXPRTYPE_TUPLE_ELEMENTS_REQUIRED"    // TupleType ElementTypes is empty or nil.
	ExprtypeTupleElementNil         Code = "EXPRTYPE_TUPLE_ELEMENT_NIL"          // TupleType ElementTypes contains a nil element.
	ExprtypeTupleElementInvalid     Code = "EXPRTYPE_TUPLE_ELEMENT_INVALID"      // TupleType ElementTypes element failed validation.
	ExprtypeRecordFieldsRequired    Code = "EXPRTYPE_RECORD_FIELDS_REQUIRED"     // RecordType Fields is empty or nil.
	ExprtypeRecordFieldNameRequired Code = "EXPRTYPE_RECORD_FIELD_NAME_REQUIRED" // RecordType field Name is empty.
	ExprtypeRecordFieldTypeRequired Code = "EXPRTYPE_RECORD_FIELD_TYPE_REQUIRED" // RecordType field Type is nil.
	ExprtypeRecordFieldTypeInvalid  Code = "EXPRTYPE_RECORD_FIELD_TYPE_INVALID"  // RecordType field Type failed validation.
	ExprtypeFunctionReturnRequired  Code = "EXPRTYPE_FUNCTION_RETURN_REQUIRED"   // FunctionType Return is nil.
	ExprtypeFunctionReturnInvalid   Code = "EXPRTYPE_FUNCTION_RETURN_INVALID"    // FunctionType Return failed validation.
	ExprtypeFunctionParamNil        Code = "EXPRTYPE_FUNCTION_PARAM_NIL"         // FunctionType Params contains a nil element.
	ExprtypeFunctionParamInvalid    Code = "EXPRTYPE_FUNCTION_PARAM_INVALID"     // FunctionType Params element failed validation.
	ExprtypeObjectClasskeyInvalid   Code = "EXPRTYPE_OBJECT_CLASSKEY_INVALID"    // ObjectType ClassKey failed validation.
)
