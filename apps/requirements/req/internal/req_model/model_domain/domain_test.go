package model_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Domain.
func (suite *DomainSuite) TestValidate() {
	validKey := helper.Must(identity.NewDomainKey("domain1"))

	tests := []struct {
		testName string
		domain   Domain
		errstr   string
	}{
		{
			testName: "valid domain",
			domain: Domain{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			domain: Domain{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			domain: Domain{
				Key:  helper.Must(identity.NewActorKey("actor1")),
				Name: "Name",
			},
			errstr: "Key: invalid key type 'actor' for domain",
		},
		{
			testName: "error blank name",
			domain: Domain{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.domain.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewDomain maps parameters correctly and calls Validate.
func (suite *DomainSuite) TestNew() {
	key := helper.Must(identity.NewDomainKey("domain1"))

	// Test parameters are mapped correctly.
	domain, err := NewDomain(key, "Name", "Details", true, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Domain{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	}, domain)

	// Test that Validate is called (invalid data should fail).
	_, err = NewDomain(key, "", "Details", true, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *DomainSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewDomainKey("domain1"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))

	// Test that Validate is called.
	domain := Domain{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := domain.ValidateWithParent(nil)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - domains should have nil parent.
	domain = Domain{
		Key:  validKey,
		Name: "Name",
	}
	err = domain.ValidateWithParent(&otherDomainKey)
	assert.ErrorContains(suite.T(), err, "should not have a parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = domain.ValidateWithParent(nil)
	assert.NoError(suite.T(), err)
}

// TestSetClassAssociations tests that SetClassAssociations validates and routes associations.
func (suite *DomainSuite) TestSetClassAssociations() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain2"))
	class1InSub1 := helper.Must(identity.NewClassKey(subdomain1Key, "class1"))
	class2InSub1 := helper.Must(identity.NewClassKey(subdomain1Key, "class2"))
	class1InSub2 := helper.Must(identity.NewClassKey(subdomain2Key, "class1"))
	class2InSub2 := helper.Must(identity.NewClassKey(subdomain2Key, "class2"))

	// Create a domain with two subdomains.
	domain := Domain{
		Key:  domainKey,
		Name: "Domain",
		Subdomains: map[identity.Key]Subdomain{
			subdomain1Key: {Key: subdomain1Key, Name: "Subdomain1"},
			subdomain2Key: {Key: subdomain2Key, Name: "Subdomain2"},
		},
	}

	// Create associations:
	// 1. Domain-level association (bridges subdomains).
	domainAssocKey := helper.Must(identity.NewClassAssociationKey(domainKey, class1InSub1, class1InSub2, "domain association"))
	domainAssoc := helper.Must(model_class.NewAssociation(domainAssocKey, "Domain Association", "", class1InSub1, helper.Must(model_class.NewMultiplicity("1")), class1InSub2, helper.Must(model_class.NewMultiplicity("any")), nil, ""))

	// 2. Subdomain1-level association.
	sub1AssocKey := helper.Must(identity.NewClassAssociationKey(subdomain1Key, class1InSub1, class2InSub1, "subdomain1 association"))
	sub1Assoc := helper.Must(model_class.NewAssociation(sub1AssocKey, "Subdomain1 Association", "", class1InSub1, helper.Must(model_class.NewMultiplicity("1")), class2InSub1, helper.Must(model_class.NewMultiplicity("any")), nil, ""))

	// 3. Subdomain2-level association.
	sub2AssocKey := helper.Must(identity.NewClassAssociationKey(subdomain2Key, class1InSub2, class2InSub2, "subdomain2 association"))
	sub2Assoc := helper.Must(model_class.NewAssociation(sub2AssocKey, "Subdomain2 Association", "", class1InSub2, helper.Must(model_class.NewMultiplicity("1")), class2InSub2, helper.Must(model_class.NewMultiplicity("any")), nil, ""))

	// Test: associations are routed correctly.
	err := domain.SetClassAssociations(map[identity.Key]model_class.Association{
		domainAssocKey: domainAssoc,
		sub1AssocKey:   sub1Assoc,
		sub2AssocKey:   sub2Assoc,
	})
	assert.NoError(suite.T(), err)

	// Verify domain-level association.
	assert.Equal(suite.T(), 1, len(domain.ClassAssociations))
	assert.Contains(suite.T(), domain.ClassAssociations, domainAssocKey)

	// Verify subdomain1 received its association.
	assert.Equal(suite.T(), 1, len(domain.Subdomains[subdomain1Key].ClassAssociations))
	assert.Contains(suite.T(), domain.Subdomains[subdomain1Key].ClassAssociations, sub1AssocKey)

	// Verify subdomain2 received its association.
	assert.Equal(suite.T(), 1, len(domain.Subdomains[subdomain2Key].ClassAssociations))
	assert.Contains(suite.T(), domain.Subdomains[subdomain2Key].ClassAssociations, sub2AssocKey)

	// Test: error when association has no parent (model-level association).
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(otherDomainKey, "subdomain1"))
	crossDomainClassKey := helper.Must(identity.NewClassKey(otherSubdomainKey, "class1"))
	modelLevelAssocKey := helper.Must(identity.NewClassAssociationKey(identity.Key{}, class1InSub1, crossDomainClassKey, "model level association"))
	modelLevelAssoc := helper.Must(model_class.NewAssociation(modelLevelAssocKey, "Model Level Association", "", class1InSub1, helper.Must(model_class.NewMultiplicity("1")), crossDomainClassKey, helper.Must(model_class.NewMultiplicity("any")), nil, ""))
	err = domain.SetClassAssociations(map[identity.Key]model_class.Association{
		modelLevelAssocKey: modelLevelAssoc,
	})
	assert.ErrorContains(suite.T(), err, "has no parent")

	// Test: error when association parent is a different domain.
	// For domain-level association, we need two classes in different subdomains of that domain.
	otherSubdomain2Key := helper.Must(identity.NewSubdomainKey(otherDomainKey, "subdomain2"))
	crossDomainClassKey2 := helper.Must(identity.NewClassKey(otherSubdomain2Key, "class2"))
	wrongDomainAssocKey := helper.Must(identity.NewClassAssociationKey(otherDomainKey, crossDomainClassKey, crossDomainClassKey2, "wrong domain association"))
	wrongDomainAssoc := helper.Must(model_class.NewAssociation(wrongDomainAssocKey, "Wrong Domain Association", "", crossDomainClassKey, helper.Must(model_class.NewMultiplicity("1")), crossDomainClassKey2, helper.Must(model_class.NewMultiplicity("any")), nil, ""))
	err = domain.SetClassAssociations(map[identity.Key]model_class.Association{
		wrongDomainAssocKey: wrongDomainAssoc,
	})
	assert.ErrorContains(suite.T(), err, "parent does not match domain")
}

// TestGetClassAssociations tests that GetClassAssociations returns associations from domain and subdomains.
func (suite *DomainSuite) TestGetClassAssociations() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain2"))
	class1InSub1 := helper.Must(identity.NewClassKey(subdomain1Key, "class1"))
	class2InSub1 := helper.Must(identity.NewClassKey(subdomain1Key, "class2"))
	class1InSub2 := helper.Must(identity.NewClassKey(subdomain2Key, "class1"))
	class2InSub2 := helper.Must(identity.NewClassKey(subdomain2Key, "class2"))

	// Create associations at different levels.
	domainAssocKey := helper.Must(identity.NewClassAssociationKey(domainKey, class1InSub1, class1InSub2, "domain association"))
	domainAssoc := helper.Must(model_class.NewAssociation(domainAssocKey, "Domain Association", "", class1InSub1, helper.Must(model_class.NewMultiplicity("1")), class1InSub2, helper.Must(model_class.NewMultiplicity("any")), nil, ""))

	sub1AssocKey := helper.Must(identity.NewClassAssociationKey(subdomain1Key, class1InSub1, class2InSub1, "subdomain1 association"))
	sub1Assoc := helper.Must(model_class.NewAssociation(sub1AssocKey, "Subdomain1 Association", "", class1InSub1, helper.Must(model_class.NewMultiplicity("1")), class2InSub1, helper.Must(model_class.NewMultiplicity("any")), nil, ""))

	sub2AssocKey := helper.Must(identity.NewClassAssociationKey(subdomain2Key, class1InSub2, class2InSub2, "subdomain2 association"))
	sub2Assoc := helper.Must(model_class.NewAssociation(sub2AssocKey, "Subdomain2 Association", "", class1InSub2, helper.Must(model_class.NewMultiplicity("1")), class2InSub2, helper.Must(model_class.NewMultiplicity("any")), nil, ""))

	// Create domain with associations at all levels.
	domain := Domain{
		Key:  domainKey,
		Name: "Domain",
		ClassAssociations: map[identity.Key]model_class.Association{
			domainAssocKey: domainAssoc,
		},
		Subdomains: map[identity.Key]Subdomain{
			subdomain1Key: {
				Key:  subdomain1Key,
				Name: "Subdomain1",
				ClassAssociations: map[identity.Key]model_class.Association{
					sub1AssocKey: sub1Assoc,
				},
			},
			subdomain2Key: {
				Key:  subdomain2Key,
				Name: "Subdomain2",
				ClassAssociations: map[identity.Key]model_class.Association{
					sub2AssocKey: sub2Assoc,
				},
			},
		},
	}

	// Test: GetClassAssociations returns all associations.
	result := domain.GetClassAssociations()
	assert.Equal(suite.T(), 3, len(result))
	assert.Contains(suite.T(), result, domainAssocKey)
	assert.Contains(suite.T(), result, sub1AssocKey)
	assert.Contains(suite.T(), result, sub2AssocKey)

	// Test: returned map is a copy.
	class3InSub1 := helper.Must(identity.NewClassKey(subdomain1Key, "class3"))
	newAssocKey := helper.Must(identity.NewClassAssociationKey(subdomain1Key, class1InSub1, class3InSub1, "new association"))
	result[newAssocKey] = model_class.Association{Key: newAssocKey, Name: "New"}
	assert.Equal(suite.T(), 1, len(domain.ClassAssociations), "Domain associations should not be modified")
	assert.Equal(suite.T(), 1, len(domain.Subdomains[subdomain1Key].ClassAssociations), "Subdomain associations should not be modified")

	// Test: empty domain returns empty map.
	emptyDomain := Domain{
		Key:  domainKey,
		Name: "Empty Domain",
	}
	emptyResult := emptyDomain.GetClassAssociations()
	assert.NotNil(suite.T(), emptyResult)
	assert.Equal(suite.T(), 0, len(emptyResult))
}

// TestValidateWithParentDeepTree tests that key validation propagates through the full tree:
// domain → subdomain → class → guard/action/query logic keys.
func (suite *DomainSuite) TestValidateWithParentDeepTree() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	reqKey := helper.Must(identity.NewActionRequireKey(actionKey, "req_1"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "query1"))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "guar_1"))

	// Test valid full tree.
	guardLogic := helper.Must(model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Guard.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil))
	guard := helper.Must(model_state.NewGuard(guardKey, "Guard", guardLogic))
	reqLogic := helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Req.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil))
	action := helper.Must(model_state.NewAction(actionKey, "Action", "", []model_logic.Logic{reqLogic}, nil, nil, nil))
	guarLogic := helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Guar.", "result", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil))
	query := helper.Must(model_state.NewQuery(queryKey, "Query", "", nil, []model_logic.Logic{guarLogic}, nil))
	class := helper.Must(model_class.NewClass(classKey, "Class", "", nil, nil, nil, ""))
	class.SetGuards(map[identity.Key]model_state.Guard{guardKey: guard})
	class.SetActions(map[identity.Key]model_state.Action{actionKey: action})
	class.SetQueries(map[identity.Key]model_state.Query{queryKey: query})
	domain := Domain{
		Key:  domainKey,
		Name: "Domain",
		Subdomains: map[identity.Key]Subdomain{
			subdomainKey: {
				Key:  subdomainKey,
				Name: "Subdomain",
				Classes: map[identity.Key]model_class.Class{
					classKey: class,
				},
			},
		},
	}
	err := domain.ValidateWithParent(nil)
	assert.NoError(suite.T(), err, "Valid full tree should pass")

	// Test that a guard logic key mismatch deep in the tree is caught.
	otherGuardKey := helper.Must(identity.NewGuardKey(classKey, "other_guard"))
	mismatchedGuardLogic := helper.Must(model_logic.NewLogic(otherGuardKey, model_logic.LogicTypeAssessment, "Guard.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil))
	mismatchedGuard := helper.Must(model_state.NewGuard(guardKey, "Guard", mismatchedGuardLogic))
	mismatchedGuardClass := helper.Must(model_class.NewClass(classKey, "Class", "", nil, nil, nil, ""))
	mismatchedGuardClass.SetGuards(map[identity.Key]model_state.Guard{guardKey: mismatchedGuard})
	domain = Domain{
		Key:  domainKey,
		Name: "Domain",
		Subdomains: map[identity.Key]Subdomain{
			subdomainKey: {
				Key:  subdomainKey,
				Name: "Subdomain",
				Classes: map[identity.Key]model_class.Class{
					classKey: mismatchedGuardClass,
				},
			},
		},
	}
	err = domain.ValidateWithParent(nil)
	assert.ErrorContains(suite.T(), err, "does not match guard key", "Should catch guard logic key mismatch in deep tree")

	// Test that an action require key with wrong parent deep in the tree is caught.
	otherActionKey := helper.Must(identity.NewActionKey(classKey, "other_action"))
	wrongReqKey := helper.Must(identity.NewActionRequireKey(otherActionKey, "req_1"))
	wrongReqLogic := helper.Must(model_logic.NewLogic(wrongReqKey, model_logic.LogicTypeAssessment, "Req.", "", model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil))
	wrongReqAction := helper.Must(model_state.NewAction(actionKey, "Action", "", []model_logic.Logic{wrongReqLogic}, nil, nil, nil))
	wrongReqClass := helper.Must(model_class.NewClass(classKey, "Class", "", nil, nil, nil, ""))
	wrongReqClass.SetActions(map[identity.Key]model_state.Action{actionKey: wrongReqAction})
	domain = Domain{
		Key:  domainKey,
		Name: "Domain",
		Subdomains: map[identity.Key]Subdomain{
			subdomainKey: {
				Key:  subdomainKey,
				Name: "Subdomain",
				Classes: map[identity.Key]model_class.Class{
					classKey: wrongReqClass,
				},
			},
		},
	}
	err = domain.ValidateWithParent(nil)
	assert.ErrorContains(suite.T(), err, "requires 0", "Should catch action require key error in deep tree")
}

// TestValidateWithParentAndActorsAndClasses tests child validation propagation.
func (suite *DomainSuite) TestValidateWithParentAndActorsAndClasses() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	defaultSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "default"))
	subdomain1Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	subdomain2Key := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain2"))
	classKey1 := helper.Must(identity.NewClassKey(subdomain1Key, "class1"))
	classKey2 := helper.Must(identity.NewClassKey(subdomain2Key, "class2"))

	actors := map[identity.Key]bool{}
	classes := map[identity.Key]bool{
		classKey1: true,
		classKey2: true,
	}

	// Test invalid Subdomain child propagates error.
	domain := Domain{
		Key:  domainKey,
		Name: "Name",
		Subdomains: map[identity.Key]Subdomain{
			defaultSubdomainKey: {Key: defaultSubdomainKey, Name: ""}, // Invalid: blank name
		},
	}
	err := domain.ValidateWithParentAndActorsAndClasses(nil, actors, classes)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Subdomains")

	// Test invalid ClassAssociation child propagates error.
	// Domain-level associations require classes in different subdomains.
	assocKey := helper.Must(identity.NewClassAssociationKey(domainKey, classKey1, classKey2, "assoc1"))
	domain = Domain{
		Key:  domainKey,
		Name: "Name",
		ClassAssociations: map[identity.Key]model_class.Association{
			assocKey: {Key: assocKey, Name: ""}, // Invalid: blank name
		},
	}
	err = domain.ValidateWithParentAndActorsAndClasses(nil, actors, classes)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child ClassAssociations")

	// Test valid domain with single subdomain named "default".
	domain = Domain{
		Key:  domainKey,
		Name: "Name",
		Subdomains: map[identity.Key]Subdomain{
			defaultSubdomainKey: {Key: defaultSubdomainKey, Name: "Subdomain"},
		},
	}
	err = domain.ValidateWithParentAndActorsAndClasses(nil, actors, classes)
	assert.NoError(suite.T(), err, "Valid domain with single 'default' subdomain should pass")

	// Test single subdomain with non-"default" key fails.
	domain = Domain{
		Key:  domainKey,
		Name: "Name",
		Subdomains: map[identity.Key]Subdomain{
			subdomain1Key: {Key: subdomain1Key, Name: "Subdomain"},
		},
	}
	err = domain.ValidateWithParentAndActorsAndClasses(nil, actors, classes)
	assert.ErrorContains(suite.T(), err, "must be 'default'", "Single subdomain must have key 'default'")

	// Test multiple subdomains with "default" key fails.
	domain = Domain{
		Key:  domainKey,
		Name: "Name",
		Subdomains: map[identity.Key]Subdomain{
			defaultSubdomainKey: {Key: defaultSubdomainKey, Name: "Default Subdomain"},
			subdomain1Key:       {Key: subdomain1Key, Name: "Subdomain1"},
		},
	}
	err = domain.ValidateWithParentAndActorsAndClasses(nil, actors, classes)
	assert.ErrorContains(suite.T(), err, "reserved for single-subdomain", "Multiple subdomains cannot include 'default'")

	// Test multiple subdomains without "default" key passes.
	domain = Domain{
		Key:  domainKey,
		Name: "Name",
		Subdomains: map[identity.Key]Subdomain{
			subdomain1Key: {Key: subdomain1Key, Name: "Subdomain1"},
			subdomain2Key: {Key: subdomain2Key, Name: "Subdomain2"},
		},
	}
	err = domain.ValidateWithParentAndActorsAndClasses(nil, actors, classes)
	assert.NoError(suite.T(), err, "Multiple subdomains without 'default' should pass")
}
