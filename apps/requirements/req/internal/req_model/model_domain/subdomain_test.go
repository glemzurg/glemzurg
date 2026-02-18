package model_domain

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSubdomainSuite(t *testing.T) {
	suite.Run(t, new(SubdomainSuite))
}

type SubdomainSuite struct {
	suite.Suite
	domainKey identity.Key
}

func (suite *SubdomainSuite) SetupTest() {
	suite.domainKey = helper.Must(identity.NewDomainKey("domain1"))
}

// TestValidate tests all validation rules for Subdomain.
func (suite *SubdomainSuite) TestValidate() {
	validKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))

	tests := []struct {
		testName  string
		subdomain Subdomain
		errstr    string
	}{
		{
			testName: "valid subdomain",
			subdomain: Subdomain{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			subdomain: Subdomain{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			subdomain: Subdomain{
				Key:  helper.Must(identity.NewActorKey("actor1")),
				Name: "Name",
			},
			errstr: "Key: invalid key type 'actor' for subdomain",
		},
		{
			testName: "error blank name",
			subdomain: Subdomain{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.subdomain.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewSubdomain maps parameters correctly and calls Validate.
func (suite *SubdomainSuite) TestNew() {
	key := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))

	// Test parameters are mapped correctly.
	subdomain, err := NewSubdomain(key, "Name", "Details", "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Subdomain{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, subdomain)

	// Test that Validate is called (invalid data should fail).
	_, err = NewSubdomain(key, "", "Details", "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *SubdomainSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))

	// Test that Validate is called.
	subdomain := Subdomain{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := subdomain.ValidateWithParent(&suite.domainKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - subdomain key has domain1 as parent, but we pass other_domain.
	subdomain = Subdomain{
		Key:  validKey,
		Name: "Name",
	}
	err = subdomain.ValidateWithParent(&otherDomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = subdomain.ValidateWithParent(&suite.domainKey)
	assert.NoError(suite.T(), err)
}

// TestSetClassAssociations tests that SetClassAssociations validates parent relationships.
func (suite *SubdomainSuite) TestSetClassAssociations() {
	subdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "other_subdomain"))
	classKey1 := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	classKey2 := helper.Must(identity.NewClassKey(subdomainKey, "class2"))
	otherClassKey1 := helper.Must(identity.NewClassKey(otherSubdomainKey, "class1"))
	otherClassKey2 := helper.Must(identity.NewClassKey(otherSubdomainKey, "class2"))

	// Create a subdomain.
	subdomain := Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
	}

	// Test: valid association with subdomain as parent.
	validAssocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, classKey1, classKey2, "valid association"))
	validAssoc := model_class.Association{
		Key:              validAssocKey,
		Name:             "Association",
		FromClassKey:     classKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       classKey2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err := subdomain.SetClassAssociations(map[identity.Key]model_class.Association{
		validAssocKey: validAssoc,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(subdomain.ClassAssociations))

	// Test: error when association has no parent (model-level association).
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	otherDomainSubdomainKey := helper.Must(identity.NewSubdomainKey(otherDomainKey, "subdomain1"))
	crossDomainClassKey := helper.Must(identity.NewClassKey(otherDomainSubdomainKey, "class1"))
	modelLevelAssocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, classKey1, crossDomainClassKey, "model level association"))
	modelLevelAssoc := model_class.Association{
		Key:              modelLevelAssocKey,
		Name:             "Model Level Association",
		FromClassKey:     classKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       crossDomainClassKey,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err = subdomain.SetClassAssociations(map[identity.Key]model_class.Association{
		modelLevelAssocKey: modelLevelAssoc,
	})
	assert.ErrorContains(suite.T(), err, "has no parent")

	// Test: error when association parent is different subdomain.
	wrongParentAssocKey := helper.Must(identity.NewClassAssociationKey(otherSubdomainKey, otherClassKey1, otherClassKey2, "wrong parent association"))
	wrongParentAssoc := model_class.Association{
		Key:              wrongParentAssocKey,
		Name:             "Wrong Parent Association",
		FromClassKey:     otherClassKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       otherClassKey2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	err = subdomain.SetClassAssociations(map[identity.Key]model_class.Association{
		wrongParentAssocKey: wrongParentAssoc,
	})
	assert.ErrorContains(suite.T(), err, "parent does not match subdomain")
}

// TestGetClassAssociations tests that GetClassAssociations returns a copy of the associations.
func (suite *SubdomainSuite) TestGetClassAssociations() {
	subdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	classKey1 := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	classKey2 := helper.Must(identity.NewClassKey(subdomainKey, "class2"))

	// Create a subdomain with associations.
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, classKey1, classKey2, "association"))
	assoc := model_class.Association{
		Key:              assocKey,
		Name:             "Association",
		FromClassKey:     classKey1,
		FromMultiplicity: model_class.Multiplicity{LowerBound: 1, HigherBound: 1},
		ToClassKey:       classKey2,
		ToMultiplicity:   model_class.Multiplicity{LowerBound: 0, HigherBound: 0},
	}
	subdomain := Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
		ClassAssociations: map[identity.Key]model_class.Association{
			assocKey: assoc,
		},
	}

	// Test: GetClassAssociations returns the association.
	result := subdomain.GetClassAssociations()
	assert.Equal(suite.T(), 1, len(result))
	assert.Contains(suite.T(), result, assocKey)
	assert.Equal(suite.T(), assoc, result[assocKey])

	// Test: returned map is a copy, not the original.
	classKey3 := helper.Must(identity.NewClassKey(subdomainKey, "class3"))
	newAssocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, classKey1, classKey3, "new association"))
	result[newAssocKey] = model_class.Association{Key: newAssocKey, Name: "New"}
	assert.Equal(suite.T(), 1, len(subdomain.ClassAssociations), "Original should not be modified")
	assert.Equal(suite.T(), 2, len(result), "Copy should have new entry")

	// Test: empty associations returns empty map.
	emptySubdomain := Subdomain{
		Key:  subdomainKey,
		Name: "Empty Subdomain",
	}
	emptyResult := emptySubdomain.GetClassAssociations()
	assert.NotNil(suite.T(), emptyResult)
	assert.Equal(suite.T(), 0, len(emptyResult))
}

// TestValidateWithParentAndActorsAndClasses tests child validation propagation.
func (suite *SubdomainSuite) TestValidateWithParentAndActorsAndClasses() {
	subdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	classKey2 := helper.Must(identity.NewClassKey(subdomainKey, "class2"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
	useCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	useCaseKey2 := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase2"))

	actors := map[identity.Key]bool{}
	classes := map[identity.Key]bool{
		classKey:  true,
		classKey2: true,
	}

	// Test invalid Generalization child propagates error.
	subdomain := Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		Generalizations: map[identity.Key]model_class.Generalization{
			genKey: {Key: genKey, Name: ""}, // Invalid: blank name
		},
	}
	err := subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Generalizations")

	// Test invalid Class child propagates error.
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		Classes: map[identity.Key]model_class.Class{
			classKey: {Key: classKey, Name: ""}, // Invalid: blank name
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Classes")

	// Test invalid UseCase child propagates error.
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		UseCases: map[identity.Key]model_use_case.UseCase{
			useCaseKey: {Key: useCaseKey, Name: "", Level: "sea"}, // Invalid: blank name
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child UseCases")

	// Test invalid ClassAssociation child propagates error.
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, classKey, classKey2, "assoc1"))
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		ClassAssociations: map[identity.Key]model_class.Association{
			assocKey: {Key: assocKey, Name: ""}, // Invalid: blank name
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child ClassAssociations")

	// Test invalid UseCaseShares - sea-level key not a use case.
	nonExistentUseCaseKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "nonexistent"))
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		UseCases: map[identity.Key]model_use_case.UseCase{
			useCaseKey: {Key: useCaseKey, Name: "UC1", Level: "sea"},
		},
		UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
			nonExistentUseCaseKey: {
				useCaseKey: {ShareType: "include"},
			},
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "sea-level key", "Should validate UseCaseShares sea-level key")

	// Test invalid UseCaseShares - mud-level key not a use case.
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		UseCases: map[identity.Key]model_use_case.UseCase{
			useCaseKey: {Key: useCaseKey, Name: "UC1", Level: "sea"},
		},
		UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
			useCaseKey: {
				nonExistentUseCaseKey: {ShareType: "include"},
			},
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "mud-level key", "Should validate UseCaseShares mud-level key")

	// Test valid subdomain with all children.
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Name",
		Generalizations: map[identity.Key]model_class.Generalization{
			genKey: {Key: genKey, Name: "Gen"},
		},
		Classes: map[identity.Key]model_class.Class{
			classKey: {Key: classKey, Name: "Class"},
		},
		UseCases: map[identity.Key]model_use_case.UseCase{
			useCaseKey:  {Key: useCaseKey, Name: "UC1", Level: "sea"},
			useCaseKey2: {Key: useCaseKey2, Name: "UC2", Level: "mud"},
		},
		UseCaseShares: map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
			useCaseKey: {
				useCaseKey2: {ShareType: "include"},
			},
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.NoError(suite.T(), err, "Valid subdomain with all children should pass")
}

// TestValidateWithParentDeepTree tests that key validation propagates through the full tree:
// subdomain → class → guard/action/query logic keys.
func (suite *SubdomainSuite) TestValidateWithParentDeepTree() {
	subdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	reqKey := helper.Must(identity.NewActionRequireKey(actionKey, "req_1"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "query1"))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "guar_1"))
	attrKey := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	derivKey := helper.Must(identity.NewAttributeDerivationKey(attrKey, "deriv1"))

	actors := map[identity.Key]bool{}
	classes := map[identity.Key]bool{classKey: true}

	// Test valid full tree.
	subdomain := Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
		Classes: map[identity.Key]model_class.Class{
			classKey: {
				Key:  classKey,
				Name: "Class",
				Guards: map[identity.Key]model_state.Guard{
					guardKey: {Key: guardKey, Name: "Guard", Logic: model_logic.Logic{
						Key: guardKey, Description: "Guard.", Notation: model_logic.NotationTLAPlus,
					}},
				},
				Actions: map[identity.Key]model_state.Action{
					actionKey: {Key: actionKey, Name: "Action", Requires: []model_logic.Logic{
						{Key: reqKey, Description: "Req.", Notation: model_logic.NotationTLAPlus},
					}},
				},
				Queries: map[identity.Key]model_state.Query{
					queryKey: {Key: queryKey, Name: "Query", Guarantees: []model_logic.Logic{
						{Key: guarKey, Description: "Guar.", Notation: model_logic.NotationTLAPlus},
					}},
				},
				Attributes: map[identity.Key]model_class.Attribute{
					attrKey: {Key: attrKey, Name: "Attr", DerivationPolicy: &model_logic.Logic{
						Key: derivKey, Description: "Computed.", Notation: model_logic.NotationTLAPlus,
					}},
				},
			},
		},
	}
	err := subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.NoError(suite.T(), err, "Valid full tree should pass")

	// Test guard logic key mismatch is caught deep in the tree.
	otherGuardKey := helper.Must(identity.NewGuardKey(classKey, "other_guard"))
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
		Classes: map[identity.Key]model_class.Class{
			classKey: {
				Key:  classKey,
				Name: "Class",
				Guards: map[identity.Key]model_state.Guard{
					guardKey: {Key: guardKey, Name: "Guard", Logic: model_logic.Logic{
						Key: otherGuardKey, Description: "Guard.", Notation: model_logic.NotationTLAPlus,
					}},
				},
			},
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "does not match guard key", "Should catch guard logic key mismatch in deep tree")

	// Test action require key with wrong parent is caught deep in the tree.
	otherActionKey := helper.Must(identity.NewActionKey(classKey, "other_action"))
	wrongReqKey := helper.Must(identity.NewActionRequireKey(otherActionKey, "req_1"))
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
		Classes: map[identity.Key]model_class.Class{
			classKey: {
				Key:  classKey,
				Name: "Class",
				Actions: map[identity.Key]model_state.Action{
					actionKey: {Key: actionKey, Name: "Action", Requires: []model_logic.Logic{
						{Key: wrongReqKey, Description: "Req.", Notation: model_logic.NotationTLAPlus},
					}},
				},
			},
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "requires 0", "Should catch action require key error in deep tree")

	// Test attribute derivation key with wrong parent is caught deep in the tree.
	otherAttrKey := helper.Must(identity.NewAttributeKey(classKey, "other_attr"))
	wrongDerivKey := helper.Must(identity.NewAttributeDerivationKey(otherAttrKey, "deriv1"))
	subdomain = Subdomain{
		Key:  subdomainKey,
		Name: "Subdomain",
		Classes: map[identity.Key]model_class.Class{
			classKey: {
				Key:  classKey,
				Name: "Class",
				Attributes: map[identity.Key]model_class.Attribute{
					attrKey: {Key: attrKey, Name: "Attr", DerivationPolicy: &model_logic.Logic{
						Key: wrongDerivKey, Description: "Computed.", Notation: model_logic.NotationTLAPlus,
					}},
				},
			},
		},
	}
	err = subdomain.ValidateWithParentAndActorsAndClasses(&suite.domainKey, actors, classes)
	assert.ErrorContains(suite.T(), err, "DerivationPolicy", "Should catch attribute derivation key error in deep tree")
}
