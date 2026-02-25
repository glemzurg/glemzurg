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

func TestClassInvariantSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ClassInvariantSuite))
}

type ClassInvariantSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	logic     model_logic.Logic
	logicB    model_logic.Logic
	classKey  identity.Key
	logicKey  identity.Key
	logicKeyB identity.Key
}

func (suite *ClassInvariantSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.classKey = helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key"))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, suite.classKey)

	// Create logic rows (class invariant keys are children of class key).
	suite.logicKey = helper.Must(identity.NewClassInvariantKey(suite.classKey, "0"))
	suite.logicKeyB = helper.Must(identity.NewClassInvariantKey(suite.classKey, "1"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *ClassInvariantSuite) TestLoad() {

	// Logic row exists from SetupTest, but no class_invariant join row yet.
	_, err := LoadClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)

	// Insert the class_invariant join row.
	_, err = dbExec(suite.db, `
		INSERT INTO class_invariant
			(model_key, class_key, logic_key)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/cinvariant/0'
			)
	`)
	assert.Nil(suite.T(), err)

	key, err := LoadClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logicKey, key)
}

func (suite *ClassInvariantSuite) TestAdd() {

	err := AddClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	key, err := LoadClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logicKey, key)
}

func (suite *ClassInvariantSuite) TestRemove() {

	err := AddClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	err = RemoveClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	// Class invariant should be gone.
	_, err = LoadClassInvariant(suite.db, suite.model.Key, suite.classKey, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *ClassInvariantSuite) TestQuery() {

	err := AddClassInvariants(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.classKey: {suite.logicKeyB, suite.logicKey},
	})
	assert.Nil(suite.T(), err)

	invariants, err := QueryClassInvariants(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]identity.Key{
		suite.classKey: {suite.logicKey, suite.logicKeyB},
	}, invariants)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddClassInvariant(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, logicKey identity.Key) identity.Key {

	err := AddClassInvariant(dbOrTx, modelKey, classKey, logicKey)
	assert.Nil(t, err)

	key, err := LoadClassInvariant(dbOrTx, modelKey, classKey, logicKey)
	assert.Nil(t, err)

	return key
}
