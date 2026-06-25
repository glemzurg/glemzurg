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

func TestParameterInvariantSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ParameterInvariantSuite))
}

type ParameterInvariantSuite struct {
	suite.Suite
	db           *sql.DB
	model        core.Model
	domain       model_domain.Domain
	subdomain    model_domain.Subdomain
	class        model_class.Class
	action       model_state.Action
	logic        model_logic.Logic
	logicB       model_logic.Logic
	parameterKey identity.Key
	logicKey     identity.Key
	logicKeyB    identity.Key
}

func (suite *ParameterInvariantSuite) SetupTest() {
	suite.db = t_ResetDatabase(suite.T())

	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	classKey := helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key"))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, classKey)

	actionKey := helper.Must(identity.NewActionKey(suite.class.Key, "action_key"))
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, actionKey)
	suite.parameterKey = helper.Must(identity.NewParameterKey(actionKey, "amount"))
	param := helper.Must(model_state.NewParameter(actionKey, "Amount", "Nat", false))
	err := AddActionParameter(suite.db, suite.model.Key, actionKey, param)
	suite.Require().NoError(err)

	suite.logicKey = helper.Must(identity.NewParameterInvariantKey(suite.parameterKey, "0"))
	suite.logicKeyB = helper.Must(identity.NewParameterInvariantKey(suite.parameterKey, "1"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *ParameterInvariantSuite) TestLoad() {
	_, err := LoadParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)

	err = dbExec(suite.db, `
		INSERT INTO parameter_invariant
			(model_key, parameter_key, logic_key)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key/parameter/amount',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key/parameter/amount/pinvariant/0'
			)
	`)
	suite.Require().NoError(err)

	key, err := LoadParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *ParameterInvariantSuite) TestAdd() {
	err := AddParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().NoError(err)

	key, err := LoadParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *ParameterInvariantSuite) TestRemove() {
	err := AddParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().NoError(err)

	err = RemoveParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().NoError(err)

	_, err = LoadParameterInvariant(suite.db, suite.model.Key, suite.parameterKey, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)
}

func (suite *ParameterInvariantSuite) TestQuery() {
	err := AddParameterInvariants(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.parameterKey: {suite.logicKeyB, suite.logicKey},
	})
	suite.Require().NoError(err)

	invariants, err := QueryParameterInvariants(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[identity.Key][]identity.Key{
		suite.parameterKey: {suite.logicKey, suite.logicKeyB},
	}, invariants)
}
