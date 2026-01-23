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
	stateKey  identity.Key
	stateKeyB identity.Key
}

func (suite *StateSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the state keys for reuse.
	suite.stateKey = helper.Must(identity.NewStateKey(suite.class.Key, "key"))
	suite.stateKeyB = helper.Must(identity.NewStateKey(suite.class.Key, "key_b"))
}

func (suite *StateSuite) TestLoad() {

	// Nothing in database yet.
	classKey, state, err := LoadState(suite.db, suite.model.Key, suite.stateKey)
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
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/state/key',
				'Name',
				'Details',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, state, err = LoadState(suite.db, suite.model.Key, suite.stateKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.State{
		Key:        suite.stateKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, state)
}

func (suite *StateSuite) TestAdd() {

	err := AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        suite.stateKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, state, err := LoadState(suite.db, suite.model.Key, suite.stateKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.State{
		Key:        suite.stateKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, state)
}

func (suite *StateSuite) TestUpdate() {

	err := AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        suite.stateKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        suite.stateKey,
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, state, err := LoadState(suite.db, suite.model.Key, suite.stateKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.State{
		Key:        suite.stateKey,
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	}, state)
}

func (suite *StateSuite) TestRemove() {

	err := AddState(suite.db, suite.model.Key, suite.class.Key, model_state.State{
		Key:        suite.stateKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveState(suite.db, suite.model.Key, suite.class.Key, suite.stateKey)
	assert.Nil(suite.T(), err)

	classKey, state, err := LoadState(suite.db, suite.model.Key, suite.stateKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), state)
}

func (suite *StateSuite) TestQuery() {

	err := AddStates(suite.db, suite.model.Key, map[identity.Key][]model_state.State{
		suite.class.Key: {
			{
				Key:        suite.stateKeyB,
				Name:       "NameX",
				Details:    "DetailsX",
				UmlComment: "UmlCommentX",
			},
			{
				Key:        suite.stateKey,
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
	})
	assert.Nil(suite.T(), err)

	states, err := QueryStates(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.State{
		suite.class.Key: {
			{
				Key:        suite.stateKey,
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
			{
				Key:        suite.stateKeyB,
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

func t_AddState(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, stateKey identity.Key) (state model_state.State) {

	err := AddState(dbOrTx, modelKey, classKey, model_state.State{
		Key:        stateKey,
		Name:       stateKey.String(),
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, state, err = LoadState(dbOrTx, modelKey, stateKey)
	assert.Nil(t, err)

	return state
}
