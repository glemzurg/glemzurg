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

func TestStateSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(StateSuite))
}

type StateSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
}

func (suite *StateSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
}

func (suite *StateSuite) TestLoad() {

	// Nothing in database yet.
	classKey, state, err := LoadState(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), state)

	_, err = dbExec(suite.db, `
		INSERT INTO state
			(
				model_key,
				class_key,
				state_key,
				name,
				details,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'class_key',
				'key',
				'Name',
				'Details',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, state, err = LoadState(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.State{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, state)
}

func (suite *StateSuite) TestAdd() {

	err := AddState(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.State{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, state, err := LoadState(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.State{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, state)
}

func (suite *StateSuite) TestUpdate() {

	err := AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateState(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.State{
		Key:        "KeY", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, state, err := LoadState(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.State{
		Key:        "key",
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	}, state)
}

func (suite *StateSuite) TestRemove() {

	err := AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveState(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	classKey, state, err := LoadState(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), state)
}

func (suite *StateSuite) TestQuery() {

	err := AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	states, err := QueryStates(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_state.State{
		"class_key": []model_state.State{
			{
				Key:        "key",
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
			{
				Key:        "keyx",
				Name:       "NameX",
				Details:    "DetailsX",
				UmlComment: "UmlCommentX",
			},
		},
	}, states)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddState(t *testing.T, dbOrTx DbOrTx, modelKey, classKey, stateKey string) (state model_state.State) {

	err := AddState(dbOrTx, modelKey, classKey, model_state.State{
		Key:        stateKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, state, err = LoadState(dbOrTx, modelKey, stateKey)
	assert.Nil(t, err)

	return state
}
