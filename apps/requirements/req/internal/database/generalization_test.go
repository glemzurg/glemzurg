package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGeneralizationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(GeneralizationSuite))
}

type GeneralizationSuite struct {
	suite.Suite
	db    *sql.DB
	model requirements.Model
}

func (suite *GeneralizationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
}

func (suite *GeneralizationSuite) TestLoad() {

	// Nothing in database yet.
	generalization, err := LoadGeneralization(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), generalization)

	_, err = dbExec(suite.db, `
		INSERT INTO generalization
			(
				model_key,
				generalization_key,
				name,
				details,
				is_complete,
				is_static,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'key',
				'Name',
				'Details',
				true,
				false,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	generalization, err = LoadGeneralization(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Generalization{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *GeneralizationSuite) TestAdd() {

	err := AddGeneralization(suite.db, strings.ToUpper(suite.model.Key), model_class.Generalization{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadGeneralization(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Generalization{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *GeneralizationSuite) TestUpdate() {

	err := AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateGeneralization(suite.db, strings.ToUpper(suite.model.Key), model_class.Generalization{
		Key:        "kEy", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadGeneralization(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Generalization{
		Key:        "key", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	}, generalization)
}

func (suite *GeneralizationSuite) TestRemove() {

	err := AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveGeneralization(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	generalization, err := LoadGeneralization(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), generalization)
}

func (suite *GeneralizationSuite) TestQuery() {

	err := AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	generalizations, err := QueryGeneralizations(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_class.Generalization{
		{
			Key:        "key",
			Name:       "Name",
			Details:    "Details",
			IsComplete: true,
			IsStatic:   false,
			UmlComment: "UmlComment",
		},
		{

			Key:        "keyx",
			Name:       "NameX",
			Details:    "DetailsX",
			IsComplete: false,
			IsStatic:   true,
			UmlComment: "UmlCommentX",
		},
	}, generalizations)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddGeneralization(t *testing.T, dbOrTx DbOrTx, modelKey, generalizationKey string) (generalization model_class.Generalization) {

	err := AddGeneralization(dbOrTx, modelKey, model_class.Generalization{
		Key:        generalizationKey,
		Name:       generalizationKey,
		Details:    "Details",
		IsComplete: true,
		IsStatic:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	generalization, err = LoadGeneralization(dbOrTx, modelKey, generalizationKey)
	assert.Nil(t, err)

	return generalization
}
