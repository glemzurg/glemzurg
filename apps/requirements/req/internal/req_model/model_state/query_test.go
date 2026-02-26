package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestQuerySuite(t *testing.T) {
	suite.Run(t, new(QuerySuite))
}

type QuerySuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Query.
func (suite *QuerySuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewQueryKey(classKey, "query1"))
	reqKey := helper.Must(identity.NewQueryRequireKey(validKey, "req_1"))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(validKey, "guar_1"))

	tests := []struct {
		testName string
		query    Query
		errstr   string
	}{
		{
			testName: "valid query minimal",
			query: Query{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "valid query with all optional fields",
			query: Query{
				Key:     validKey,
				Name:    "Name",
				Details: "Details",
				Requires: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Precondition 1.", "", model_logic.NotationTLAPlus, "req1")),
				},
				Guarantees: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Guarantee 1.", "result", model_logic.NotationTLAPlus, "guar1")),
				},
			},
		},
		{
			testName: "valid query with requires only",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "x must be positive.", "", model_logic.NotationTLAPlus, "x > 0")),
				},
			},
		},
		{
			testName: "valid query with guarantees only",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Result in S.", "result", model_logic.NotationTLAPlus, "result \\in S")),
				},
			},
		},
		{
			testName: "error empty key",
			query: Query{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			query: Query{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for query",
		},
		{
			testName: "error blank name",
			query: Query{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
		{
			testName: "error blank name with logic fields set",
			query: Query{
				Key:  validKey,
				Name: "",
				Requires: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "x must be positive.", "", model_logic.NotationTLAPlus, "x > 0")),
				},
				Guarantees: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Result in S.", "result", model_logic.NotationTLAPlus, "result \\in S")),
				},
			},
			errstr: "Name",
		},
		{
			testName: "error invalid requires logic missing key",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					{Key: identity.Key{}, Type: model_logic.LogicTypeAssessment, Description: "x must be positive.", Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "requires 0",
		},
		{
			testName: "error invalid guarantee logic missing key",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					{Key: identity.Key{}, Type: model_logic.LogicTypeQuery, Description: "Result in S.", Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "guarantee 0",
		},
		{
			testName: "error requires wrong kind",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Requires: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeStateChange, "x must be positive.", "x", model_logic.NotationTLAPlus, "")),
				},
			},
			errstr: "requires 0: logic kind must be 'assessment'",
		},
		{
			testName: "error guarantee wrong kind",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeAssessment, "Result in S.", "", model_logic.NotationTLAPlus, "")),
				},
			},
			errstr: "guarantee 0: logic kind must be 'query'",
		},
		{
			testName: "error duplicate guarantee target",
			query: Query{
				Key:  validKey,
				Name: "Name",
				Guarantees: []model_logic.Logic{
					helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Result 1.", "result", model_logic.NotationTLAPlus, "expr1")),
					helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Result 2.", "result", model_logic.NotationTLAPlus, "expr2")),
				},
			},
			errstr: "duplicate target",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewQuery maps parameters correctly and calls Validate.
func (suite *QuerySuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewQueryKey(classKey, "query1"))
	reqKey := helper.Must(identity.NewQueryRequireKey(key, "req_1"))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(key, "guar_1"))

	requires := []model_logic.Logic{
		helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "tla_req")),
	}
	guarantees := []model_logic.Logic{
		helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Guarantee.", "result", model_logic.NotationTLAPlus, "tla_guar")),
	}

	// Test all parameters are mapped correctly.
	params := []Parameter{
		helper.Must(NewParameter("ParamA", "Nat")),
		helper.Must(NewParameter("ParamB", "Int")),
	}
	query, err := NewQuery(key, "Name", "Details",
		requires, guarantees, params)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Query{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Requires:   requires,
		Guarantees: guarantees,
		Parameters: []Parameter{
			helper.Must(NewParameter("ParamA", "Nat")),
			helper.Must(NewParameter("ParamB", "Int")),
		},
	}, query)

	// Test with nil optional fields (all Logic slice fields are optional).
	query, err = NewQuery(key, "Name", "Details",
		nil, nil, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Query{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, query)

	// Test that Validate is called (invalid data should fail).
	_, err = NewQuery(key, "", "Details", nil, nil, nil)
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *QuerySuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewQueryKey(classKey, "query1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))
	reqKey := helper.Must(identity.NewQueryRequireKey(validKey, "req_1"))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(validKey, "guar_1"))

	// Test that Validate is called.
	query := Query{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := query.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - query key has class1 as parent, but we pass other_class.
	query = Query{
		Key:  validKey,
		Name: "Name",
	}
	err = query.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = query.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)

	// Test valid with logic children.
	query = Query{
		Key:  validKey,
		Name: "Name",
		Requires: []model_logic.Logic{
			helper.Must(model_logic.NewLogic(reqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "")),
		},
		Guarantees: []model_logic.Logic{
			helper.Must(model_logic.NewLogic(guarKey, model_logic.LogicTypeQuery, "Guarantee.", "result", model_logic.NotationTLAPlus, "")),
		},
	}
	err = query.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)

	// Test logic key validation - require with wrong parent should fail.
	otherQueryKey := helper.Must(identity.NewQueryKey(classKey, "other_query"))
	wrongReqKey := helper.Must(identity.NewQueryRequireKey(otherQueryKey, "req_1"))
	query = Query{
		Key:  validKey,
		Name: "Name",
		Requires: []model_logic.Logic{
			helper.Must(model_logic.NewLogic(wrongReqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "")),
		},
	}
	err = query.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "requires 0", "ValidateWithParent should validate logic key parent")

	// Test child Parameter validation propagates error.
	query = Query{
		Key:  validKey,
		Name: "Name",
		Parameters: []Parameter{
			{Name: "", DataTypeRules: "Nat"}, // Invalid: blank name
		},
	}
	err = query.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should validate child Parameters")

	// Test valid with child Parameters.
	query = Query{
		Key:  validKey,
		Name: "Name",
		Parameters: []Parameter{
			helper.Must(NewParameter("param1", "Nat")),
		},
	}
	err = query.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
