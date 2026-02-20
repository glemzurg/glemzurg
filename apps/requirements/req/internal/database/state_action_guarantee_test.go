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
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActionGuaranteeSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActionGuaranteeSuite))
}

type ActionGuaranteeSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	action    model_state.Action
	logic     model_logic.Logic
	logicB    model_logic.Logic
	actionKey identity.Key
	logicKey  identity.Key
	logicKeyB identity.Key
}

func (suite *ActionGuaranteeSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.actionKey = helper.Must(identity.NewActionKey(suite.class.Key, "action_key"))
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.actionKey)

	// Create logic rows (action guarantee keys are children of action key).
	suite.logicKey = helper.Must(identity.NewActionGuaranteeKey(suite.actionKey, "guar_a"))
	suite.logicKeyB = helper.Must(identity.NewActionGuaranteeKey(suite.actionKey, "guar_b"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *ActionGuaranteeSuite) TestLoad() {

	// Logic row exists from SetupTest, but no action_guarantee join row yet.
	_, err := LoadActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)

	// Insert the action_guarantee join row.
	_, err = dbExec(suite.db, `
		INSERT INTO action_guarantee
			(model_key, action_key, logic_key)
		VALUES
			($1, $2, $3)
	`, suite.model.Key, suite.actionKey.String(), suite.logicKey.String())
	assert.Nil(suite.T(), err)

	key, err := LoadActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logicKey, key)
}

func (suite *ActionGuaranteeSuite) TestAdd() {

	err := AddActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	key, err := LoadActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logicKey, key)
}

func (suite *ActionGuaranteeSuite) TestRemove() {

	err := AddActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	err = RemoveActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	// Action guarantee should be gone.
	_, err = LoadActionGuarantee(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *ActionGuaranteeSuite) TestQuery() {

	err := AddActionGuarantees(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.actionKey: {suite.logicKeyB, suite.logicKey},
	})
	assert.Nil(suite.T(), err)

	guarantees, err := QueryActionGuarantees(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]identity.Key{
		suite.actionKey: {suite.logicKey, suite.logicKeyB},
	}, guarantees)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddActionGuarantee(t *testing.T, dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) identity.Key {

	err := AddActionGuarantee(dbOrTx, modelKey, actionKey, logicKey)
	assert.Nil(t, err)

	key, err := LoadActionGuarantee(dbOrTx, modelKey, actionKey, logicKey)
	assert.Nil(t, err)

	return key
}
