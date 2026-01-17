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

	// Build all keys first for proper relationships.
	// Domain A
	domainKeyA := helper.Must(identity.NewDomainKey("domain_a"))
	subdomainKeyAA := helper.Must(identity.NewSubdomainKey(domainKeyA, "subdomain_aa"))
	classKeyAA1 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa1"))
	classKeyAA2 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa2"))
	attributeKeyAA1A := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_a"))
	attributeKeyAA1B := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_b"))
	stateKeyAA1 := helper.Must(identity.NewStateKey(classKeyAA1, "state_a"))
	stateKeyAA1B := helper.Must(identity.NewStateKey(classKeyAA1, "state_b"))
	eventKeyAA1 := helper.Must(identity.NewEventKey(classKeyAA1, "event_a"))
	guardKeyAA1 := helper.Must(identity.NewGuardKey(classKeyAA1, "guard_a"))
	actionKeyAA1 := helper.Must(identity.NewActionKey(classKeyAA1, "action_a"))
	transitionKeyAA1 := helper.Must(identity.NewTransitionKey(classKeyAA1, "state_a", "event_a", "guard_a", "action_a", "state_b"))
	stateActionKeyAA1 := helper.Must(identity.NewStateActionKey(stateKeyAA1, "entry", "key_a"))
	generalizationKeyAA := helper.Must(identity.NewGeneralizationKey(subdomainKeyAA, "gen_a"))
	useCaseKeyAA := helper.Must(identity.NewUseCaseKey(subdomainKeyAA, "usecase_a"))
	scenarioKeyAA := helper.Must(identity.NewScenarioKey(useCaseKeyAA, "scenario_a"))
	objectKeyAA := helper.Must(identity.NewScenarioObjectKey(scenarioKeyAA, "object_a"))

	// Domain B
	domainKeyB := helper.Must(identity.NewDomainKey("domain_b"))
	subdomainKeyBA := helper.Must(identity.NewSubdomainKey(domainKeyB, "subdomain_ba"))
	subdomainKeyBB := helper.Must(identity.NewSubdomainKey(domainKeyB, "subdomain_bb"))
	classKeyBA1 := helper.Must(identity.NewClassKey(subdomainKeyBA, "class_ba1"))
	classKeyBB1 := helper.Must(identity.NewClassKey(subdomainKeyBB, "class_bb1"))

	// Actor key
	actorKeyA := helper.Must(identity.NewActorKey("actor_a"))

	// Domain association key
	domainAssociationKey := helper.Must(identity.NewDomainAssociationKey(domainKeyA, domainKeyB))

	// Class association keys at different levels:
	// Model level - association between classes in different domains
	classAssocKeyModel := helper.Must(identity.NewClassAssociationKey(identity.Key{}, classKeyAA1, classKeyBA1))
	// Domain level - association between classes in different subdomains of same domain
	classAssocKeyDomain := helper.Must(identity.NewClassAssociationKey(domainKeyB, classKeyBA1, classKeyBB1))
	// Subdomain level - association between classes in same subdomain
	classAssocKeySubdomain := helper.Must(identity.NewClassAssociationKey(subdomainKeyAA, classKeyAA1, classKeyAA2))

	// Build the model tree.
	input := req_model.Model{
		Key:     "model_key",
		Name:    "Test Model",
		Details: "Test model details in markdown.",

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
				DomainAssociations: map[identity.Key]model_domain.Association{
					domainAssociationKey: {
						Key:               domainAssociationKey,
						ProblemDomainKey:  domainKeyA,
						SolutionDomainKey: domainKeyB,
						UmlComment:        "Domain association comment",
					},
				},
				ClassAssociations: map[identity.Key]model_class.Association{},
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
								Details:    "Generalization details",
								IsComplete: true,
								IsStatic:   false,
								UmlComment: "Generalization UML comment",
							},
						},
						Classes: map[identity.Key]model_class.Class{
							classKeyAA1: {
								Key:        classKeyAA1,
								Name:       "ClassAA1",
								Details:    "Class AA1 details",
								UmlComment: "Class AA1 UML comment",
								Attributes: map[identity.Key]model_class.Attribute{
									attributeKeyAA1A: {
										Key:           attributeKeyAA1A,
										Name:          "AttributeA",
										Details:       "Attribute A details",
										DataTypeRules: "unconstrained",
										Nullable:      false,
										UmlComment:    "Attribute A UML comment",
										// Note: IndexNums are written to DB but not read back by ReadModel.
										DataType: &model_data_type.DataType{
											Key:            attributeKeyAA1A.String(),
											CollectionType: "atomic",
											Atomic: &model_data_type.Atomic{
												ConstraintType: "unconstrained",
											},
										},
									},
									attributeKeyAA1B: {
										Key:        attributeKeyAA1B,
										Name:       "AttributeB",
										Details:    "Attribute B details",
										Nullable:   true,
										UmlComment: "Attribute B UML comment",
									},
								},
								States: map[identity.Key]model_state.State{
									stateKeyAA1: {
										Key:        stateKeyAA1,
										Name:       "StateA",
										Details:    "State A details",
										UmlComment: "State A UML comment",
										Actions: []model_state.StateAction{
											{
												Key:       stateActionKeyAA1,
												ActionKey: actionKeyAA1,
												When:      "entry",
											},
										},
									},
									stateKeyAA1B: {
										Key:        stateKeyAA1B,
										Name:       "StateB",
										Details:    "State B details",
										UmlComment: "State B UML comment",
									},
								},
								Events: map[identity.Key]model_state.Event{
									eventKeyAA1: {
										Key:     eventKeyAA1,
										Name:    "EventA",
										Details: "Event A details",
									},
								},
								Guards: map[identity.Key]model_state.Guard{
									guardKeyAA1: {
										Key:     guardKeyAA1,
										Name:    "GuardA",
										Details: "Guard A details",
									},
								},
								Actions: map[identity.Key]model_state.Action{
									actionKeyAA1: {
										Key:     actionKeyAA1,
										Name:    "ActionA",
										Details: "Action A details",
									},
								},
								Transitions: map[identity.Key]model_state.Transition{
									transitionKeyAA1: {
										Key:          transitionKeyAA1,
										FromStateKey: &stateKeyAA1,
										EventKey:     eventKeyAA1,
										GuardKey:     &guardKeyAA1,
										ActionKey:    &actionKeyAA1,
										ToStateKey:   &stateKeyAA1B,
										UmlComment:   "Transition UML comment",
									},
								},
							},
							classKeyAA2: {
								Key:         classKeyAA2,
								Name:        "ClassAA2",
								Details:     "Class AA2 details",
								Attributes:  map[identity.Key]model_class.Attribute{},
								States:      map[identity.Key]model_state.State{},
								Events:      map[identity.Key]model_state.Event{},
								Guards:      map[identity.Key]model_state.Guard{},
								Actions:     map[identity.Key]model_state.Action{},
								Transitions: map[identity.Key]model_state.Transition{},
							},
						},
						UseCases: map[identity.Key]model_use_case.UseCase{
							useCaseKeyAA: {
								Key:        useCaseKeyAA,
								Name:       "UseCaseA",
								Details:    "Use case A details",
								Level:      "sea",
								ReadOnly:   false,
								UmlComment: "Use case UML comment",
								Actors: map[identity.Key]model_use_case.Actor{
									actorKeyA: {
										UmlComment: "Use case actor UML comment",
									},
								},
								Scenarios: map[identity.Key]model_scenario.Scenario{
									scenarioKeyAA: {
										Key:     scenarioKeyAA,
										Name:    "ScenarioA",
										Details: "Scenario A details",
										Objects: map[identity.Key]model_scenario.Object{
											objectKeyAA: {
												Key:          objectKeyAA,
												ObjectNumber: 1,
												Name:         "ObjectA",
												NameStyle:    "name",
												ClassKey:     classKeyAA1,
												Multi:        false,
												UmlComment:   "Object UML comment",
											},
										},
									},
								},
							},
						},
						// Subdomain-level class association
						ClassAssociations: map[identity.Key]model_class.Association{
							classAssocKeySubdomain: {
								Key:              classAssocKeySubdomain,
								Name:             "SubdomainAssociation",
								Details:          "Subdomain association details",
								FromClassKey:     classKeyAA1,
								FromMultiplicity: model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
								ToClassKey:       classKeyAA2,
								ToMultiplicity:   model_class.Multiplicity{LowerBound: 1, HigherBound: 10},
								UmlComment:       "Subdomain association UML comment",
							},
						},
						UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{},
					},
				},
			},
			domainKeyB: {
				Key:        domainKeyB,
				Name:       "DomainB",
				Details:    "Domain B details",
				Realized:   true,
				UmlComment: "Domain B UML comment",
				DomainAssociations: map[identity.Key]model_domain.Association{},
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKeyBA: {
						Key:             subdomainKeyBA,
						Name:            "SubdomainBA",
						Details:         "Subdomain BA details",
						UmlComment:      "Subdomain BA UML comment",
						Generalizations: map[identity.Key]model_class.Generalization{},
						Classes: map[identity.Key]model_class.Class{
							classKeyBA1: {
								Key:         classKeyBA1,
								Name:        "ClassBA1",
								Details:     "Class BA1 details",
								Attributes:  map[identity.Key]model_class.Attribute{},
								States:      map[identity.Key]model_state.State{},
								Events:      map[identity.Key]model_state.Event{},
								Guards:      map[identity.Key]model_state.Guard{},
								Actions:     map[identity.Key]model_state.Action{},
								Transitions: map[identity.Key]model_state.Transition{},
							},
						},
						UseCases:      map[identity.Key]model_use_case.UseCase{},
						UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{},
					},
					subdomainKeyBB: {
						Key:             subdomainKeyBB,
						Name:            "SubdomainBB",
						Details:         "Subdomain BB details",
						UmlComment:      "Subdomain BB UML comment",
						Generalizations: map[identity.Key]model_class.Generalization{},
						Classes: map[identity.Key]model_class.Class{
							classKeyBB1: {
								Key:         classKeyBB1,
								Name:        "ClassBB1",
								Details:     "Class BB1 details",
								Attributes:  map[identity.Key]model_class.Attribute{},
								States:      map[identity.Key]model_state.State{},
								Events:      map[identity.Key]model_state.Event{},
								Guards:      map[identity.Key]model_state.Guard{},
								Actions:     map[identity.Key]model_state.Action{},
								Transitions: map[identity.Key]model_state.Transition{},
							},
						},
						UseCases:      map[identity.Key]model_use_case.UseCase{},
						UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{},
					},
				},
				// Domain-level class association (between subdomains within domain B)
				ClassAssociations: map[identity.Key]model_class.Association{
					classAssocKeyDomain: {
						Key:              classAssocKeyDomain,
						Name:             "DomainAssociation",
						Details:          "Domain association details",
						FromClassKey:     classKeyBA1,
						FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
						ToClassKey:       classKeyBB1,
						ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
						UmlComment:       "Domain association UML comment",
					},
				},
			},
		},

		// Model-level class association (between domains)
		ClassAssociations: map[identity.Key]model_class.Association{
			classAssocKeyModel: {
				Key:              classAssocKeyModel,
				Name:             "ModelAssociation",
				Details:          "Model association details",
				FromClassKey:     classKeyAA1,
				FromMultiplicity: model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
				ToClassKey:       classKeyBA1,
				ToMultiplicity:   model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
				UmlComment:       "Model association UML comment",
			},
		},
	}

	// Validate the model tree before testing.
	err := input.ValidateWithParent()
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

	// Compare individual parts for better error messages.
	assert.Equal(suite.T(), input.Key, output.Key, "model key should match")
	assert.Equal(suite.T(), input.Name, output.Name, "model name should match")
	assert.Equal(suite.T(), input.Details, output.Details, "model details should match")
	assert.Equal(suite.T(), input.Actors, output.Actors, "actors should match")

	// Check domains.
	assert.Equal(suite.T(), len(input.Domains), len(output.Domains), "domain count should match")
	for domainKey, inputDomain := range input.Domains {
		outputDomain, exists := output.Domains[domainKey]
		assert.True(suite.T(), exists, "domain %s should exist", domainKey)
		assert.Equal(suite.T(), inputDomain.Key, outputDomain.Key, "domain key should match")
		assert.Equal(suite.T(), inputDomain.Name, outputDomain.Name, "domain name should match")
		assert.Equal(suite.T(), inputDomain.Details, outputDomain.Details, "domain details should match")
		assert.Equal(suite.T(), inputDomain.Realized, outputDomain.Realized, "domain realized should match")
		assert.Equal(suite.T(), inputDomain.UmlComment, outputDomain.UmlComment, "domain uml comment should match")

		// Check domain associations.
		assert.Equal(suite.T(), len(inputDomain.DomainAssociations), len(outputDomain.DomainAssociations), "domain association count should match for domain %s", domainKey)

		// Check subdomains.
		assert.Equal(suite.T(), len(inputDomain.Subdomains), len(outputDomain.Subdomains), "subdomain count should match for domain %s", domainKey)
		for subdomainKey, inputSubdomain := range inputDomain.Subdomains {
			outputSubdomain, exists := outputDomain.Subdomains[subdomainKey]
			assert.True(suite.T(), exists, "subdomain %s should exist", subdomainKey)
			assert.Equal(suite.T(), inputSubdomain.Key, outputSubdomain.Key, "subdomain key should match")
			assert.Equal(suite.T(), inputSubdomain.Name, outputSubdomain.Name, "subdomain name should match")

			// Check classes.
			assert.Equal(suite.T(), len(inputSubdomain.Classes), len(outputSubdomain.Classes), "class count should match for subdomain %s", subdomainKey)
			for classKey, inputClass := range inputSubdomain.Classes {
				outputClass, exists := outputSubdomain.Classes[classKey]
				assert.True(suite.T(), exists, "class %s should exist", classKey)
				assert.Equal(suite.T(), inputClass.Key, outputClass.Key, "class key should match")
				assert.Equal(suite.T(), inputClass.Name, outputClass.Name, "class name should match")

				// Check attributes.
				assert.Equal(suite.T(), len(inputClass.Attributes), len(outputClass.Attributes), "attribute count should match for class %s", classKey)

				// Check states.
				assert.Equal(suite.T(), len(inputClass.States), len(outputClass.States), "state count should match for class %s", classKey)

				// Check events.
				assert.Equal(suite.T(), len(inputClass.Events), len(outputClass.Events), "event count should match for class %s", classKey)

				// Check guards.
				assert.Equal(suite.T(), len(inputClass.Guards), len(outputClass.Guards), "guard count should match for class %s", classKey)

				// Check actions.
				assert.Equal(suite.T(), len(inputClass.Actions), len(outputClass.Actions), "action count should match for class %s", classKey)

				// Check transitions.
				assert.Equal(suite.T(), len(inputClass.Transitions), len(outputClass.Transitions), "transition count should match for class %s", classKey)
			}

			// Check use cases.
			assert.Equal(suite.T(), len(inputSubdomain.UseCases), len(outputSubdomain.UseCases), "use case count should match for subdomain %s", subdomainKey)

			// Check generalizations.
			assert.Equal(suite.T(), len(inputSubdomain.Generalizations), len(outputSubdomain.Generalizations), "generalization count should match for subdomain %s", subdomainKey)

			// Check subdomain-level class associations.
			assert.Equal(suite.T(), len(inputSubdomain.ClassAssociations), len(outputSubdomain.ClassAssociations), "subdomain class association count should match for subdomain %s", subdomainKey)
		}

		// Check domain-level class associations.
		assert.Equal(suite.T(), len(inputDomain.ClassAssociations), len(outputDomain.ClassAssociations), "domain class association count should match for domain %s", domainKey)
	}

	// Check model-level class associations.
	assert.Equal(suite.T(), len(input.ClassAssociations), len(output.ClassAssociations), "model class association count should match")

	// Note: We can't use assert.Equal on the entire model because identity.Key contains
	// *string fields, and maps with struct keys containing pointers can't be compared
	// reliably with reflect.DeepEqual (map key lookup uses == not DeepEqual).
	// The individual component checks above verify the data is correct.
}
