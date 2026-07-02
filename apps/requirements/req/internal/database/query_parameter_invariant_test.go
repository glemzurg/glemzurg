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

func TestQueryParameterInvariantSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(QueryParameterInvariantSuite))
}

type QueryParameterInvariantSuite struct {
	suite.Suite
	db           *sql.DB
	model        core.Model
	domain       model_domain.Domain
	subdomain    model_domain.Subdomain
	class        model_class.Class
	query        model_state.Query
	logic        model_logic.Logic
	logicB       model_logic.Logic
	queryKey     identity.Key
	parameterKey identity.Key
	parameterSub string
	logicKey     identity.Key
	logicKeyB    identity.Key
}

func (suite *QueryParameterInvariantSuite) SetupTest() {
	suite.db = t_ResetDatabase(suite.T())

	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	classKey := helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key"))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, classKey)

	suite.queryKey = helper.Must(identity.NewQueryKey(suite.class.Key, "query_key"))
	suite.query = t_AddQuery(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.queryKey)
	param := helper.Must(model_state.NewParameter(suite.queryKey, "ProductID", "Nat", false))
	suite.parameterKey = param.Key
	suite.parameterSub = param.Key.SubKey
	err := AddQueryParameter(suite.db, suite.model.Key, suite.queryKey, param)
	suite.Require().NoError(err)

	suite.logicKey = helper.Must(identity.NewParameterInvariantKey(suite.parameterKey, "0"))
	suite.logicKeyB = helper.Must(identity.NewParameterInvariantKey(suite.parameterKey, "1"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *QueryParameterInvariantSuite) TestLoad() {
	_, err := LoadQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)

	err = dbExec(suite.db, `
		INSERT INTO query_parameter_invariant
			(model_key, query_key, parameter_key, logic_key)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/query/query_key',
				'productid',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/query/query_key/parameter/productid/pinvariant/0'
			)
	`)
	suite.Require().NoError(err)

	key, err := LoadQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *QueryParameterInvariantSuite) TestAdd() {
	err := AddQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().NoError(err)

	key, err := LoadQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().NoError(err)
	suite.Equal(suite.logicKey, key)
}

func (suite *QueryParameterInvariantSuite) TestRemove() {
	err := AddQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().NoError(err)

	err = RemoveQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().NoError(err)

	_, err = LoadQueryParameterInvariant(suite.db, suite.model.Key, suite.queryKey, suite.parameterSub, suite.logicKey)
	suite.Require().ErrorIs(err, ErrNotFound)
}

func (suite *QueryParameterInvariantSuite) TestQuery() {
	err := AddQueryParameterInvariants(suite.db, suite.model.Key, map[identity.Key]map[string][]identity.Key{
		suite.queryKey: {suite.parameterSub: {suite.logicKeyB, suite.logicKey}},
	})
	suite.Require().NoError(err)

	invariants, err := QueryQueryParameterInvariants(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[identity.Key][]identity.Key{
		suite.parameterKey: {suite.logicKey, suite.logicKeyB},
	}, invariants)
}
