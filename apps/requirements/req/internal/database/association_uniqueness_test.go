package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestAssociationUniquenessSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AssociationUniquenessSuite))
}

type AssociationUniquenessSuite struct {
	suite.Suite
	db           *sql.DB
	modelKey     string
	assocKey     identity.Key
	fromAttrKey  identity.Key
	toAttrKey    identity.Key
	bogusAttrKey identity.Key
}

func (suite *AssociationUniquenessSuite) SetupTest() {
	suite.db = t_ResetDatabase(suite.T())

	model := t_AddModel(suite.T(), suite.db)
	suite.modelKey = model.Key
	domain := t_AddDomain(suite.T(), suite.db, model.Key, helper.Must(identity.NewDomainKey("finance")))
	subdomain := t_AddSubdomain(suite.T(), suite.db, model.Key, domain.Key, helper.Must(identity.NewSubdomainKey(domain.Key, "wallet")))
	partner := t_AddClass(suite.T(), suite.db, model.Key, subdomain.Key, helper.Must(identity.NewClassKey(subdomain.Key, "partner")))
	jurisdiction := t_AddClass(suite.T(), suite.db, model.Key, subdomain.Key, helper.Must(identity.NewClassKey(subdomain.Key, "jurisdiction")))

	suite.assocKey = helper.Must(identity.NewClassAssociationKey(subdomain.Key, partner.Key, jurisdiction.Key, "configures_customers_for"))
	suite.fromAttrKey = helper.Must(identity.NewAttributeKey(partner.Key, "partner_code"))
	suite.toAttrKey = helper.Must(identity.NewAttributeKey(jurisdiction.Key, "jurisdiction_code"))
	suite.bogusAttrKey = helper.Must(identity.NewAttributeKey(jurisdiction.Key, "missing_attribute"))

	t_AddAttribute(suite.T(), suite.db, suite.modelKey, partner.Key, suite.fromAttrKey)
	t_AddAttribute(suite.T(), suite.db, suite.modelKey, jurisdiction.Key, suite.toAttrKey)

	uniqueness := model_class.NewAssociationUniqueness(
		[]identity.Key{suite.fromAttrKey},
		[]identity.Key{suite.toAttrKey},
	)
	err := AddAssociations(suite.db, suite.modelKey, []model_class.Association{
		model_class.NewAssociation(
			suite.assocKey,
			model_class.AssociationDetails{Name: "Configures Customers For"},
			model_class.AssociationEnd{ClassKey: partner.Key, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
			model_class.AssociationEnd{ClassKey: jurisdiction.Key, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
			model_class.AssociationOptions{Uniqueness: &uniqueness},
		),
	})
	suite.Require().NoError(err)
}

func (suite *AssociationUniquenessSuite) TestRoundTrip() {
	association, err := LoadAssociation(suite.db, suite.modelKey, suite.assocKey)
	suite.Require().NoError(err)
	suite.Require().NotNil(association.Uniqueness)
	suite.Equal(model_class.AssociationUniqueness{
		FromAttributeKeys: []identity.Key{suite.fromAttrKey},
		ToAttributeKeys:   []identity.Key{suite.toAttrKey},
	}, *association.Uniqueness)
}

func (suite *AssociationUniquenessSuite) TestFKAttribute() {
	err := dbExec(suite.db, `
		INSERT INTO association_uniqueness_attribute
			(model_key, association_key, end_side, attribute_sort_order, attribute_key)
		VALUES
			($1, $2, 'to'::association_end, 0, $3)`,
		suite.modelKey,
		suite.assocKey.String(),
		suite.bogusAttrKey.String(),
	)
	suite.Require().Error(err)
}
