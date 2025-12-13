package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
	db              *sql.DB
	model           requirements.Model
	domain          requirements.Domain
	subdomain       requirements.Subdomain
	generalization  requirements.Generalization
	generalizationB requirements.Generalization
	actor           requirements.Actor
	actorB          requirements.Actor
}

func (suite *ClassSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.generalization = t_AddGeneralization(suite.T(), suite.db, suite.model.Key, "generalization_key")
	suite.generalizationB = t_AddGeneralization(suite.T(), suite.db, suite.model.Key, "generalization_key_b")
	suite.actor = t_AddActor(suite.T(), suite.db, suite.model.Key, "actor_key")
	suite.actorB = t_AddActor(suite.T(), suite.db, suite.model.Key, "actor_key_b")
}

func (suite *ClassSuite) TestLoad() {

	// Nothing in database yet.
	subdomainKey, class, err := LoadClass(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), class)

	_, err = dbExec(suite.db, `
		INSERT INTO class
			(
				model_key,
				subdomain_key,
				class_key,
				name,
				details,
				actor_key,
				superclass_of_key,
				subclass_of_key,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'subdomain_key',
				'key',
				'Name',
				'Details',
				'actor_key',
				'generalization_key',
				'generalization_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	subdomainKey, class, err = LoadClass(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.Class{
		Key:             "key", // Test case-insensitive.
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "actor_key",
		SuperclassOfKey: "generalization_key",
		SubclassOfKey:   "generalization_key_b",
		UmlComment:      "UmlComment",
	}, class)
}

func (suite *ClassSuite) TestAdd() {

	err := AddClass(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.subdomain.Key), requirements.Class{
		Key:             "KeY", // Test case-insensitive.
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "acTor_Key",            // Test case-insensitive.
		SuperclassOfKey: "generalization_KEY",   // Test case-insensitive.
		SubclassOfKey:   "generalization_KEY_b", // Test case-insensitive.
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.Class{
		Key:             "key",
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "actor_key",
		SuperclassOfKey: "generalization_key",
		SubclassOfKey:   "generalization_key_b",
		UmlComment:      "UmlComment",
	}, class)
}

func (suite *ClassSuite) TestAddNulls() {

	err := AddClass(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.subdomain.Key), requirements.Class{
		Key:             "KeY", // Test case-insensitive.
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "", // No foreign key.
		SuperclassOfKey: "", // No foreign key.
		SubclassOfKey:   "", // No foreign key.
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.Class{
		Key:             "key",
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "", // No foreign key.
		SuperclassOfKey: "", // No foreign key.
		SubclassOfKey:   "", // No foreign key.
		UmlComment:      "UmlComment",
	}, class)
}

func (suite *ClassSuite) TestUpdate() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, requirements.Class{
		Key:             "key",
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "actor_key",
		SuperclassOfKey: "generalization_key",
		SubclassOfKey:   "generalization_key_b",
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateClass(suite.db, strings.ToUpper(suite.model.Key), requirements.Class{
		Key:             "kEy", // Test case-insensitive.
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        "actor_Key_B",          // Test case-insensitive.
		SuperclassOfKey: "generalization_KEY_b", // Test case-insensitive.
		SubclassOfKey:   "generalization_KEY",   // Test case-insensitive.
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.Class{
		Key:             "key",
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        "actor_key_b",
		SuperclassOfKey: "generalization_key_b",
		SubclassOfKey:   "generalization_key",
		UmlComment:      "UmlCommentX",
	}, class)
}

func (suite *ClassSuite) TestUpdateNulls() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, requirements.Class{
		Key:             "key",
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "actor_key",
		SuperclassOfKey: "generalization_key",
		SubclassOfKey:   "generalization_key_b",
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateClass(suite.db, strings.ToUpper(suite.model.Key), requirements.Class{
		Key:             "kEy", // Test case-insensitive.
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        "", // No foreign key.
		SuperclassOfKey: "", // No foreign key.
		SubclassOfKey:   "", // No foreign key.
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `subdomain_key`, subdomainKey)
	assert.Equal(suite.T(), requirements.Class{
		Key:             "key",
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        "", // No foreign key.
		SuperclassOfKey: "", // No foreign key.
		SubclassOfKey:   "", // No foreign key.
		UmlComment:      "UmlCommentX",
	}, class)
}

func (suite *ClassSuite) TestRemove() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, requirements.Class{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		ActorKey:   "actor_key",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveClass(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), class)
}

func (suite *ClassSuite) TestQuery() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, requirements.Class{
		Key:             "keyx",
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        "actor_key_b",
		SuperclassOfKey: "generalization_key_b",
		SubclassOfKey:   "generalization_key",
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddClass(suite.db, suite.model.Key, suite.subdomain.Key, requirements.Class{
		Key:             "key",
		Name:            "Name",
		Details:         "Details",
		ActorKey:        "actor_key",
		SuperclassOfKey: "generalization_key",
		SubclassOfKey:   "generalization_key_b",
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classes, err := QueryClasses(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]requirements.Class{
		suite.subdomain.Key: {
			{
				Key:             "key",
				Name:            "Name",
				Details:         "Details",
				ActorKey:        "actor_key",
				SuperclassOfKey: "generalization_key",
				SubclassOfKey:   "generalization_key_b",
				UmlComment:      "UmlComment",
			},
			{

				Key:             "keyx",
				Name:            "NameX",
				Details:         "DetailsX",
				ActorKey:        "actor_key_b",
				SuperclassOfKey: "generalization_key_b",
				SubclassOfKey:   "generalization_key",
				UmlComment:      "UmlCommentX",
			},
		},
	}, classes)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddClass(t *testing.T, dbOrTx DbOrTx, modelKey, subdomainKey, classKey string) (class requirements.Class) {

	err := AddClass(dbOrTx, modelKey, subdomainKey, requirements.Class{
		Key:        classKey,
		Name:       "Name",
		Details:    "Details",
		ActorKey:   "", // No actor.
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, class, err = LoadClass(dbOrTx, modelKey, classKey)
	assert.Nil(t, err)

	return class
}

func (suite *ClassSuite) TestVerifyTestObjects() {

	class := t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
	assert.Equal(suite.T(), requirements.Class{
		Key:        "class_key",
		Name:       "Name",
		Details:    "Details",
		ActorKey:   "", // No actor.
		UmlComment: "UmlComment",
	}, class)

}
