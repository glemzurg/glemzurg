package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
				Key:           validKey,
				Name:          "Name",
				Details:       "Details",
				Requires:      []string{"req1"},
				Guarantees:    []string{"guar1"},
				TlaRequires:   []string{"tla_req1"},
				TlaGuarantees: []string{"tla_guar1"},
			},
		},
		{
			testName: "valid query with tla requires only",
			query: Query{
				Key:         validKey,
				Name:        "Name",
				TlaRequires: []string{"x > 0"},
			},
		},
		{
			testName: "valid query with tla guarantees only",
			query: Query{
				Key:           validKey,
				Name:          "Name",
				TlaGuarantees: []string{"result \\in S"},
			},
		},
		{
			testName: "error empty key",
			query: Query{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
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
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error blank name with tla fields set",
			query: Query{
				Key:           validKey,
				Name:          "",
				TlaRequires:   []string{"x > 0"},
				TlaGuarantees: []string{"result \\in S"},
			},
			errstr: "Name: cannot be blank",
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

	// Test all parameters are mapped correctly.
	query, err := NewQuery(key, "Name", "Details",
		[]string{"Requires"}, []string{"Guarantees"},
		[]string{"tla_req"}, []string{"tla_guar"},
		nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Query{
		Key:           key,
		Name:          "Name",
		Details:       "Details",
		Requires:      []string{"Requires"},
		Guarantees:    []string{"Guarantees"},
		TlaRequires:   []string{"tla_req"},
		TlaGuarantees: []string{"tla_guar"},
	}, query)

	// Test with nil optional fields (all Tla* fields are optional).
	query, err = NewQuery(key, "Name", "Details",
		[]string{"Requires"}, []string{"Guarantees"},
		nil, nil, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Query{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"Requires"},
		Guarantees: []string{"Guarantees"},
	}, query)

	// Test that Validate is called (invalid data should fail).
	_, err = NewQuery(key, "", "Details", nil, nil, nil, nil, nil)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *QuerySuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewQueryKey(classKey, "query1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	query := Query{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := query.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

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
}
