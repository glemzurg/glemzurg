package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
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

	// Actor generalization keys.
	actorGeneralizationKeyA := helper.Must(identity.NewActorGeneralizationKey("agen_a"))
	actorGeneralizationKeyB := helper.Must(identity.NewActorGeneralizationKey("agen_b"))

	// Actor keys.
	actorKeyA := helper.Must(identity.NewActorKey("actor_a"))

	// Domain keys.
	domainKeyA := helper.Must(identity.NewDomainKey("domain_a"))
	domainKeyB := helper.Must(identity.NewDomainKey("domain_b"))

	// Subdomain keys.
	subdomainKeyAA := helper.Must(identity.NewSubdomainKey(domainKeyA, "subdomain_aa"))
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

	// Scenario keys.
	scenarioKeyAA := helper.Must(identity.NewScenarioKey(useCaseKeyAA, "scenario_a"))

	// Scenario object keys.
	scenarioObjectKey1 := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAA, "obj1"))
	scenarioObjectKey2 := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAA, "obj2"))

	// Step keys.
	stepKeyRoot := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "root"))
	stepKeyChild1 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "child1"))
	stepKeyChild2 := helper.Must(identity.NewScenarioStepKey(scenarioKeyAA, "child2"))

	// Class keys.
	classKeyAA1 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa1"))
	classKeyAA2 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa2"))
	classKeyBA1 := helper.Must(identity.NewClassKey(subdomainKeyBA, "class_ba1"))

	// Attribute keys.
	attributeKeyAA1A := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_a"))
	attributeKeyAA1B := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_b"))

	// Derivation policy logic key (child of attribute).
	derivationKeyAA1A := helper.Must(identity.NewAttributeDerivationKey(attributeKeyAA1A, "deriv"))

	// Class association key (within subdomain AA).
	classAssociationKeyAA := helper.Must(identity.NewClassAssociationKey(subdomainKeyAA, classKeyAA1, classKeyAA2, "assoc_aa"))

	// Guard keys.
	guardKeyAA1A := helper.Must(identity.NewGuardKey(classKeyAA1, "guard_a"))

	// Action keys.
	actionKeyAA1A := helper.Must(identity.NewActionKey(classKeyAA1, "action_a"))

	// Action require keys (children of action key).
	actionRequireKeyA := helper.Must(identity.NewActionRequireKey(actionKeyAA1A, "req_a"))

	// Action guarantee keys (children of action key).
	actionGuaranteeKeyA := helper.Must(identity.NewActionGuaranteeKey(actionKeyAA1A, "guar_a"))

	// Action safety keys (children of action key).
	actionSafetyKeyA := helper.Must(identity.NewActionSafetyKey(actionKeyAA1A, "safety_a"))

	// State keys.
	stateKeyAA1A := helper.Must(identity.NewStateKey(classKeyAA1, "state_a"))
	stateKeyAA1B := helper.Must(identity.NewStateKey(classKeyAA1, "state_b"))

	// Event keys.
	eventKeyAA1A := helper.Must(identity.NewEventKey(classKeyAA1, "event_a"))

	// Query keys.
	queryKeyAA1A := helper.Must(identity.NewQueryKey(classKeyAA1, "query_a"))

	// Query require keys (children of query key).
	queryRequireKeyA := helper.Must(identity.NewQueryRequireKey(queryKeyAA1A, "req_a"))
	queryRequireKeyB := helper.Must(identity.NewQueryRequireKey(queryKeyAA1A, "req_b"))

	// Query guarantee keys (children of query key).
	queryGuaranteeKeyA := helper.Must(identity.NewQueryGuaranteeKey(queryKeyAA1A, "guar_a"))

	// State action key (child of state).
	stateActionKeyAA1A := helper.Must(identity.NewStateActionKey(stateKeyAA1A, "entry", "action_a"))

	// Transition key (child of class).
	transitionKeyAA1A := helper.Must(identity.NewTransitionKey(classKeyAA1, "state_a", "event_a", "guard_a", "action_a", "state_b"))

	// Domain association key.
	domainAssociationKey := helper.Must(identity.NewDomainAssociationKey(domainKeyA, domainKeyB))

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

		// Actors at model level.
		Actors: map[identity.Key]model_actor.Actor{
			actorKeyA: {
				Key:        actorKeyA,
				Name:       "ActorA",
				Details:    "Actor A details",
				Type:       "person",
				UmlComment: "Actor UML comment",
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
								Actors: map[identity.Key]model_use_case.Actor{
									classKeyAA1: {
										UmlComment: "UC actor AA1 comment",
									},
								},
								Scenarios: map[identity.Key]model_scenario.Scenario{
									scenarioKeyAA: {
										Key:     scenarioKeyAA,
										Name:    "ScenarioA",
										Details: "Scenario A details",
										Steps: &model_scenario.Step{
											Key:       stepKeyRoot,
											SortOrder: 0,
											StepType:  model_scenario.STEP_TYPE_SEQUENCE,
											Statements: []model_scenario.Step{
												{
													Key:           stepKeyChild1,
													SortOrder:     0,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_EVENT),
													Description:   "Send event from Obj1 to Obj2",
													FromObjectKey: &scenarioObjectKey1,
													ToObjectKey:   &scenarioObjectKey2,
													EventKey:      &eventKeyAA1A,
												},
												{
													Key:           stepKeyChild2,
													SortOrder:     1,
													StepType:      model_scenario.STEP_TYPE_LEAF,
													LeafType:      t_strPtr(model_scenario.LEAF_TYPE_QUERY),
													Description:   "Query from Obj2 to Obj1",
													FromObjectKey: &scenarioObjectKey2,
													ToObjectKey:   &scenarioObjectKey1,
													QueryKey:      &queryKeyAA1A,
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
								},
							},
							useCaseKeyAAB: {
								Key:        useCaseKeyAAB,
								Name:       "UseCaseB",
								Details:    "Use case B details",
								Level:      "mud",
								ReadOnly:   true,
								UmlComment: "Use case B UML comment",
							},
						},
						UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
							useCaseKeyAA: {
								useCaseKeyAAB: {
									ShareType:  "include",
									UmlComment: "UC share comment",
								},
							},
						},
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
								},
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
										},
										Requires: []model_logic.Logic{
											{
												Key:           actionRequireKeyA,
												Description:   "Action require A description",
												Notation:      "tla_plus",
												Specification: "ActionReqA == x > 0",
											},
										},
										Guarantees: []model_logic.Logic{
											{
												Key:           actionGuaranteeKeyA,
												Description:   "Action guarantee A description",
												Notation:      "tla_plus",
												Specification: "ActionGuarA == result' = x + 1",
											},
										},
										SafetyRules: []model_logic.Logic{
											{
												Key:           actionSafetyKeyA,
												Description:   "Action safety A description",
												Notation:      "tla_plus",
												Specification: "ActionSafetyA == balance' >= 0",
											},
										},
									},
								},
								States: map[identity.Key]model_state.State{
									stateKeyAA1A: {
										Key:        stateKeyAA1A,
										Name:       "StateA",
										Details:    "State A details",
										UmlComment: "State A UML comment",
										Actions: []model_state.StateAction{
											{
												Key:       stateActionKeyAA1A,
												ActionKey: actionKeyAA1A,
												When:      "entry",
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
								Events: map[identity.Key]model_state.Event{
									eventKeyAA1A: {
										Key:     eventKeyAA1A,
										Name:    "EventA",
										Details: "Event A details",
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
										},
									},
								},
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
										Guarantees: []model_logic.Logic{
											{
												Key:           queryGuaranteeKeyA,
												Description:   "Query guarantee A description",
												Notation:      "tla_plus",
												Specification: "QueryGuarA == result > 0",
											},
										},
									},
								},
								Transitions: map[identity.Key]model_state.Transition{
									transitionKeyAA1A: {
										Key:          transitionKeyAA1A,
										FromStateKey: &stateKeyAA1A,
										EventKey:     eventKeyAA1A,
										GuardKey:     &guardKeyAA1A,
										ActionKey:    &actionKeyAA1A,
										ToStateKey:   &stateKeyAA1B,
										UmlComment:   "Transition UML comment",
									},
								},
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
										Key:              attributeKeyAA1B,
										Name:             "AttributeB",
										Details:          "Attribute B details",
										DataTypeRules:    "constrained",
										DerivationPolicy: nil,
										Nullable:         true,
										UmlComment:       "Attribute B UML comment",
										IndexNums:        []uint{1},
									},
								},
							},
							classKeyAA2: {
								Key:     classKeyAA2,
								Name:    "ClassAA2",
								Details: "Class AA2 details",
							},
						},
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
		},

		// Model-level domain associations.
		DomainAssociations: map[identity.Key]model_domain.Association{
			domainAssociationKey: {
				Key:               domainAssociationKey,
				ProblemDomainKey:  domainKeyA,
				SolutionDomainKey: domainKeyB,
				UmlComment:        "Domain association comment",
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
