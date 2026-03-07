package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/suite"
)

func TestAttributeInvariantSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AttributeInvariantSuite))
}

type AttributeInvariantSuite struct {
	suite.Suite
	db           *sql.DB
	model        core.Model
	domain       model_domain.Domain
	subdomain    model_domain.Subdomain
	class        model_class.Class
	logic        model_logic.Logic
	logicB       model_logic.Logic
	attributeKey identity.Key
	logicKey     identity.Key
	logicKeyB    identity.Key
}

func (suite *AttributeInvariantSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	classKey := helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key"))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, classKey)

	// Create the attribute key (attribute row must exist for FK).
	suite.attributeKey = helper.Must(identity.NewAttributeKey(suite.class.Key, "attr_key"))
	// Insert a minimal attribute row so the FK is satisfied.
	t_AddAttribute(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.attributeKey)

	// Create logic rows (attribute invariant keys are children of attribute key).
	suite.logicKey = helper.Must(identity.NewAttributeInvariantKey(suite.attributeKey, "0"))
	suite.logicKeyB = helper.Must(identity.NewAttributeInvariantKey(suite.attributeKey, "1"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *AttributeInvariantSuite) TestLoad() {
	// Logic row exists from SetupTest, but no attribute_invariant join row yet.
	_, err := LoadAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)

	// Insert the attribute_invariant join row.
	err = dbExec(suite.db, `
		INSERT INTO attribute_invariant
			(model_key, attribute_key, logic_key)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/attribute/attr_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/attribute/attr_key/ainvariant/0'
			)
	`)
	suite.Require().NoError(err)

	key, err := LoadAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *AttributeInvariantSuite) TestAdd() {
	err := AddAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().NoError(err)

	key, err := LoadAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *AttributeInvariantSuite) TestRemove() {
	err := AddAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().NoError(err)

	err = RemoveAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().NoError(err)

	// Attribute invariant should be gone.
	_, err = LoadAttributeInvariant(suite.db, suite.model.Key, suite.attributeKey, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)
}

func (suite *AttributeInvariantSuite) TestQuery() {
	err := AddAttributeInvariants(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.attributeKey: {suite.logicKeyB, suite.logicKey},
	})
	suite.Require().NoError(err)

	invariants, err := QueryAttributeInvariants(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[identity.Key][]identity.Key{
		suite.attributeKey: {suite.logicKey, suite.logicKeyB},
	}, invariants)
}

//==================================================
