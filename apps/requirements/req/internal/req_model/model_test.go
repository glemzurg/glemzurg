package req_model

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Model.
func (suite *ModelSuite) TestValidate() {
	invKey1 := helper.Must(identity.NewInvariantKey("0"))
	invKey2 := helper.Must(identity.NewInvariantKey("1"))
	gfKey := helper.Must(identity.NewGlobalFunctionKey("_max"))
	gfKey1 := helper.Must(identity.NewGlobalFunctionKey("_some"))

	tests := []struct {
		testName string
		model    Model
		errstr   string
	}{
		{
			testName: "valid model minimal",
			model: Model{
				Key:  "model1",
				Name: "Name",
			},
		},
		{
			testName: "valid model with invariants",
			model: Model{
				Key:  "model1",
				Name: "Name",
				Invariants: []model_logic.Logic{
					{Key: invKey1, Type: model_logic.LogicTypeAssessment, Description: "x must be positive.", Notation: model_logic.NotationTLAPlus, Specification: "x > 0"},
					{Key: invKey2, Type: model_logic.LogicTypeAssessment, Description: "y must be under 100.", Notation: model_logic.NotationTLAPlus, Specification: "y < 100"},
				},
			},
		},
		{
			testName: "valid model with global functions",
			model: Model{
				Key:  "model1",
				Name: "Name",
				GlobalFunctions: map[identity.Key]model_logic.GlobalFunction{
					gfKey: {
						Key:        gfKey,
						Name:       "_Max",
						Parameters: []string{"x", "y"},
						Logic: model_logic.Logic{
							Key:           gfKey,
							Type:          model_logic.LogicTypeValue,
							Description:   "Max of two values.",
							Notation:      model_logic.NotationTLAPlus,
							Specification: "IF x > y THEN x ELSE y",
						},
					},
				},
			},
		},
		{
			testName: "valid model with invariants and global functions",
			model: Model{
				Key:  "model1",
				Name: "Name",
				Invariants: []model_logic.Logic{
					{Key: invKey1, Type: model_logic.LogicTypeAssessment, Description: "x must be positive.", Notation: model_logic.NotationTLAPlus, Specification: "x > 0"},
				},
				GlobalFunctions: map[identity.Key]model_logic.GlobalFunction{
					gfKey: {
						Key:        gfKey,
						Name:       "_Max",
						Parameters: []string{"x", "y"},
						Logic: model_logic.Logic{
							Key:           gfKey,
							Type:          model_logic.LogicTypeValue,
							Description:   "Max of two values.",
							Notation:      model_logic.NotationTLAPlus,
							Specification: "IF x > y THEN x ELSE y",
						},
					},
				},
			},
		},
		{
			testName: "error blank key",
			model: Model{
				Key:  "",
				Name: "Name",
			},
			errstr: "Key",
		},
		{
			testName: "error blank name",
			model: Model{
				Key:  "model1",
				Name: "",
			},
			errstr: "Name",
		},
		{
			testName: "error blank name with invariants set",
			model: Model{
				Key:  "model1",
				Name: "",
				Invariants: []model_logic.Logic{
					{Key: invKey1, Type: model_logic.LogicTypeAssessment, Description: "x must be positive.", Notation: model_logic.NotationTLAPlus, Specification: "x > 0"},
				},
			},
			errstr: "Name",
		},
		{
			testName: "error invalid invariant missing key",
			model: Model{
				Key:  "model1",
				Name: "Name",
				Invariants: []model_logic.Logic{
					{Key: identity.Key{}, Type: model_logic.LogicTypeAssessment, Description: "x must be positive.", Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "invariant 0",
		},
		{
			testName: "error invalid global function name",
			model: Model{
				Key:  "model1",
				Name: "Name",
				GlobalFunctions: map[identity.Key]model_logic.GlobalFunction{
					gfKey1: {
						Key:  gfKey1,
						Name: "Some", // Missing underscore
						Logic: model_logic.Logic{
							Key:         gfKey1,
							Type:        model_logic.LogicTypeValue,
							Description: "Some desc.",
							Notation:    model_logic.NotationTLAPlus,
						},
					},
				},
			},
			errstr: "must start with underscore",
		},
		{
			testName: "error global function map key mismatch",
			model: Model{
				Key:  "model1",
				Name: "Name",
				GlobalFunctions: map[identity.Key]model_logic.GlobalFunction{
					gfKey: { // Map key is gfKey ("_max")
						Key:  gfKey1, // But struct Key is gfKey1 ("_some")
						Name: "_Some",
						Logic: model_logic.Logic{
							Key:         gfKey1,
							Type:        model_logic.LogicTypeValue,
							Description: "Some desc.",
							Notation:    model_logic.NotationTLAPlus,
						},
					},
				},
			},
			errstr: "does not match function key",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewModel maps parameters correctly and calls Validate.
func (suite *ModelSuite) TestNew() {
	invKey1 := helper.Must(identity.NewInvariantKey("0"))
	gfKey := helper.Must(identity.NewGlobalFunctionKey("_max"))

	// Test all parameters are mapped correctly (key is normalized to lowercase and trimmed).
	globalFuncs := map[identity.Key]model_logic.GlobalFunction{
		gfKey: {
			Key:        gfKey,
			Name:       "_Max",
			Parameters: []string{"x", "y"},
			Logic: model_logic.Logic{
				Key:           gfKey,
				Type:          model_logic.LogicTypeValue,
				Description:   "Max of two values.",
				Notation:      model_logic.NotationTLAPlus,
				Specification: "IF x > y THEN x ELSE y",
			},
		},
	}
	invariants := []model_logic.Logic{
		{Key: invKey1, Type: model_logic.LogicTypeAssessment, Description: "First invariant.", Notation: model_logic.NotationTLAPlus, Specification: "inv1"},
	}
	model, err := NewModel("  MODEL1  ", "Name", "Details",
		invariants, globalFuncs)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Model{
		Key:             "model1",
		Name:            "Name",
		Details:         "Details",
		Invariants:      invariants,
		GlobalFunctions: globalFuncs,
	}, model)

	// Test with nil optional fields (Invariants and GlobalFunctions are optional).
	model, err = NewModel("  MODEL1  ", "Name", "Details", nil, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Model{
		Key:     "model1",
		Name:    "Name",
		Details: "Details",
	}, model)

	// Test that Validate is called (invalid data should fail).
	_, err = NewModel("model1", "", "Details", nil, nil)
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateTree tests that Validate validates the entire model tree.
// This tests that Validate validates children and their parent relationships.
func (suite *ModelSuite) TestValidateTree() {
	// Test 1: Validate validates Model fields - empty name should fail.
	model := Model{
		Key:     "model1",
		Name:    "", // Invalid - will fail Validate()
		Details: "Details",
	}
	err := model.Validate()
	assert.ErrorContains(suite.T(), err, "Name", "Validate should validate Model fields")

	// Test 2: Validate validates child Actor fields through the tree.
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Actors: map[identity.Key]model_actor.Actor{
			actorKey: {
				Key:  actorKey,
				Name: "", // Invalid - will fail Validate()
				Type: "person",
			},
		},
	}
	err = model.Validate()
	assert.ErrorContains(suite.T(), err, "Name", "Validate should validate child fields")

	// Test 3: Validate validates parent relationships - wrong parent key should fail.
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	wrongParentSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Domains: map[identity.Key]model_domain.Domain{
			otherDomainKey: {
				Key:     otherDomainKey, // Domain key is other_domain
				Name:    "Domain Name",
				Details: "Details",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					wrongParentSubdomainKey: {
						Key:     wrongParentSubdomainKey, // Parent is domain1, but attached to other_domain
						Name:    "Subdomain Name",
						Details: "Details",
					},
				},
			},
		},
	}
	err = model.Validate()
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "Validate should validate parent relationships")

	// Test 4: Validate validates child DomainAssociation fields through the tree.
	domain2Key := helper.Must(identity.NewDomainKey("domain2"))
	domainAssocKey := helper.Must(identity.NewDomainAssociationKey(domainKey, domain2Key))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey:  {Key: domainKey, Name: "Domain1"},
			domain2Key: {Key: domain2Key, Name: "Domain2"},
		},
		DomainAssociations: map[identity.Key]model_domain.Association{
			domainAssocKey: {
				Key:               domainAssocKey,
				ProblemDomainKey:  domainKey,
				SolutionDomainKey: domainKey, // Invalid - same as ProblemDomainKey
			},
		},
	}
	err = model.Validate()
	assert.ErrorContains(suite.T(), err, "ProblemDomainKey and SolutionDomainKey cannot be the same", "Validate should validate child DomainAssociations")

	// Test 5: Validate validates child ClassAssociation fields through the tree.
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domain2Key, "subdomain1"))
	classKey1 := helper.Must(identity.NewClassKey(subdomain1Key, "class1"))
	classKey2 := helper.Must(identity.NewClassKey(subdomain2Key, "class2"))
	classAssocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, classKey1, classKey2, "model assoc"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		ClassAssociations: map[identity.Key]model_class.Association{
			classAssocKey: {
				Key:          classAssocKey,
				Name:         "", // Invalid - blank name
				FromClassKey: classKey1,
				ToClassKey:   classKey2,
			},
		},
	}
	err = model.Validate()
	assert.ErrorContains(suite.T(), err, "Name", "Validate should validate child ClassAssociations")

	// Test 6: Valid model should pass.
	validSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	model = Model{
		Key:     "model1",
		Name:    "Model Name",
		Details: "Details",
		Actors: map[identity.Key]model_actor.Actor{
			actorKey: {
				Key:     actorKey,
				Name:    "Actor Name",
				Type:    "person",
				Details: "Details",
			},
		},
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:     domainKey,
				Name:    "Domain Name",
				Details: "Details",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					validSubdomainKey: {
						Key:     validSubdomainKey,
						Name:    "Subdomain Name",
						Details: "Details",
					},
				},
			},
		},
	}
	err = model.Validate()
	assert.NoError(suite.T(), err, "Valid model should pass Validate()")
}

// TestSetClassAssociations tests that SetClassAssociations validates and routes associations.
func (suite *ModelSuite) TestSetClassAssociations() {
	// Create two domains with subdomains.
	domain1Key := helper.Must(identity.NewDomainKey("domain1"))
	domain2Key := helper.Must(identity.NewDomainKey("domain2"))
	subdomain1InD1Key := helper.Must(identity.NewSubdomainKey(domain1Key, "subdomain1"))
	subdomain2InD1Key := helper.Must(identity.NewSubdomainKey(domain1Key, "subdomain2"))
	subdomain1InD2Key := helper.Must(identity.NewSubdomainKey(domain2Key, "subdomain1"))
	subdomain2InD2Key := helper.Must(identity.NewSubdomainKey(domain2Key, "subdomain2"))

	// Create classes in each subdomain.
	class1InS1D1 := helper.Must(identity.NewClassKey(subdomain1InD1Key, "class1"))
	class2InS1D1 := helper.Must(identity.NewClassKey(subdomain1InD1Key, "class2"))
	class1InS2D1 := helper.Must(identity.NewClassKey(subdomain2InD1Key, "class1"))
	class1InS1D2 := helper.Must(identity.NewClassKey(subdomain1InD2Key, "class1"))
	class1InS2D2 := helper.Must(identity.NewClassKey(subdomain2InD2Key, "class1"))

	// Create a model with two domains.
	model := Model{
		Key:  "model1",
		Name: "Model",
		Domains: map[identity.Key]model_domain.Domain{
			domain1Key: {
				Key:  domain1Key,
				Name: "Domain1",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomain1InD1Key: {Key: subdomain1InD1Key, Name: "Subdomain1"},
					subdomain2InD1Key: {Key: subdomain2InD1Key, Name: "Subdomain2"},
				},
			},
			domain2Key: {
				Key:  domain2Key,
				Name: "Domain2",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomain1InD2Key: {Key: subdomain1InD2Key, Name: "Subdomain1"},
					subdomain2InD2Key: {Key: subdomain2InD2Key, Name: "Subdomain2"},
				},
			},
		},
	}

	// Create associations at different levels:
	// 1. Model-level association (bridges domains).
	modelAssocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, class1InS1D1, class1InS1D2, "model association"))
	modelAssoc := model_class.Association{
		Key:              modelAssocKey,
		Name:             "Model Association",
		FromClassKey:     class1InS1D1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class1InS1D2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// 2. Domain1-level association (bridges subdomains in domain1).
	domain1AssocKey := helper.Must(identity.NewClassAssociationKey(domain1Key, class1InS1D1, class1InS2D1, "domain1 association"))
	domain1Assoc := model_class.Association{
		Key:              domain1AssocKey,
		Name:             "Domain1 Association",
		FromClassKey:     class1InS1D1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class1InS2D1,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// 3. Domain2-level association (bridges subdomains in domain2).
	domain2AssocKey := helper.Must(identity.NewClassAssociationKey(domain2Key, class1InS1D2, class1InS2D2, "domain2 association"))
	domain2Assoc := model_class.Association{
		Key:              domain2AssocKey,
		Name:             "Domain2 Association",
		FromClassKey:     class1InS1D2,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class1InS2D2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// 4. Subdomain-level association (within subdomain1 in domain1).
	subdomainAssocKey := helper.Must(identity.NewClassAssociationKey(subdomain1InD1Key, class1InS1D1, class2InS1D1, "subdomain association"))
	subdomainAssoc := model_class.Association{
		Key:              subdomainAssocKey,
		Name:             "Subdomain Association",
		FromClassKey:     class1InS1D1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class2InS1D1,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// Test: associations are routed correctly.
	err := model.SetClassAssociations(map[identity.Key]model_class.Association{
		modelAssocKey:     modelAssoc,
		domain1AssocKey:   domain1Assoc,
		domain2AssocKey:   domain2Assoc,
		subdomainAssocKey: subdomainAssoc,
	})
	assert.NoError(suite.T(), err)

	// Verify model-level association.
	assert.Equal(suite.T(), 1, len(model.ClassAssociations))
	assert.Contains(suite.T(), model.ClassAssociations, modelAssocKey)

	// Verify domain1 received its association.
	assert.Equal(suite.T(), 1, len(model.Domains[domain1Key].ClassAssociations))
	assert.Contains(suite.T(), model.Domains[domain1Key].ClassAssociations, domain1AssocKey)

	// Verify domain2 received its association.
	assert.Equal(suite.T(), 1, len(model.Domains[domain2Key].ClassAssociations))
	assert.Contains(suite.T(), model.Domains[domain2Key].ClassAssociations, domain2AssocKey)

	// Verify subdomain received its association (routed through domain1).
	assert.Equal(suite.T(), 1, len(model.Domains[domain1Key].Subdomains[subdomain1InD1Key].ClassAssociations))
	assert.Contains(suite.T(), model.Domains[domain1Key].Subdomains[subdomain1InD1Key].ClassAssociations, subdomainAssocKey)

	// Test: error when association has a parent that doesn't match any domain.
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(otherDomainKey, "subdomain1"))
	otherSubdomain2Key := helper.Must(identity.NewSubdomainKey(otherDomainKey, "subdomain2"))
	otherClassKey1 := helper.Must(identity.NewClassKey(otherSubdomainKey, "class1"))
	otherClassKey2 := helper.Must(identity.NewClassKey(otherSubdomain2Key, "class2"))
	wrongParentAssocKey := helper.Must(identity.NewClassAssociationKey(otherDomainKey, otherClassKey1, otherClassKey2, "wrong parent association"))
	wrongParentAssoc := model_class.Association{
		Key:              wrongParentAssocKey,
		Name:             "Wrong Parent Association",
		FromClassKey:     otherClassKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       otherClassKey2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err = model.SetClassAssociations(map[identity.Key]model_class.Association{
		wrongParentAssocKey: wrongParentAssoc,
	})
	assert.ErrorContains(suite.T(), err, "does not match any domain")
}

// TestGetClassAssociations tests that GetClassAssociations returns associations from model, domains, and subdomains.
func (suite *ModelSuite) TestGetClassAssociations() {
	// Create two domains with subdomains.
	domain1Key := helper.Must(identity.NewDomainKey("domain1"))
	domain2Key := helper.Must(identity.NewDomainKey("domain2"))
	subdomain1InD1Key := helper.Must(identity.NewSubdomainKey(domain1Key, "subdomain1"))
	subdomain2InD1Key := helper.Must(identity.NewSubdomainKey(domain1Key, "subdomain2"))
	subdomain1InD2Key := helper.Must(identity.NewSubdomainKey(domain2Key, "subdomain1"))

	// Create classes.
	class1InS1D1 := helper.Must(identity.NewClassKey(subdomain1InD1Key, "class1"))
	class2InS1D1 := helper.Must(identity.NewClassKey(subdomain1InD1Key, "class2"))
	class1InS2D1 := helper.Must(identity.NewClassKey(subdomain2InD1Key, "class1"))
	class1InS1D2 := helper.Must(identity.NewClassKey(subdomain1InD2Key, "class1"))

	// Create associations at all levels.
	// 1. Model-level association (spans domains).
	modelAssocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, class1InS1D1, class1InS1D2, "model association"))
	modelAssoc := model_class.Association{
		Key:              modelAssocKey,
		Name:             "Model Association",
		FromClassKey:     class1InS1D1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class1InS1D2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// 2. Domain-level association (spans subdomains in domain1).
	domain1AssocKey := helper.Must(identity.NewClassAssociationKey(domain1Key, class1InS1D1, class1InS2D1, "domain1 association"))
	domain1Assoc := model_class.Association{
		Key:              domain1AssocKey,
		Name:             "Domain1 Association",
		FromClassKey:     class1InS1D1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class1InS2D1,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// 3. Subdomain-level association.
	subdomainAssocKey := helper.Must(identity.NewClassAssociationKey(subdomain1InD1Key, class1InS1D1, class2InS1D1, "subdomain association"))
	subdomainAssoc := model_class.Association{
		Key:              subdomainAssocKey,
		Name:             "Subdomain Association",
		FromClassKey:     class1InS1D1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       class2InS1D1,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}

	// Create model with associations at all levels.
	model := Model{
		Key:  "model1",
		Name: "Model",
		ClassAssociations: map[identity.Key]model_class.Association{
			modelAssocKey: modelAssoc,
		},
		Domains: map[identity.Key]model_domain.Domain{
			domain1Key: {
				Key:  domain1Key,
				Name: "Domain1",
				ClassAssociations: map[identity.Key]model_class.Association{
					domain1AssocKey: domain1Assoc,
				},
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomain1InD1Key: {
						Key:  subdomain1InD1Key,
						Name: "Subdomain1",
						ClassAssociations: map[identity.Key]model_class.Association{
							subdomainAssocKey: subdomainAssoc,
						},
					},
					subdomain2InD1Key: {
						Key:  subdomain2InD1Key,
						Name: "Subdomain2",
					},
				},
			},
			domain2Key: {
				Key:  domain2Key,
				Name: "Domain2",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomain1InD2Key: {
						Key:  subdomain1InD2Key,
						Name: "Subdomain1",
					},
				},
			},
		},
	}

	// Test: GetClassAssociations returns all associations.
	result := model.GetClassAssociations()
	assert.Equal(suite.T(), 3, len(result))
	assert.Contains(suite.T(), result, modelAssocKey)
	assert.Contains(suite.T(), result, domain1AssocKey)
	assert.Contains(suite.T(), result, subdomainAssocKey)

	// Test: returned map is a copy.
	class3InS1D1 := helper.Must(identity.NewClassKey(subdomain1InD1Key, "class3"))
	newAssocKey := helper.Must(identity.NewClassAssociationKey(subdomain1InD1Key, class1InS1D1, class3InS1D1, "new association"))
	result[newAssocKey] = model_class.Association{Key: newAssocKey, Name: "New"}
	assert.Equal(suite.T(), 1, len(model.ClassAssociations), "Model associations should not be modified")
	assert.Equal(suite.T(), 1, len(model.Domains[domain1Key].ClassAssociations), "Domain associations should not be modified")
	assert.Equal(suite.T(), 1, len(model.Domains[domain1Key].Subdomains[subdomain1InD1Key].ClassAssociations), "Subdomain associations should not be modified")

	// Test: empty model returns empty map.
	emptyModel := Model{
		Key:  "empty",
		Name: "Empty Model",
	}
	emptyResult := emptyModel.GetClassAssociations()
	assert.NotNil(suite.T(), emptyResult)
	assert.Equal(suite.T(), 0, len(emptyResult))
}
