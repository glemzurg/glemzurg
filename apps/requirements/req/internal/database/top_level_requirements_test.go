package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRequirementsSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(RequirementsSuite))
}

type RequirementsSuite struct {
	suite.Suite
	db *sql.DB
}

func (suite *RequirementsSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

}

func (suite *RequirementsSuite) TestWriteRead() {

	// Invariant keys.
	invariantKeyA := helper.Must(identity.NewInvariantKey("inv_a"))
	invariantKeyB := helper.Must(identity.NewInvariantKey("inv_b"))

	// Global function keys.
	globalFunctionKeyA := helper.Must(identity.NewGlobalFunctionKey("gfunc_a"))
	globalFunctionKeyB := helper.Must(identity.NewGlobalFunctionKey("gfunc_b"))

	// Actor generalization keys.
	actorGeneralizationKeyA := helper.Must(identity.NewActorGeneralizationKey("agen_a"))
	actorGeneralizationKeyB := helper.Must(identity.NewActorGeneralizationKey("agen_b"))

	// Actor keys.
	actorKeyA := helper.Must(identity.NewActorKey("actor_a"))
	actorKeyB := helper.Must(identity.NewActorKey("actor_b"))

	// Domain keys.
	domainKeyA := helper.Must(identity.NewDomainKey("domain_a"))
	domainKeyB := helper.Must(identity.NewDomainKey("domain_b"))
	domainKeyC := helper.Must(identity.NewDomainKey("domain_c"))

	// Subdomain keys.
	subdomainKeyAA := helper.Must(identity.NewSubdomainKey(domainKeyA, "subdomain_aa"))
	subdomainKeyAB := helper.Must(identity.NewSubdomainKey(domainKeyA, "subdomain_ab"))
	subdomainKeyBA := helper.Must(identity.NewSubdomainKey(domainKeyB, "subdomain_ba"))

	// Generalization keys.
	generalizationKeyAA := helper.Must(identity.NewGeneralizationKey(subdomainKeyAA, "gen_a"))
	generalizationKeyAAB := helper.Must(identity.NewGeneralizationKey(subdomainKeyAA, "gen_b"))

	// Use case generalization keys.
	ucGeneralizationKeyAA := helper.Must(identity.NewUseCaseGeneralizationKey(subdomainKeyAA, "ucgen_a"))
	ucGeneralizationKeyAAB := helper.Must(identity.NewUseCaseGeneralizationKey(subdomainKeyAA, "ucgen_b"))

	// Use case keys.
	useCaseKeyAA := helper.Must(identity.NewUseCaseKey(subdomainKeyAA, "uc_a"))
	useCaseKeyAAB := helper.Must(identity.NewUseCaseKey(subdomainKeyAA, "uc_b"))
	useCaseKeyAAC := helper.Must(identity.NewUseCaseKey(subdomainKeyAA, "uc_c"))

	// Scenario keys.
	scenarioKeyAA := helper.Must(identity.NewScenarioKey(useCaseKeyAA, "scenario_a"))
	scenarioKeyAB := helper.Must(identity.NewScenarioKey(useCaseKeyAA, "scenario_b"))

	// Scenario object keys.
	scenarioObjectKey1 := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAA, "obj1"))
	scenarioObjectKey2 := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAA, "obj2"))

	// Scenario B object keys.
	scenarioObjectKeyB1 := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAB, "obj1"))
	scenarioObjectKeyB2 := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAB, "obj2"))

	// Step keys.
	stepKeyRoot := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "root"))
	stepKeyChild1 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "child1"))
	stepKeyChild2 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "child2"))
	stepKeyChild3 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "child3"))
	stepKeyChild4 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "child4"))

	// Scenario B step keys.
	stepKeyRootB := helper.Must(identity.NewScenarioStepKey(scenarioKeyAB, "root"))
	stepKeyChildB1 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAB, "child1"))

	// Class keys.
	classKeyAA1 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa1"))
	classKeyAA2 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa2"))
	classKeyAA3 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa3"))
	classKeyBA1 := helper.Must(identity.NewClassKey(subdomainKeyBA, "class_ba1"))

	// Attribute keys.
	attributeKeyAA1A := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_a"))
	attributeKeyAA1B := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_b"))

	// Derivation policy logic keys (children of attributes).
	derivationKeyAA1A := helper.Must(identity.NewAttributeDerivationKey(attributeKeyAA1A, "deriv"))
	derivationKeyAA1B := helper.Must(identity.NewAttributeDerivationKey(attributeKeyAA1B, "deriv"))

	// Class association keys (within subdomain AA).
	classAssociationKeyAA := helper.Must(identity.NewClassAssociationKey(subdomainKeyAA, classKeyAA1, classKeyAA2, "assoc_aa"))
	classAssociationKeyAB := helper.Must(identity.NewClassAssociationKey(subdomainKeyAA, classKeyAA2, classKeyAA1, "assoc_ab"))

	// Guard keys.
	guardKeyAA1A := helper.Must(identity.NewGuardKey(classKeyAA1, "guard_a"))
	guardKeyAA1B := helper.Must(identity.NewGuardKey(classKeyAA1, "guard_b"))

	// Action keys.
	actionKeyAA1A := helper.Must(identity.NewActionKey(classKeyAA1, "action_a"))
	actionKeyAA1B := helper.Must(identity.NewActionKey(classKeyAA1, "action_b"))

	// Action require keys (children of action key).
	actionRequireKeyA := helper.Must(identity.NewActionRequireKey(actionKeyAA1A, "req_a"))
	actionRequireKeyB := helper.Must(identity.NewActionRequireKey(actionKeyAA1A, "req_b"))

	// Action guarantee keys (children of action key).
	actionGuaranteeKeyA := helper.Must(identity.NewActionGuaranteeKey(actionKeyAA1A, "guar_a"))
	actionGuaranteeKeyB := helper.Must(identity.NewActionGuaranteeKey(actionKeyAA1A, "guar_b"))

	// Action safety keys (children of action key).
	actionSafetyKeyA := helper.Must(identity.NewActionSafetyKey(actionKeyAA1A, "safety_a"))
	actionSafetyKeyB := helper.Must(identity.NewActionSafetyKey(actionKeyAA1A, "safety_b"))

	// State keys.
	stateKeyAA1A := helper.Must(identity.NewStateKey(classKeyAA1, "state_a"))
	stateKeyAA1B := helper.Must(identity.NewStateKey(classKeyAA1, "state_b"))

	// Event keys.
	eventKeyAA1A := helper.Must(identity.NewEventKey(classKeyAA1, "event_a"))
	eventKeyAA1B := helper.Must(identity.NewEventKey(classKeyAA1, "event_b"))

	// Query keys.
	queryKeyAA1A := helper.Must(identity.NewQueryKey(classKeyAA1, "query_a"))
	queryKeyAA1B := helper.Must(identity.NewQueryKey(classKeyAA1, "query_b"))

	// Query require keys (children of query key).
	queryRequireKeyA := helper.Must(identity.NewQueryRequireKey(queryKeyAA1A, "req_a"))
	queryRequireKeyB := helper.Must(identity.NewQueryRequireKey(queryKeyAA1A, "req_b"))

	// Query guarantee keys (children of query key).
	queryGuaranteeKeyA := helper.Must(identity.NewQueryGuaranteeKey(queryKeyAA1A, "guar_a"))
	queryGuaranteeKeyB := helper.Must(identity.NewQueryGuaranteeKey(queryKeyAA1A, "guar_b"))

	// State action keys (children of state).
	stateActionKeyAA1A := helper.Must(identity.NewStateActionKey(stateKeyAA1A, "entry", "action_a"))
	stateActionKeyAA1B := helper.Must(identity.NewStateActionKey(stateKeyAA1A, "exit", "action_b"))

	// Transition keys (children of class).
	transitionKeyAA1A := helper.Must(identity.NewTransitionKey(classKeyAA1, "state_a", "event_a", "guard_a", "action_a", "state_b"))
	transitionKeyAA1B := helper.Must(identity.NewTransitionKey(classKeyAA1, "state_b", "event_b", "guard_b", "action_b", "state_a"))

	// Domain association keys.
	domainAssociationKeyAB := helper.Must(identity.NewDomainAssociationKey(domainKeyA, domainKeyB))
	domainAssociationKeyAC := helper.Must(identity.NewDomainAssociationKey(domainKeyA, domainKeyC))

	// Build the model tree.
	input := req_model.Model{
		Key:     "model_key",
		Name:    "Test Model",
		Details: "Test model details in markdown.",

		Invariants: []model_logic.Logic{
			{
				Key:           invariantKeyA,
				Description:   "Invariant A description",
				Notation:      "tla_plus",
				Specification: "InvariantA == TRUE",
			},
			{
				Key:           invariantKeyB,
				Description:   "Invariant B description",
				Notation:      "tla_plus",
				Specification: "",
			},
		},

		// Two global functions cover fk_global_logic x2.
		GlobalFunctions: map[identity.Key]model_logic.GlobalFunction{
			globalFunctionKeyA: {
				Key:        globalFunctionKeyA,
				Name:       "_Max",
				Comment:    "Returns the maximum",
				Parameters: []string{"x", "y"},
				Specification: model_logic.Logic{
					Key:           globalFunctionKeyA,
					Description:   "Max specification",
					Notation:      "tla_plus",
					Specification: "_Max(x, y) == IF x > y THEN x ELSE y",
				},
			},
			globalFunctionKeyB: {
				Key:        globalFunctionKeyB,
				Name:       "_Min",
				Comment:    "Returns the minimum",
				Parameters: []string{"a", "b"},
				Specification: model_logic.Logic{
					Key:           globalFunctionKeyB,
					Description:   "Min specification",
					Notation:      "tla_plus",
					Specification: "_Min(a, b) == IF a < b THEN a ELSE b",
				},
			},
		},

		// Actor generalizations at model level.
		ActorGeneralizations: map[identity.Key]model_actor.Generalization{
			actorGeneralizationKeyA: {
				Key:        actorGeneralizationKeyA,
				Name:       "ActorGenA",
				Details:    "Actor generalization A details",
				IsComplete: true,
				IsStatic:   false,
				UmlComment: "Actor gen A UML comment",
			},
			actorGeneralizationKeyB: {
				Key:        actorGeneralizationKeyB,
				Name:       "ActorGenB",
				Details:    "Actor generalization B details",
				IsComplete: false,
				IsStatic:   true,
				UmlComment: "Actor gen B UML comment",
			},
		},

		// Two actors covering fk_actor_model x2, fk_actor_superclass, fk_actor_subclass.
		Actors: map[identity.Key]model_actor.Actor{
			actorKeyA: {
				Key:        actorKeyA,
				Name:       "ActorA",
				Details:    "Actor A details",
				Type:       "person",
				UmlComment: "Actor A UML comment",
			},
			actorKeyB: {
				Key:        actorKeyB,
				Name:       "ActorB",
				Details:    "Actor B details",
				Type:       "system",
				UmlComment: "Actor B UML comment",
			},
		},

		// Domains with nested content.
		Domains: map[identity.Key]model_domain.Domain{
			domainKeyA: {
				Key:        domainKeyA,
				Name:       "DomainA",
				Details:    "Domain A details",
				Realized:   false,
				UmlComment: "Domain A UML comment",
				// Two subdomains cover fk_subdomain_domain x2 for domainKeyA.
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKeyAA: {
						Key:        subdomainKeyAA,
						Name:       "SubdomainAA",
						Details:    "Subdomain AA details",
						UmlComment: "Subdomain AA UML comment",
						Generalizations: map[identity.Key]model_class.Generalization{
							generalizationKeyAA: {
								Key:        generalizationKeyAA,
								Name:       "GeneralizationA",
								Details:    "Generalization A details",
								IsComplete: true,
								IsStatic:   false,
								UmlComment: "Generalization A UML comment",
							},
							generalizationKeyAAB: {
								Key:        generalizationKeyAAB,
								Name:       "GeneralizationB",
								Details:    "Generalization B details",
								IsComplete: false,
								IsStatic:   true,
								UmlComment: "Generalization B UML comment",
							},
						},
						UseCaseGeneralizations: map[identity.Key]model_use_case.Generalization{
							ucGeneralizationKeyAA: {
								Key:        ucGeneralizationKeyAA,
								Name:       "UCGeneralizationA",
								Details:    "UC Generalization A details",
								IsComplete: true,
								IsStatic:   false,
								UmlComment: "UC Generalization A UML comment",
							},
							ucGeneralizationKeyAAB: {
								Key:        ucGeneralizationKeyAAB,
								Name:       "UCGeneralizationB",
								Details:    "UC Generalization B details",
								IsComplete: false,
								IsStatic:   true,
								UmlComment: "UC Generalization B UML comment",
							},
						},
						// Three use cases: sea-level A (superclass+subclass), mud-level B (superclass), mud-level C.
						// Covers fk_use_case_superclass x2, fk_use_case_subclass.
						UseCases: map[identity.Key]model_use_case.UseCase{
							useCaseKeyAA: {
								Key:             useCaseKeyAA,
								Name:            "UseCaseA",
								Details:         "Use case A details",
								Level:           "sea",
								ReadOnly:        false,
								SuperclassOfKey: &ucGeneralizationKeyAA,
								SubclassOfKey:   &ucGeneralizationKeyAAB,
								UmlComment:      "Use case A UML comment",
								// Two use case actors cover fk_uca_use_case x2, fk_uca_actor_class x2.
								Actors: map[identity.Key]model_use_case.Actor{
									classKeyAA1: {
										UmlComment: "UC actor AA1 comment",
									},
									classKeyAA2: {
										UmlComment: "UC actor AA2 comment",
									},
								},
								// Two scenarios cover fk_scenario_use_case x2.
								Scenarios: map[identity.Key]model_scenario.Scenario{
									scenarioKeyAA: {
										Key:     scenarioKeyAA,
										Name:    "ScenarioA",
										Details: "Scenario A details",
										Steps: &model_scenario.Step{
											Key:      stepKeyRoot,
											StepType: model_scenario.STEP_TYPE_SEQUENCE,
											Statements: []model_scenario.Step{
												{
													Key:           stepKeyChild1,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
													Description:   "Send event from Obj1 to Obj2",
													FromObjectKey: &scenarioObjectKey1,
													ToObjectKey:   &scenarioObjectKey2,
													EventKey:      &eventKeyAA1A,
												},
												{
													Key:           stepKeyChild2,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
													Description:   "Query from Obj2 to Obj1",
													FromObjectKey: &scenarioObjectKey2,
													ToObjectKey:   &scenarioObjectKey1,
													QueryKey:      &queryKeyAA1A,
												},
												{
													Key:           stepKeyChild3,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
													Description:   "Send event B from Obj1 to Obj2",
													FromObjectKey: &scenarioObjectKey1,
													ToObjectKey:   &scenarioObjectKey2,
													EventKey:      &eventKeyAA1B,
												},
												// Covers fk_step_scenario_ref.
												{
													Key:           stepKeyChild4,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_SCENARIO),
													Description:   "Reference scenario B",
													FromObjectKey: &scenarioObjectKey1,
													ToObjectKey:   &scenarioObjectKey2,
													ScenarioKey:   &scenarioKeyAB,
												},
											},
										},
										Objects: map[identity.Key]model_scenario.Object{
											scenarioObjectKey1: {
												Key:          scenarioObjectKey1,
												ObjectNumber: 0,
												Name:         "Obj1",
												NameStyle:    "name",
												ClassKey:     classKeyAA1,
												Multi:        false,
												UmlComment:   "Object 1 UML comment",
											},
											scenarioObjectKey2: {
												Key:          scenarioObjectKey2,
												ObjectNumber: 1,
												Name:         "Obj2",
												NameStyle:    "name",
												ClassKey:     classKeyAA2,
												Multi:        true,
												UmlComment:   "Object 2 UML comment",
											},
										},
									},
									scenarioKeyAB: {
										Key:     scenarioKeyAB,
										Name:    "ScenarioB",
										Details: "Scenario B details",
										Steps: &model_scenario.Step{
											Key:      stepKeyRootB,
											StepType: model_scenario.STEP_TYPE_SEQUENCE,
											Statements: []model_scenario.Step{
												{
													Key:           stepKeyChildB1,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
													Description:   "Query from ObjB1 to ObjB2",
													FromObjectKey: &scenarioObjectKeyB1,
													ToObjectKey:   &scenarioObjectKeyB2,
													QueryKey:      &queryKeyAA1B,
												},
											},
										},
										Objects: map[identity.Key]model_scenario.Object{
											scenarioObjectKeyB1: {
												Key:          scenarioObjectKeyB1,
												ObjectNumber: 0,
												Name:         "ObjB1",
												NameStyle:    "name",
												ClassKey:     classKeyAA1,
												Multi:        false,
												UmlComment:   "Object B1 UML comment",
											},
											scenarioObjectKeyB2: {
												Key:          scenarioObjectKeyB2,
												ObjectNumber: 1,
												Name:         "ObjB2",
												NameStyle:    "name",
												ClassKey:     classKeyAA2,
												Multi:        false,
												UmlComment:   "Object B2 UML comment",
											},
										},
									},
								},
							},
							useCaseKeyAAB: {
								Key:             useCaseKeyAAB,
								Name:            "UseCaseB",
								Details:         "Use case B details",
								Level:           "mud",
								ReadOnly:        true,
								SuperclassOfKey: &ucGeneralizationKeyAAB,
								UmlComment:      "Use case B UML comment",
							},
							useCaseKeyAAC: {
								Key:        useCaseKeyAAC,
								Name:       "UseCaseC",
								Details:    "Use case C details",
								Level:      "mud",
								ReadOnly:   true,
								UmlComment: "Use case C UML comment",
							},
						},
						// Two shares from same sea use case cover fk_shared_sea x2, fk_shared_mud x2.
						UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
							useCaseKeyAA: {
								useCaseKeyAAB: {
									ShareType:  "include",
									UmlComment: "UC share comment",
								},
								useCaseKeyAAC: {
									ShareType:  "extend",
									UmlComment: "UC share C comment",
								},
							},
						},
						// Two class associations cover fk_association_model x2, fk_association_from x2,
						// fk_association_to x2. Second has AssociationClassKey covering fk_association_class.
						ClassAssociations: map[identity.Key]model_class.Association{
							classAssociationKeyAA: {
								Key:              classAssociationKeyAA,
								Name:             "AssociationAA",
								Details:          "Association AA details",
								FromClassKey:     classKeyAA1,
								FromMultiplicity: model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
								ToClassKey:       classKeyAA2,
								ToMultiplicity:   model_class.Multiplicity{LowerBound: 1, HigherBound: 5},
								UmlComment:       "Association AA UML comment",
							},
							classAssociationKeyAB: {
								Key:                 classAssociationKeyAB,
								Name:                "AssociationAB",
								Details:             "Association AB details",
								FromClassKey:        classKeyAA2,
								FromMultiplicity:    model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
								ToClassKey:          classKeyAA1,
								ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 3},
								AssociationClassKey: &classKeyAA3, // Covers fk_association_class.
								UmlComment:          "Association AB UML comment",
							},
						},
						Classes: map[identity.Key]model_class.Class{
							classKeyAA1: {
								Key:             classKeyAA1,
								Name:            "ClassAA1",
								Details:         "Class AA1 details",
								ActorKey:        &actorKeyA,
								SuperclassOfKey: &generalizationKeyAA,
								SubclassOfKey:   &generalizationKeyAAB,
								UmlComment:      "Class AA1 UML comment",
								// Two guards cover fk_guard_class x2, fk_guard_logic x2.
								Guards: map[identity.Key]model_state.Guard{
									guardKeyAA1A: {
										Key:  guardKeyAA1A,
										Name: "GuardA",
										Logic: model_logic.Logic{
											Key:           guardKeyAA1A,
											Description:   "Guard A logic description",
											Notation:      "tla_plus",
											Specification: "GuardA == x > 0",
										},
									},
									guardKeyAA1B: {
										Key:  guardKeyAA1B,
										Name: "GuardB",
										Logic: model_logic.Logic{
											Key:           guardKeyAA1B,
											Description:   "Guard B logic description",
											Notation:      "tla_plus",
											Specification: "GuardB == y > 0",
										},
									},
								},
								// Two actions cover fk_action_class x2.
								// ActionA has 2 params, 2 requires, 2 guarantees, 2 safeties.
								Actions: map[identity.Key]model_state.Action{
									actionKeyAA1A: {
										Key:     actionKeyAA1A,
										Name:    "ActionA",
										Details: "Action A details",
										Parameters: []model_state.Parameter{
											{
												Name:          "Amount",
												SortOrder:     0,
												DataTypeRules: "unconstrained",
												DataType: &model_data_type.DataType{
													Key:            "amount",
													CollectionType: "atomic",
													Atomic: &model_data_type.Atomic{
														ConstraintType: "unconstrained",
													},
												},
											},
											{
												Name:          "Currency",
												SortOrder:     1,
												// Covers fk_enum_atomic via enumeration data type.
												DataTypeRules: "enum of USD, EUR, GBP",
												DataType: &model_data_type.DataType{
													Key:            "currency",
													CollectionType: "atomic",
													Atomic: &model_data_type.Atomic{
														ConstraintType: "enumeration",
														EnumOrdered:    t_BoolPtr(false),
														Enums: []model_data_type.AtomicEnum{
															{Value: "USD", SortOrder: 0},
															{Value: "EUR", SortOrder: 1},
															{Value: "GBP", SortOrder: 2},
														},
													},
												},
											},
										},
										Requires: []model_logic.Logic{
											{
												Key:           actionRequireKeyA,
												Description:   "Action require A description",
												Notation:      "tla_plus",
												Specification: "ActionReqA == x > 0",
											},
											{
												Key:           actionRequireKeyB,
												Description:   "Action require B description",
												Notation:      "tla_plus",
												Specification: "ActionReqB == y > 0",
											},
										},
										Guarantees: []model_logic.Logic{
											{
												Key:           actionGuaranteeKeyA,
												Description:   "Action guarantee A description",
												Notation:      "tla_plus",
												Specification: "ActionGuarA == result' = x + 1",
											},
											{
												Key:           actionGuaranteeKeyB,
												Description:   "Action guarantee B description",
												Notation:      "tla_plus",
												Specification: "ActionGuarB == result' = y + 1",
											},
										},
										SafetyRules: []model_logic.Logic{
											{
												Key:           actionSafetyKeyA,
												Description:   "Action safety A description",
												Notation:      "tla_plus",
												Specification: "ActionSafetyA == balance' >= 0",
											},
											{
												Key:           actionSafetyKeyB,
												Description:   "Action safety B description",
												Notation:      "tla_plus",
												Specification: "ActionSafetyB == count' >= 0",
											},
										},
									},
									actionKeyAA1B: {
										Key:     actionKeyAA1B,
										Name:    "ActionB",
										Details: "Action B details",
									},
								},
								States: map[identity.Key]model_state.State{
									stateKeyAA1A: {
										Key:        stateKeyAA1A,
										Name:       "StateA",
										Details:    "State A details",
										UmlComment: "State A UML comment",
										// Two state actions cover fk_state_action_state x2, fk_state_action_action x2.
										Actions: []model_state.StateAction{
											{
												Key:       stateActionKeyAA1A,
												ActionKey: actionKeyAA1A,
												When:      "entry",
											},
											{
												Key:       stateActionKeyAA1B,
												ActionKey: actionKeyAA1B,
												When:      "exit",
											},
										},
									},
									stateKeyAA1B: {
										Key:        stateKeyAA1B,
										Name:       "StateB",
										Details:    "State B details",
										UmlComment: "",
									},
								},
								// Two events cover fk_event_class x2.
								Events: map[identity.Key]model_state.Event{
									eventKeyAA1A: {
										Key:     eventKeyAA1A,
										Name:    "EventA",
										Details: "Event A details",
										// Two event parameters cover fk_event_parameter_event x2, fk_event_parameter_data_type x2.
										Parameters: []model_state.Parameter{
											{
												Name:          "Payload",
												SortOrder:     0,
												DataTypeRules: "unconstrained",
												DataType: &model_data_type.DataType{
													Key:            "payload",
													CollectionType: "atomic",
													Atomic: &model_data_type.Atomic{
														ConstraintType: "unconstrained",
													},
												},
											},
											{
												Name:          "Timestamp",
												SortOrder:     1,
												// Covers fk_span_atomic via span data type.
												DataTypeRules: "[0 .. 999] at 1 seconds",
												DataType: &model_data_type.DataType{
													Key:            "timestamp",
													CollectionType: "atomic",
													Atomic: &model_data_type.Atomic{
														ConstraintType: "span",
														Span: &model_data_type.AtomicSpan{
															LowerType:         "closed",
															LowerValue:        t_IntPtr(0),
															LowerDenominator:  t_IntPtr(1),
															HigherType:        "closed",
															HigherValue:       t_IntPtr(999),
															HigherDenominator: t_IntPtr(1),
															Units:             "seconds",
															Precision:         1,
														},
													},
												},
											},
										},
									},
									eventKeyAA1B: {
										Key:     eventKeyAA1B,
										Name:    "EventB",
										Details: "Event B details",
									},
								},
								// Two queries cover fk_query_class x2.
								Queries: map[identity.Key]model_state.Query{
									queryKeyAA1A: {
										Key:     queryKeyAA1A,
										Name:    "QueryA",
										Details: "Query A details",
										Parameters: []model_state.Parameter{
											{
												Name:          "Limit",
												SortOrder:     0,
												DataTypeRules: "unconstrained",
												DataType: &model_data_type.DataType{
													Key:            "limit",
													CollectionType: "atomic",
													Atomic: &model_data_type.Atomic{
														ConstraintType: "unconstrained",
													},
												},
											},
											{
												Name:          "Offset",
												SortOrder:     1,
												DataTypeRules: "Int",
											},
										},
										Requires: []model_logic.Logic{
											{
												Key:           queryRequireKeyA,
												Description:   "Query require A description",
												Notation:      "tla_plus",
												Specification: "QueryReqA == TRUE",
											},
											{
												Key:           queryRequireKeyB,
												Description:   "Query require B description",
												Notation:      "tla_plus",
												Specification: "QueryReqB == x > 0",
											},
										},
										// Two query guarantees cover fk_query_guarantee_query x2.
										Guarantees: []model_logic.Logic{
											{
												Key:           queryGuaranteeKeyA,
												Description:   "Query guarantee A description",
												Notation:      "tla_plus",
												Specification: "QueryGuarA == result > 0",
											},
											{
												Key:           queryGuaranteeKeyB,
												Description:   "Query guarantee B description",
												Notation:      "tla_plus",
												Specification: "QueryGuarB == result >= 0",
											},
										},
									},
									queryKeyAA1B: {
										Key:     queryKeyAA1B,
										Name:    "QueryB",
										Details: "Query B details",
									},
								},
								// Two transitions cover fk_transition_from x2, fk_transition_event x2,
								// fk_transition_guard x2, fk_transition_action x2, fk_transition_to x2.
								Transitions: map[identity.Key]model_state.Transition{
									transitionKeyAA1A: {
										Key:          transitionKeyAA1A,
										FromStateKey: &stateKeyAA1A,
										EventKey:     eventKeyAA1A,
										GuardKey:     &guardKeyAA1A,
										ActionKey:    &actionKeyAA1A,
										ToStateKey:   &stateKeyAA1B,
										UmlComment:   "Transition A UML comment",
									},
									transitionKeyAA1B: {
										Key:          transitionKeyAA1B,
										FromStateKey: &stateKeyAA1B,
										EventKey:     eventKeyAA1B,
										GuardKey:     &guardKeyAA1B,
										ActionKey:    &actionKeyAA1B,
										ToStateKey:   &stateKeyAA1A,
										UmlComment:   "Transition B UML comment",
									},
								},
								// Two attributes both with derivation policies and data types.
								// Covers fk_attribute_derivation_logic x2, fk_attribute_data_type x2.
								// AttributeB has record data type covering fk_field_data_type, fk_field_field_data_type.
								Attributes: map[identity.Key]model_class.Attribute{
									attributeKeyAA1A: {
										Key:           attributeKeyAA1A,
										Name:          "AttributeA",
										Details:       "Attribute A details",
										DataTypeRules: "unconstrained",
										DerivationPolicy: &model_logic.Logic{
											Key:           derivationKeyAA1A,
											Description:   "Derivation A description",
											Notation:      "tla_plus",
											Specification: "DeriveA == value + 1",
										},
										Nullable:   false,
										UmlComment: "Attribute A UML comment",
										IndexNums:  []uint{1, 2},
										DataType: &model_data_type.DataType{
											Key:            attributeKeyAA1A.String(),
											CollectionType: "atomic",
											Atomic: &model_data_type.Atomic{
												ConstraintType: "unconstrained",
											},
										},
									},
									attributeKeyAA1B: {
										Key:           attributeKeyAA1B,
										Name:          "AttributeB",
										Details:       "Attribute B details",
										DataTypeRules: "constrained",
										DerivationPolicy: &model_logic.Logic{
											Key:           derivationKeyAA1B,
											Description:   "Derivation B description",
											Notation:      "tla_plus",
											Specification: "DeriveB == value - 1",
										},
										Nullable:   true,
										UmlComment: "Attribute B UML comment",
										IndexNums:  []uint{1},
										// Record data type covers fk_field_data_type, fk_field_field_data_type.
										DataType: &model_data_type.DataType{
											Key:            attributeKeyAA1B.String(),
											CollectionType: "record",
											RecordFields: []model_data_type.Field{
												{
													Name: "fieldx",
													FieldDataType: &model_data_type.DataType{
														Key:            attributeKeyAA1B.String() + "/fieldx",
														CollectionType: "atomic",
														Atomic: &model_data_type.Atomic{
															ConstraintType: "unconstrained",
														},
													},
												},
												{
													Name: "fieldy",
													FieldDataType: &model_data_type.DataType{
														Key:            attributeKeyAA1B.String() + "/fieldy",
														CollectionType: "atomic",
														Atomic: &model_data_type.Atomic{
															ConstraintType: "unconstrained",
														},
													},
												},
											},
										},
									},
								},
							},
							classKeyAA2: {
								Key:             classKeyAA2,
								Name:            "ClassAA2",
								Details:         "Class AA2 details",
								ActorKey:        &actorKeyB,            // Covers fk_class_actor x2.
								SuperclassOfKey: &generalizationKeyAAB, // Second class referencing generalization.
							},
							classKeyAA3: {
								Key:     classKeyAA3,
								Name:    "ClassAA3",
								Details: "Class AA3 details",
							},
						},
					},
					// Second subdomain in domainA covers fk_subdomain_domain x2.
					subdomainKeyAB: {
						Key:        subdomainKeyAB,
						Name:       "SubdomainAB",
						Details:    "Subdomain AB details",
						UmlComment: "Subdomain AB UML comment",
					},
				},
			},
			domainKeyB: {
				Key:        domainKeyB,
				Name:       "DomainB",
				Details:    "Domain B details",
				Realized:   true,
				UmlComment: "Domain B UML comment",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKeyBA: {
						Key:        subdomainKeyBA,
						Name:       "SubdomainBA",
						Details:    "Subdomain BA details",
						UmlComment: "Subdomain BA UML comment",
						Classes: map[identity.Key]model_class.Class{
							classKeyBA1: {
								Key:     classKeyBA1,
								Name:    "ClassBA1",
								Details: "Class BA1 details",
							},
						},
					},
				},
			},
			// Third domain enables second domain association.
			domainKeyC: {
				Key:        domainKeyC,
				Name:       "DomainC",
				Details:    "Domain C details",
				Realized:   false,
				UmlComment: "Domain C UML comment",
			},
		},

		// Two domain associations cover fk_domain_association_model x2,
		// fk_domain_association_problem x2 (both from domainKeyA),
		// fk_domain_association_solution x2 (domainKeyB and domainKeyC).
		DomainAssociations: map[identity.Key]model_domain.Association{
			domainAssociationKeyAB: {
				Key:               domainAssociationKeyAB,
				ProblemDomainKey:  domainKeyA,
				SolutionDomainKey: domainKeyB,
				UmlComment:        "Domain association AB comment",
			},
			domainAssociationKeyAC: {
				Key:               domainAssociationKeyAC,
				ProblemDomainKey:  domainKeyA,
				SolutionDomainKey: domainKeyC,
				UmlComment:        "Domain association AC comment",
			},
		},
	}

	// Validate the model tree before testing.
	err := input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	// Nothing in database yet.
	output, err := ReadModel(suite.db, "model_key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), output)

	// Write model to the database.
	err = WriteModel(suite.db, input)
	assert.Nil(suite.T(), err)

	// Write model to the database a second time, should be safe (idempotent).
	err = WriteModel(suite.db, input)
	assert.Nil(suite.T(), err)

	// Read model from the database.
	output, err = ReadModel(suite.db, "model_key")
	assert.Nil(suite.T(), err)

	// Compare the entire model tree.
	// This works because identity.Key no longer contains pointer fields.
	assert.Equal(suite.T(), input, output)
}
