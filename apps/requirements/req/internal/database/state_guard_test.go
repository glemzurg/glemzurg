package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGuardSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(GuardSuite))
}

type GuardSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	guardKey  identity.Key
	guardKeyB identity.Key
}

func (suite *GuardSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the guard keys for reuse.
	suite.guardKey = helper.Must(identity.NewGuardKey(suite.class.Key, "key"))
	suite.guardKeyB = helper.Must(identity.NewGuardKey(suite.class.Key, "key_b"))

	// Logic rows must exist before guards can be inserted (FK constraint).
	t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.guardKey)
	t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.guardKeyB)
}

func (suite *GuardSuite) TestLoad() {

	// Nothing in database yet.
	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), guard)

	_, err = dbExec(suite.db, `
		INSERT INTO guard
			(
				model_key,
				class_key,
				guard_key,
				name
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/guard/key',
				'Name'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, guard, err = LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	}, guard)
}

func (suite *GuardSuite) TestAdd() {

	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	})
	assert.Nil(suite.T(), err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	}, guard)
}

func (suite *GuardSuite) TestUpdate() {

	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	})
	assert.Nil(suite.T(), err)

	err = UpdateGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "NameX",
	})
	assert.Nil(suite.T(), err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Guard{
		Key:  suite.guardKey,
		Name: "NameX",
	}, guard)
}

func (suite *GuardSuite) TestRemove() {

	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	})
	assert.Nil(suite.T(), err)

	err = RemoveGuard(suite.db, suite.model.Key, suite.class.Key, suite.guardKey)
	assert.Nil(suite.T(), err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), guard)
}

func (suite *GuardSuite) TestQuery() {

	err := AddGuards(suite.db, suite.model.Key, map[identity.Key][]model_state.Guard{
		suite.class.Key: {
			{
				Key:  suite.guardKeyB,
				Name: "NameX",
			},
			{
				Key:  suite.guardKey,
				Name: "Name",
			},
		},
	})
	assert.Nil(suite.T(), err)

	guards, err := QueryGuards(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Guard{
		suite.class.Key: {
			{
				Key:  suite.guardKey,
				Name: "Name",
			},
			{
				Key:  suite.guardKeyB,
				Name: "NameX",
			},
		},
	}, guards)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddGuard(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, guardKey identity.Key) (guard model_state.Guard) {

	// Logic row must exist before guard (FK constraint).
	t_AddLogic(t, dbOrTx, modelKey, guardKey)

	err := AddGuard(dbOrTx, modelKey, classKey, model_state.Guard{
		Key:  guardKey,
		Name: guardKey.String(),
	})
	assert.Nil(t, err)

	_, guard, err = LoadGuard(dbOrTx, modelKey, guardKey)
	assert.Nil(t, err)

	return guard
}
