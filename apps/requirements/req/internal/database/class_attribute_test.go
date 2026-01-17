package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAttributeSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AttributeSuite))
}

type AttributeSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
}

func (suite *AttributeSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
}

func (suite *AttributeSuite) TestLoad() {

	// Nothing in database yet.
	classKey, attribute, err := LoadAttribute(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), attribute)

	_, err = dbExec(suite.db, `
		INSERT INTO attribute
			(
				model_key,
				class_key,
				attribute_key,
				name,
				details,
				data_type_rules,
				derivation_policy,
				nullable,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'class_key',
				'key',
				'Name',
				'Details',
				'DataTypeRules',
				'DerivationPolicy',
				true,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, attribute, err = LoadAttribute(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              "key", // Test case-insensitive.
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	}, attribute)
}

func (suite *AttributeSuite) TestAdd() {

	err := AddAttribute(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_class.Attribute{
		Key:              "KeY", // Test case-insensitive.
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              "key",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	}, attribute)
}

func (suite *AttributeSuite) TestUpdate() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              "key",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAttribute(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_class.Attribute{
		Key:              "KeY", // Test case-insensitive.
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: "DerivationPolicyX",
		Nullable:         false,
		UmlComment:       "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              "key",
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: "DerivationPolicyX",
		Nullable:         false,
		UmlComment:       "UmlCommentX",
	}, attribute)
}

func (suite *AttributeSuite) TestRemove() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              "key",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveAttribute(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), attribute)
}

func (suite *AttributeSuite) TestQuery() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              "keyx",
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: "DerivationPolicyX",
		Nullable:         true,
		UmlComment:       "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              "key",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	attributes, err := QueryAttributes(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_class.Attribute{
		"class_key": []model_class.Attribute{
			{
				Key:              "key",
				Name:             "Name",
				Details:          "Details",
				DataTypeRules:    "DataTypeRules",
				DerivationPolicy: "DerivationPolicy",
				Nullable:         true,
				UmlComment:       "UmlComment",
			},
			{
				Key:              "keyx",
				Name:             "NameX",
				Details:          "DetailsX",
				DataTypeRules:    "DataTypeRulesX",
				DerivationPolicy: "DerivationPolicyX",
				Nullable:         true,
				UmlComment:       "UmlCommentX",
			},
		},
	}, attributes)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddAttribute(t *testing.T, dbOrTx DbOrTx, modelKey, classKey, attributeKey string) (attribute model_class.Attribute) {

	err := AddAttribute(dbOrTx, modelKey, classKey, model_class.Attribute{
		Key:              attributeKey,
		Name:             attributeKey,
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: "DerivationPolicy",
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(t, err)

	_, attribute, err = LoadAttribute(dbOrTx, modelKey, attributeKey)
	assert.Nil(t, err)

	return attribute
}
