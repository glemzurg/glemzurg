package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/suite"
)

func TestActionSafetySuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActionSafetySuite))
}

type ActionSafetySuite struct {
	suite.Suite
	db        *sql.DB
	model     core.Model
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

func (suite *ActionSafetySuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.actionKey = helper.Must(identity.NewActionKey(suite.class.Key, "action_key"))
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.actionKey)

	// Create logic rows (action safety keys are children of action key).
	suite.logicKey = helper.Must(identity.NewActionSafetyKey(suite.actionKey, "safety_a"))
	suite.logicKeyB = helper.Must(identity.NewActionSafetyKey(suite.actionKey, "safety_b"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *ActionSafetySuite) TestLoad() {
	// Logic row exists from SetupTest, but no action_safety join row yet.
	_, err := LoadActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.ErrorIs(err, ErrNotFound)

	// Insert the action_safety join row.
	err = dbExec(suite.db, `
		INSERT INTO action_safety
			(model_key, action_key, logic_key)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key/asafety/safety_a'
			)
	`)
	suite.Require().NoError(err)

	key, err := LoadActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *ActionSafetySuite) TestAdd() {
	err := AddActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.Require().NoError(err)

	key, err := LoadActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *ActionSafetySuite) TestRemove() {
	err := AddActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.Require().NoError(err)

	err = RemoveActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.Require().NoError(err)

	// Action safety should be gone.
	_, err = LoadActionSafety(suite.db, suite.model.Key, suite.actionKey, suite.logicKey)
	suite.ErrorIs(err, ErrNotFound)
}

func (suite *ActionSafetySuite) TestQuery() {
	err := AddActionSafeties(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.actionKey: {suite.logicKeyB, suite.logicKey},
	})
	suite.Require().NoError(err)

	safeties, err := QueryActionSafeties(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[identity.Key][]identity.Key{
		suite.actionKey: {suite.logicKey, suite.logicKeyB},
	}, safeties)
}
