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

	// Class keys.
	classKeyAA1 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa1"))
	classKeyAA2 := helper.Must(identity.NewClassKey(subdomainKeyAA, "class_aa2"))
	classKeyBA1 := helper.Must(identity.NewClassKey(subdomainKeyBA, "class_ba1"))

	// Attribute keys.
	attributeKeyAA1A := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_a"))
	attributeKeyAA1B := helper.Must(identity.NewAttributeKey(classKeyAA1, "attr_b"))

	// Derivation policy logic key (child of attribute).
	derivationKeyAA1A := helper.Must(identity.NewAttributeDerivationKey(attributeKeyAA1A, "deriv"))

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
						Classes: map[identity.Key]model_class.Class{
							classKeyAA1: {
								Key:             classKeyAA1,
								Name:            "ClassAA1",
								Details:         "Class AA1 details",
								ActorKey:        &actorKeyA,
								SuperclassOfKey: &generalizationKeyAA,
								SubclassOfKey:   &generalizationKeyAAB,
								UmlComment:      "Class AA1 UML comment",
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
