package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(UseCaseSuite))
}

type UseCaseSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    requirements.Domain
	subdomain requirements.Subdomain
}

func (suite *UseCaseSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
}

func (suite *UseCaseSuite) TestLoad() {

	// Nothing in database yet.
	subdomainKey, useCase, err := LoadUseCase(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), useCase)

	_, err = dbExec(suite.db, `
		INSERT INTO use_case
			(
				model_key,
				subdomain_key,
				use_case_key,
				name,
				details,
				level,
				read_only,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'subdomain_key',
				'key',
				'Name',
				'Details',
				'sea',
				true,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err = LoadUseCase(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.UseCase{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	}, useCase)
}

func (suite *UseCaseSuite) TestAdd() {

	err := AddUseCase(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.subdomain.Key), requirements.UseCase{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Level:      "mud",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.UseCase{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Level:      "mud",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	}, useCase)
}

func (suite *UseCaseSuite) TestUpdate() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, requirements.UseCase{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCase(suite.db, strings.ToUpper(suite.model.Key), requirements.UseCase{
		Key:        "kEy", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Level:      "sky",
		ReadOnly:   false,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.UseCase{
		Key:        "key", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Level:      "sky",
		ReadOnly:   false,
		UmlComment: "UmlCommentX",
	}, useCase)
}

func (suite *UseCaseSuite) TestRemove() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, requirements.UseCase{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveUseCase(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), useCase)
}

func (suite *UseCaseSuite) TestQuery() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, requirements.UseCase{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, requirements.UseCase{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKeys, useCases, err := QueryUseCases(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string]string{
		"keyx": "subdomain_key",
		"key":  "subdomain_key",
	}, subdomainKeys)
	assert.Equal(suite.T(), []requirements.UseCase{
		{
			Key:        "key",
			Name:       "Name",
			Details:    "Details",
			Level:      "sea",
			ReadOnly:   true,
			UmlComment: "UmlComment",
		},
		{

			Key:        "keyx",
			Name:       "NameX",
			Details:    "DetailsX",
			Level:      "sea",
			ReadOnly:   true,
			UmlComment: "UmlCommentX",
		},
	}, useCases)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddUseCase(t *testing.T, dbOrTx DbOrTx, modelKey, subdomainKey, useCaseKey string) (useCase requirements.UseCase) {

	err := AddUseCase(dbOrTx, modelKey, subdomainKey, requirements.UseCase{
		Key:        useCaseKey,
		Name:       "Name",
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, useCase, err = LoadUseCase(dbOrTx, modelKey, useCaseKey)
	assert.Nil(t, err)

	return useCase
}
