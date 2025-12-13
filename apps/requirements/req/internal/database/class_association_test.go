package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAssociationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AssociationSuite))
}

type AssociationSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    requirements.Domain
	subdomain requirements.Subdomain
	class     requirements.Class
	classB    requirements.Class
	classC    requirements.Class
}

func (suite *AssociationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
	suite.classB = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key_b")
	suite.classC = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key_c")
}

func (suite *AssociationSuite) TestLoad() {

	// Nothing in database yet.
	association, err := LoadAssociation(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), association)

	_, err = dbExec(suite.db, `
		INSERT INTO association
			(
				model_key,
				association_key,
				name,
				details,
				from_class_key,
				from_multiplicity_lower,
				from_multiplicity_higher,
				to_class_key,
				to_multiplicity_lower,
				to_multiplicity_higher,
				association_class_key,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'key',
				'Name',
				'Details',
				'class_key',
				'0',
				'1',
				'class_key_b',
				'2',
				'3',
				'class_key_c',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	association, err = LoadAssociation(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Association{
		Key:                 "key", // Test case-insensitive.
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlComment",
	}, association)
}

func (suite *AssociationSuite) TestAdd() {

	err := AddAssociation(suite.db, strings.ToUpper(suite.model.Key), requirements.Association{
		Key:                 "KeY", // Test case-insensitive.
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_KEY", // Test case-insensitive.
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_KEY_b", // Test case-insensitive.
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_KEY_c", // Test case-insensitive.
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlComment",
	}, association)
}

func (suite *AssociationSuite) TestAddNulls() {

	err := AddAssociation(suite.db, strings.ToUpper(suite.model.Key), requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "",
		UmlComment:          "",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "",
		UmlComment:          "",
	}, association)
}

func (suite *AssociationSuite) TestUpdate() {

	err := AddAssociation(suite.db, suite.model.Key, requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAssociation(suite.db, strings.ToUpper(suite.model.Key), requirements.Association{
		Key:                 "KeY", // Test case-insensitive.
		Name:                "NameX",
		Details:             "DetailsX",
		FromClassKey:        "class_KEY_b", // Test case-insensitive.
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          "class_KEY_c", // Test case-insensitive.
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: "class_KEY", // Test case-insensitive.
		UmlComment:          "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Association{
		Key:                 "key",
		Name:                "NameX",
		Details:             "DetailsX",
		FromClassKey:        "class_key_b",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          "class_key_c",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: "class_key",
		UmlComment:          "UmlCommentX",
	}, association)
}

func (suite *AssociationSuite) TestUpdateNulls() {

	err := AddAssociation(suite.db, suite.model.Key, requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAssociation(suite.db, strings.ToUpper(suite.model.Key), requirements.Association{
		Key:                 "key",
		Name:                "NameX",
		Details:             "",
		FromClassKey:        "class_key_b",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          "class_key_c",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: "",
		UmlComment:          "",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Association{
		Key:                 "key",
		Name:                "NameX",
		Details:             "",
		FromClassKey:        "class_key_b",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          "class_key_c",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: "",
		UmlComment:          "",
	}, association)
}

func (suite *AssociationSuite) TestRemove() {

	err := AddAssociation(suite.db, suite.model.Key, requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveAssociation(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), association)
}

func (suite *AssociationSuite) TestQuery() {

	err := AddAssociation(suite.db, suite.model.Key, requirements.Association{
		Key:                 "keyx",
		Name:                "NameX",
		Details:             "DetailsX",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddAssociation(suite.db, suite.model.Key, requirements.Association{
		Key:                 "key",
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        "class_key",
		FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          "class_key_b",
		ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: "class_key_c",
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	associations, err := QueryAssociations(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []requirements.Association{
		{
			Key:                 "key",
			Name:                "Name",
			Details:             "Details",
			FromClassKey:        "class_key",
			FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
			ToClassKey:          "class_key_b",
			ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
			AssociationClassKey: "class_key_c",
			UmlComment:          "UmlComment",
		},
		{
			Key:                 "keyx",
			Name:                "NameX",
			Details:             "DetailsX",
			FromClassKey:        "class_key",
			FromMultiplicity:    requirements.Multiplicity{LowerBound: 0, HigherBound: 1},
			ToClassKey:          "class_key_b",
			ToMultiplicity:      requirements.Multiplicity{LowerBound: 2, HigherBound: 3},
			AssociationClassKey: "class_key_c",
			UmlComment:          "UmlCommentX",
		},
	}, associations)
}
