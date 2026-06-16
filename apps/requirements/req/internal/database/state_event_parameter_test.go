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

func TestEventParameterSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(EventParameterSuite))
}

type EventParameterSuite struct {
	suite.Suite
	db        *sql.DB
	model     core.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	event     model_state.Event
	eventKey  identity.Key
}

func (suite *EventParameterSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the event key for reuse.
	suite.eventKey = helper.Must(identity.NewEventKey(suite.class.Key, "event_key"))
	suite.event = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.eventKey)
}

func (suite *EventParameterSuite) TestLoad() {
	// Nothing in database yet.
	name, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(name)

	err = dbExec(suite.db, `
		INSERT INTO event_parameter
			(
				model_key,
				event_key,
				parameter_key,
				name,
				sort_order
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/event/event_key',
				'amount',
				'Amount',
				1
			)
	`)
	suite.Require().NoError(err)

	name, err = LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	suite.Require().NoError(err)
	suite.Equal("Amount", name)
}

func (suite *EventParameterSuite) TestAdd() {
	err := AddEventParameter(suite.db, suite.model.Key, suite.eventKey, "Amount")
	suite.Require().NoError(err)

	name, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	suite.Require().NoError(err)
	suite.Equal("Amount", name)
}

func (suite *EventParameterSuite) TestUpdate() {
	err := AddEventParameter(suite.db, suite.model.Key, suite.eventKey, "Amount")
	suite.Require().NoError(err)

	err = UpdateEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount", "Total", 2)
	suite.Require().NoError(err)

	name, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	suite.Require().NoError(err)
	suite.Equal("Total", name)
}

func (suite *EventParameterSuite) TestRemove() {
	err := AddEventParameter(suite.db, suite.model.Key, suite.eventKey, "Amount")
	suite.Require().NoError(err)

	err = RemoveEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	suite.Require().NoError(err)

	name, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(name)
}

func (suite *EventParameterSuite) TestQuery() {
	err := AddEventParameters(suite.db, suite.model.Key, map[identity.Key][]string{
		suite.eventKey: {"Alpha", "Bravo"},
	})
	suite.Require().NoError(err)

	names, err := QueryEventParameters(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal(map[identity.Key][]string{
		suite.eventKey: {"Alpha", "Bravo"},
	}, names)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddEventParameter(t *testing.T, dbOrTx DbOrTx, modelKey string, eventKey identity.Key, name string) string {
	err := AddEventParameter(dbOrTx, modelKey, eventKey, name)
	require.NoError(t, err)

	loaded, err := LoadEventParameter(dbOrTx, modelKey, eventKey, identity.NormalizeSubKey(name))
	require.NoError(t, err)

	return loaded
}

func (suite *EventParameterSuite) TestVerifyTestObjects() {
	name := t_AddEventParameter(suite.T(), suite.db, suite.model.Key, suite.eventKey, "Amount")
	suite.Equal("Amount", name)
}