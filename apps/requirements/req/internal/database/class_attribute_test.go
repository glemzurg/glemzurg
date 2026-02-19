package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

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
	db            *sql.DB
	model         req_model.Model
	domain        model_domain.Domain
	subdomain     model_domain.Subdomain
	class         model_class.Class
	logic         model_logic.Logic
	logicB        model_logic.Logic
	attributeKey  identity.Key
	attributeKeyB identity.Key
}

func (suite *AttributeSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the attribute keys for reuse.
	suite.attributeKey = helper.Must(identity.NewAttributeKey(suite.class.Key, "key"))
	suite.attributeKeyB = helper.Must(identity.NewAttributeKey(suite.class.Key, "key_b"))

	// Create logic rows for derivation policies (logic must exist before attribute references it).
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewAttributeDerivationKey(suite.attributeKey, "deriv")))
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewAttributeDerivationKey(suite.attributeKeyB, "deriv")))
}

func (suite *AttributeSuite) TestLoad() {

	// Nothing in database yet.
	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
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
				derivation_policy_key,
				nullable,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/attribute/key',
				'Name',
				'Details',
				'DataTypeRules',
				$1,
				true,
				'UmlComment'
			)
	`, suite.logic.Key.String())
	assert.Nil(suite.T(), err)

	classKey, attribute, err = LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
		Nullable:         true,
		UmlComment:       "UmlComment",
	}, attribute)
}

func (suite *AttributeSuite) TestAdd() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
		Nullable:         true,
		UmlComment:       "UmlComment",
	}, attribute)
}

func (suite *AttributeSuite) TestAddNulls() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: nil, // No derivation policy.
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: nil, // No derivation policy.
		Nullable:         true,
		UmlComment:       "UmlComment",
	}, attribute)
}

func (suite *AttributeSuite) TestUpdate() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: &model_logic.Logic{Key: suite.logicB.Key},
		Nullable:         false,
		UmlComment:       "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: &model_logic.Logic{Key: suite.logicB.Key},
		Nullable:         false,
		UmlComment:       "UmlCommentX",
	}, attribute)
}

func (suite *AttributeSuite) TestUpdateNulls() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: nil, // No derivation policy.
		Nullable:         false,
		UmlComment:       "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "NameX",
		Details:          "DetailsX",
		DataTypeRules:    "DataTypeRulesX",
		DerivationPolicy: nil, // No derivation policy.
		Nullable:         false,
		UmlComment:       "UmlCommentX",
	}, attribute)
}

func (suite *AttributeSuite) TestRemove() {

	err := AddAttribute(suite.db, suite.model.Key, suite.class.Key, model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveAttribute(suite.db, suite.model.Key, suite.class.Key, suite.attributeKey)
	assert.Nil(suite.T(), err)

	classKey, attribute, err := LoadAttribute(suite.db, suite.model.Key, suite.attributeKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), attribute)
}

func (suite *AttributeSuite) TestQuery() {

	err := AddAttributes(suite.db, suite.model.Key, map[identity.Key][]model_class.Attribute{
		suite.class.Key: {
			{
				Key:              suite.attributeKeyB,
				Name:             "NameX",
				Details:          "DetailsX",
				DataTypeRules:    "DataTypeRulesX",
				DerivationPolicy: &model_logic.Logic{Key: suite.logicB.Key},
				Nullable:         true,
				UmlComment:       "UmlCommentX",
			},
			{
				Key:              suite.attributeKey,
				Name:             "Name",
				Details:          "Details",
				DataTypeRules:    "DataTypeRules",
				DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
				Nullable:         true,
				UmlComment:       "UmlComment",
			},
		},
	})
	assert.Nil(suite.T(), err)

	attributes, err := QueryAttributes(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_class.Attribute{
		suite.class.Key: {
			{
				Key:              suite.attributeKey,
				Name:             "Name",
				Details:          "Details",
				DataTypeRules:    "DataTypeRules",
				DerivationPolicy: &model_logic.Logic{Key: suite.logic.Key},
				Nullable:         true,
				UmlComment:       "UmlComment",
			},
			{
				Key:              suite.attributeKeyB,
				Name:             "NameX",
				Details:          "DetailsX",
				DataTypeRules:    "DataTypeRulesX",
				DerivationPolicy: &model_logic.Logic{Key: suite.logicB.Key},
				Nullable:         true,
				UmlComment:       "UmlCommentX",
			},
		},
	}, attributes)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddAttribute(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, attributeKey identity.Key) (attribute model_class.Attribute) {

	err := AddAttribute(dbOrTx, modelKey, classKey, model_class.Attribute{
		Key:              attributeKey,
		Name:             attributeKey.String(),
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: nil, // No derivation policy.
		Nullable:         true,
		UmlComment:       "UmlComment",
	})
	assert.Nil(t, err)

	_, attribute, err = LoadAttribute(dbOrTx, modelKey, attributeKey)
	assert.Nil(t, err)

	return attribute
}

func (suite *AttributeSuite) TestVerifyTestObjects() {

	attribute := t_AddAttribute(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.attributeKey)
	assert.Equal(suite.T(), model_class.Attribute{
		Key:              suite.attributeKey,
		Name:             suite.attributeKey.String(),
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: nil, // No derivation policy.
		Nullable:         true,
		UmlComment:       "UmlComment",
	}, attribute)
}
