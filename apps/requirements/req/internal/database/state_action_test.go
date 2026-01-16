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

func TestActionSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
}

func (suite *ActionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
}

func (suite *ActionSuite) TestLoad() {

	// Nothing in database yet.
	classKey, action, err := LoadAction(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), action)

	_, err = dbExec(suite.db, `
		INSERT INTO action
			(
				model_key,
				class_key,
				action_key,
				name,
				details,
				requires,
				guarantees
			)
		VALUES
			(
				'model_key',
				'class_key',
				'key',
				'Name',
				'Details',
				'{"RequiresA","RequiresB"}',
				'{"GuaranteesA","GuaranteesB"}'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, action, err = LoadAction(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Action{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	}, action)
}

func (suite *ActionSuite) TestAdd() {

	err := AddAction(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Action{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	})
	assert.Nil(suite.T(), err)

	classKey, action, err := LoadAction(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Action{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	}, action)
}

func (suite *ActionSuite) TestUpdate() {

	err := AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	})
	assert.Nil(suite.T(), err)

	err = UpdateAction(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Action{
		Key:        "KeY", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Requires:   []string{"RequiresAX", "RequiresBX"},
		Guarantees: []string{"GuaranteesAX", "GuaranteesBX"},
	})
	assert.Nil(suite.T(), err)

	classKey, action, err := LoadAction(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Action{
		Key:        "key",
		Name:       "NameX",
		Details:    "DetailsX",
		Requires:   []string{"RequiresAX", "RequiresBX"},
		Guarantees: []string{"GuaranteesAX", "GuaranteesBX"},
	}, action)
}

func (suite *ActionSuite) TestRemove() {

	err := AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	})
	assert.Nil(suite.T(), err)

	err = RemoveAction(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	classKey, action, err := LoadAction(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), action)
}

func (suite *ActionSuite) TestQuery() {

	err := AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		Requires:   []string{"RequiresAX", "RequiresBX"},
		Guarantees: []string{"GuaranteesAX", "GuaranteesBX"},
	})
	assert.Nil(suite.T(), err)

	err = AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	})
	assert.Nil(suite.T(), err)

	actions, err := QueryActions(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_state.Action{
		"class_key": []model_state.Action{
			{
				Key:        "key",
				Name:       "Name",
				Details:    "Details",
				Requires:   []string{"RequiresA", "RequiresB"},
				Guarantees: []string{"GuaranteesA", "GuaranteesB"},
			},
			{
				Key:        "keyx",
				Name:       "NameX",
				Details:    "DetailsX",
				Requires:   []string{"RequiresAX", "RequiresBX"},
				Guarantees: []string{"GuaranteesAX", "GuaranteesBX"},
			},
		},
	}, actions)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddAction(t *testing.T, dbOrTx DbOrTx, modelKey, classKey, actionKey string) (action model_state.Action) {

	err := AddAction(dbOrTx, modelKey, classKey, model_state.Action{
		Key:        actionKey,
		Name:       "Name",
		Details:    "Details",
		Requires:   []string{"RequiresA", "RequiresB"},
		Guarantees: []string{"GuaranteesA", "GuaranteesB"},
	})
	assert.Nil(t, err)

	_, action, err = LoadAction(dbOrTx, modelKey, actionKey)
	assert.Nil(t, err)

	return action
}
