package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

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
	db             *sql.DB
	model          req_model.Model
	domain         model_domain.Domain
	subdomain      model_domain.Subdomain
	class          model_class.Class
	classB         model_class.Class
	classC         model_class.Class
	associationKey identity.Key
}

func (suite *AssociationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.classB = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key_b")))
	suite.classC = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key_c")))
	suite.associationKey = helper.Must(identity.NewClassAssociationKey(suite.subdomain.Key, suite.class.Key, suite.classB.Key))
}

func (suite *AssociationSuite) TestLoad() {

	// Nothing in database yet.
	association, err := LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
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
				'domain/domain_key/subdomain/subdomain_key/cassociation/class/class_key/class/class_key_b',
				'Name',
				'Details',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'0',
				'1',
				'domain/domain_key/subdomain/subdomain_key/class/class_key_b',
				'2',
				'3',
				'domain/domain_key/subdomain/subdomain_key/class/class_key_c',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	association, err = LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	}, association)
}

func (suite *AssociationSuite) TestAdd() {

	err := AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	}, association)
}

func (suite *AssociationSuite) TestAddNulls() {

	err := AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: nil, // No association class
		UmlComment:          "",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: nil, // No association class
		UmlComment:          "",
	}, association)
}

func (suite *AssociationSuite) TestUpdate() {

	err := AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey, // Same key, updating other fields
		Name:                "NameX",
		Details:             "DetailsX",
		FromClassKey:        suite.classB.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          suite.classC.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: &suite.class.Key,
		UmlComment:          "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Association{
		Key:                 suite.associationKey,
		Name:                "NameX",
		Details:             "DetailsX",
		FromClassKey:        suite.classB.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          suite.classC.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: &suite.class.Key,
		UmlComment:          "UmlCommentX",
	}, association)
}

func (suite *AssociationSuite) TestUpdateNulls() {

	err := AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "NameX",
		Details:             "",
		FromClassKey:        suite.classB.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          suite.classC.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: nil, // No association class
		UmlComment:          "",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Association{
		Key:                 suite.associationKey,
		Name:                "NameX",
		Details:             "",
		FromClassKey:        suite.classB.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		ToClassKey:          suite.classC.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		AssociationClassKey: nil, // No association class
		UmlComment:          "",
	}, association)
}

func (suite *AssociationSuite) TestRemove() {

	err := AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)

	association, err := LoadAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), association)
}

func (suite *AssociationSuite) TestQuery() {

	// Create a second association key
	associationKeyX := helper.Must(identity.NewClassAssociationKey(suite.subdomain.Key, suite.classB.Key, suite.classC.Key))

	err := AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 associationKeyX, // This key comes after suite.associationKey alphabetically
		Name:                "NameX",
		Details:             "DetailsX",
		FromClassKey:        suite.classB.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classC.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.class.Key,
		UmlComment:          "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddAssociation(suite.db, suite.model.Key, model_class.Association{
		Key:                 suite.associationKey,
		Name:                "Name",
		Details:             "Details",
		FromClassKey:        suite.class.Key,
		FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
		ToClassKey:          suite.classB.Key,
		ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
		AssociationClassKey: &suite.classC.Key,
		UmlComment:          "UmlComment",
	})
	assert.Nil(suite.T(), err)

	associations, err := QueryAssociations(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_class.Association{
		{
			Key:                 suite.associationKey, // class/class_key comes before class/class_key_b
			Name:                "Name",
			Details:             "Details",
			FromClassKey:        suite.class.Key,
			FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
			ToClassKey:          suite.classB.Key,
			ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
			AssociationClassKey: &suite.classC.Key,
			UmlComment:          "UmlComment",
		},
		{
			Key:                 associationKeyX,
			Name:                "NameX",
			Details:             "DetailsX",
			FromClassKey:        suite.classB.Key,
			FromMultiplicity:    model_class.Multiplicity{LowerBound: 0, HigherBound: 1},
			ToClassKey:          suite.classC.Key,
			ToMultiplicity:      model_class.Multiplicity{LowerBound: 2, HigherBound: 3},
			AssociationClassKey: &suite.class.Key,
			UmlComment:          "UmlCommentX",
		},
	}, associations)
}
