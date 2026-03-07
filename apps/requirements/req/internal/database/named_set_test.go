package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_named_set"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNamedSetSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(NamedSetSuite))
}

type NamedSetSuite struct {
	suite.Suite
	db     *sql.DB
	model  core.Model
	nsKey  identity.Key
	nsKeyB identity.Key
}

func (suite *NamedSetSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)

	// Create the keys for reuse.
	suite.nsKey = helper.Must(identity.NewNamedSetKey("_key"))
	suite.nsKeyB = helper.Must(identity.NewNamedSetKey("_key_b"))
}

func (suite *NamedSetSuite) TestLoad() {
	// Nothing in database yet.
	_, err := LoadNamedSet(suite.db, suite.model.Key, suite.nsKey)
	suite.ErrorIs(err, ErrNotFound)

	// Insert the named set row with raw SQL.
	err = dbExec(suite.db, `
		INSERT INTO named_set
			(model_key, set_key, name, description, notation, specification)
		VALUES
			('model_key', 'nset/_key', '_ValidStatuses', 'Valid statuses', 'tla_plus', '{"pending", "active"}')
	`)
	assert.Nil(suite.T(), err)

	ns, err := LoadNamedSet(suite.db, suite.model.Key, suite.nsKey)
	assert.Nil(suite.T(), err)
	suite.Equal(model_named_set.NamedSet{
		Key:         suite.nsKey,
		Name:        "_ValidStatuses",
		Description: "Valid statuses",
		Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{"pending", "active"}`},
	}, ns)
}

func (suite *NamedSetSuite) TestAdd() {
	err := AddNamedSet(suite.db, suite.model.Key, model_named_set.NamedSet{
		Key:         suite.nsKey,
		Name:        "_ValidStatuses",
		Description: "Valid statuses",
		Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{"pending", "active"}`},
	})
	assert.Nil(suite.T(), err)

	ns, err := LoadNamedSet(suite.db, suite.model.Key, suite.nsKey)
	assert.Nil(suite.T(), err)
	suite.Equal(model_named_set.NamedSet{
		Key:         suite.nsKey,
		Name:        "_ValidStatuses",
		Description: "Valid statuses",
		Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{"pending", "active"}`},
	}, ns)
}

func (suite *NamedSetSuite) TestAddWithTypeSpec() {
	ts := model_spec.TypeSpec{Notation: "tla_plus", Specification: "SUBSET STRING"}
	err := AddNamedSet(suite.db, suite.model.Key, model_named_set.NamedSet{
		Key:         suite.nsKey,
		Name:        "_ValidStatuses",
		Description: "Valid statuses",
		Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{"pending", "active"}`},
		TypeSpec:    &ts,
	})
	assert.Nil(suite.T(), err)

	ns, err := LoadNamedSet(suite.db, suite.model.Key, suite.nsKey)
	assert.Nil(suite.T(), err)
	suite.Equal(model_named_set.NamedSet{
		Key:         suite.nsKey,
		Name:        "_ValidStatuses",
		Description: "Valid statuses",
		Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{"pending", "active"}`},
		TypeSpec:    &model_spec.TypeSpec{Notation: "tla_plus", Specification: "SUBSET STRING"},
	}, ns)
}

func (suite *NamedSetSuite) TestRemove() {
	err := AddNamedSet(suite.db, suite.model.Key, model_named_set.NamedSet{
		Key:         suite.nsKey,
		Name:        "_ValidStatuses",
		Description: "Valid statuses",
		Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{"pending", "active"}`},
	})
	assert.Nil(suite.T(), err)

	err = RemoveNamedSet(suite.db, suite.model.Key, suite.nsKey)
	assert.Nil(suite.T(), err)

	// Named set should be gone.
	_, err = LoadNamedSet(suite.db, suite.model.Key, suite.nsKey)
	suite.ErrorIs(err, ErrNotFound)
}

func (suite *NamedSetSuite) TestQuery() {
	err := AddNamedSets(suite.db, suite.model.Key, []model_named_set.NamedSet{
		{
			Key:         suite.nsKeyB,
			Name:        "_Min",
			Description: "Min set",
			Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{1, 2}`},
		},
		{
			Key:         suite.nsKey,
			Name:        "_Max",
			Description: "Max set",
			Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{3, 4}`},
		},
	})
	assert.Nil(suite.T(), err)

	nss, err := QueryNamedSets(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	suite.Equal([]model_named_set.NamedSet{
		{
			Key:         suite.nsKey,
			Name:        "_Max",
			Description: "Max set",
			Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{3, 4}`},
		},
		{
			Key:         suite.nsKeyB,
			Name:        "_Min",
			Description: "Min set",
			Spec:        model_spec.ExpressionSpec{Notation: "tla_plus", Specification: `{1, 2}`},
		},
	}, nss)
}

//==================================================
