package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

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
	db                 *sql.DB
	model              req_model.Model
	domain             model_domain.Domain
	subdomain          model_domain.Subdomain
	generalization     model_class.Generalization
	generalizationB    model_class.Generalization
	actor              model_actor.Actor
	actorB             model_actor.Actor
	classKey           identity.Key
	classKeyB          identity.Key
}

func (suite *ClassSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.generalization = t_AddGeneralization(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewGeneralizationKey(suite.subdomain.Key, "generalization_key")))
	suite.generalizationB = t_AddGeneralization(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewGeneralizationKey(suite.subdomain.Key, "generalization_key_b")))
	suite.actor = t_AddActor(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewActorKey("actor_key")))
	suite.actorB = t_AddActor(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewActorKey("actor_key_b")))

	// Create the class keys for reuse.
	suite.classKey = helper.Must(identity.NewClassKey(suite.subdomain.Key, "key"))
	suite.classKeyB = helper.Must(identity.NewClassKey(suite.subdomain.Key, "key_b"))
}

func (suite *ClassSuite) TestLoad() {

	// Nothing in database yet.
	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, suite.classKey)
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
				'domain/domain_key/subdomain/subdomain_key',
				'domain/domain_key/subdomain/subdomain_key/class/key',
				'Name',
				'Details',
				'actor/actor_key',
				'domain/domain_key/subdomain/subdomain_key/generalization/generalization_key',
				'domain/domain_key/subdomain/subdomain_key/generalization/generalization_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	subdomainKey, class, err = LoadClass(suite.db, suite.model.Key, suite.classKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &suite.actor.Key,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	}, class)
}

func (suite *ClassSuite) TestAdd() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &suite.actor.Key,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, suite.classKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &suite.actor.Key,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	}, class)
}

func (suite *ClassSuite) TestAddNulls() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        nil, // No foreign key.
		SuperclassOfKey: nil, // No foreign key.
		SubclassOfKey:   nil, // No foreign key.
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, suite.classKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        nil, // No foreign key.
		SuperclassOfKey: nil, // No foreign key.
		SubclassOfKey:   nil, // No foreign key.
		UmlComment:      "UmlComment",
	}, class)
}

func (suite *ClassSuite) TestUpdate() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &suite.actor.Key,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateClass(suite.db, suite.model.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        &suite.actorB.Key,
		SuperclassOfKey: &suite.generalizationB.Key,
		SubclassOfKey:   &suite.generalization.Key,
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, suite.classKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_class.Class{
		Key:             suite.classKey,
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        &suite.actorB.Key,
		SuperclassOfKey: &suite.generalizationB.Key,
		SubclassOfKey:   &suite.generalization.Key,
		UmlComment:      "UmlCommentX",
	}, class)
}

func (suite *ClassSuite) TestUpdateNulls() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &suite.actor.Key,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateClass(suite.db, suite.model.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        nil, // No foreign key.
		SuperclassOfKey: nil, // No foreign key.
		SubclassOfKey:   nil, // No foreign key.
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, suite.classKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_class.Class{
		Key:             suite.classKey,
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        nil, // No foreign key.
		SuperclassOfKey: nil, // No foreign key.
		SubclassOfKey:   nil, // No foreign key.
		UmlComment:      "UmlCommentX",
	}, class)
}

func (suite *ClassSuite) TestRemove() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:        suite.classKey,
		Name:       "Name",
		Details:    "Details",
		ActorKey:   &suite.actor.Key,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveClass(suite.db, suite.model.Key, suite.classKey)
	assert.Nil(suite.T(), err)

	subdomainKey, class, err := LoadClass(suite.db, suite.model.Key, suite.classKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), class)
}

func (suite *ClassSuite) TestQuery() {

	err := AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:             suite.classKeyB,
		Name:            "NameX",
		Details:         "DetailsX",
		ActorKey:        &suite.actorB.Key,
		SuperclassOfKey: &suite.generalizationB.Key,
		SubclassOfKey:   &suite.generalization.Key,
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddClass(suite.db, suite.model.Key, suite.subdomain.Key, model_class.Class{
		Key:             suite.classKey,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &suite.actor.Key,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classes, err := QueryClasses(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_class.Class{
		suite.subdomain.Key: {
			{
				Key:             suite.classKey,
				Name:            "Name",
				Details:         "Details",
				ActorKey:        &suite.actor.Key,
				SuperclassOfKey: &suite.generalization.Key,
				SubclassOfKey:   &suite.generalizationB.Key,
				UmlComment:      "UmlComment",
			},
			{
				Key:             suite.classKeyB,
				Name:            "NameX",
				Details:         "DetailsX",
				ActorKey:        &suite.actorB.Key,
				SuperclassOfKey: &suite.generalizationB.Key,
				SubclassOfKey:   &suite.generalization.Key,
				UmlComment:      "UmlCommentX",
			},
		},
	}, classes)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddClass(t *testing.T, dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key, classKey identity.Key) (class model_class.Class) {

	err := AddClass(dbOrTx, modelKey, subdomainKey, model_class.Class{
		Key:        classKey,
		Name:       classKey.String(),
		Details:    "Details",
		ActorKey:   nil, // No actor.
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, class, err = LoadClass(dbOrTx, modelKey, classKey)
	assert.Nil(t, err)

	return class
}

func (suite *ClassSuite) TestVerifyTestObjects() {

	class := t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, suite.classKey)
	assert.Equal(suite.T(), model_class.Class{
		Key:        suite.classKey,
		Name:       suite.classKey.String(),
		Details:    "Details",
		ActorKey:   nil, // No actor.
		UmlComment: "UmlComment",
	}, class)
}
