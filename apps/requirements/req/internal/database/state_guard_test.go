package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/require"
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
	model     core.Model
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
	suite.ErrorIs(err, ErrNotFound)
	suite.Empty(classKey)
	suite.Empty(guard)

	err = dbExec(suite.db, `
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
	suite.Require().NoError(err)

	classKey, guard, err = LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	suite.Require().NoError(err)
	suite.Equal(suite.class.Key, classKey)
	suite.Equal(model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	}, guard)
}

func (suite *GuardSuite) TestAdd() {
	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	})
	suite.Require().NoError(err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	suite.Require().NoError(err)
	suite.Equal(suite.class.Key, classKey)
	suite.Equal(model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	}, guard)
}

func (suite *GuardSuite) TestUpdate() {
	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	})
	suite.Require().NoError(err)

	err = UpdateGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "NameX",
	})
	suite.Require().NoError(err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	suite.Require().NoError(err)
	suite.Equal(suite.class.Key, classKey)
	suite.Equal(model_state.Guard{
		Key:  suite.guardKey,
		Name: "NameX",
	}, guard)
}

func (suite *GuardSuite) TestRemove() {
	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:  suite.guardKey,
		Name: "Name",
	})
	suite.Require().NoError(err)

	err = RemoveGuard(suite.db, suite.model.Key, suite.class.Key, suite.guardKey)
	suite.Require().NoError(err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, suite.guardKey)
	suite.ErrorIs(err, ErrNotFound)
	suite.Empty(classKey)
	suite.Empty(guard)
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
	suite.Require().NoError(err)

	guards, err := QueryGuards(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[identity.Key][]model_state.Guard{
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
	require.NoError(t, err)

	_, guard, err = LoadGuard(dbOrTx, modelKey, guardKey)
	require.NoError(t, err)

	return guard
}
