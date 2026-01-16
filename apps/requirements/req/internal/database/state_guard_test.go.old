package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
}

func (suite *GuardSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
}

func (suite *GuardSuite) TestLoad() {

	// Nothing in database yet.
	classKey, guard, err := LoadGuard(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), guard)

	_, err = dbExec(suite.db, `
		INSERT INTO guard
			(
				model_key,
				class_key,
				guard_key,
				name,
				details
			)
		VALUES
			(
				'model_key',
				'class_key',
				'key',
				'Name',
				'Details'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, guard, err = LoadGuard(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Guard{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	}, guard)
}

func (suite *GuardSuite) TestAdd() {

	err := AddGuard(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Guard{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Guard{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	}, guard)
}

func (suite *GuardSuite) TestUpdate() {

	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = UpdateGuard(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Guard{
		Key:     "KeY", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
	})
	assert.Nil(suite.T(), err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Guard{
		Key:     "key",
		Name:    "NameX",
		Details: "DetailsX",
	}, guard)
}

func (suite *GuardSuite) TestRemove() {

	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = RemoveGuard(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	classKey, guard, err := LoadGuard(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), guard)
}

func (suite *GuardSuite) TestQuery() {

	err := AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:     "keyx",
		Name:    "NameX",
		Details: "DetailsX",
	})
	assert.Nil(suite.T(), err)

	err = AddGuard(suite.db, suite.model.Key, suite.class.Key, model_state.Guard{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	guards, err := QueryGuards(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_state.Guard{
		"class_key": []model_state.Guard{
			{
				Key:     "key",
				Name:    "Name",
				Details: "Details",
			},
			{
				Key:     "keyx",
				Name:    "NameX",
				Details: "DetailsX",
			},
		},
	}, guards)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddGuard(t *testing.T, dbOrTx DbOrTx, modelKey, classKey, guardKey string) (guard model_state.Guard) {

	err := AddGuard(dbOrTx, modelKey, classKey, model_state.Guard{
		Key:     guardKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(t, err)

	_, guard, err = LoadGuard(dbOrTx, modelKey, guardKey)
	assert.Nil(t, err)

	return guard
}
