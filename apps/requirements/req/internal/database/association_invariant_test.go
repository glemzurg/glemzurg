package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAssociationInvariantSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(AssociationInvariantSuite))
}

type AssociationInvariantSuite struct {
	suite.Suite
	db *sql.DB
}

func (suite *AssociationInvariantSuite) SetupTest() {
	suite.db = t_ResetDatabase(suite.T())
}

func TestFromAnchoredInvariantKeysByAssociation(t *testing.T) {
	t.Parallel()

	assocKey := helper.Must(identity.ParseKey("domain/finance/subdomain/wallet/cassociation/class/partner/class/jurisdiction/configures_customers_for"))
	logicA := helper.Must(identity.ParseKey("domain/finance/subdomain/wallet/cassociation/class/partner/class/jurisdiction/configures_customers_for/cassocinvariant/0"))
	logicB := helper.Must(identity.ParseKey("domain/finance/subdomain/wallet/class/jurisdiction/cinvariant/0"))

	links := []AssociationInvariantLink{
		{AssociationKey: assocKey, LogicKey: logicA},
		{AssociationKey: assocKey, LogicKey: logicB, ToClassAnchor: true},
	}

	fromAnchored := FromAnchoredInvariantKeysByAssociation(links)
	require.Equal(t, []identity.Key{logicA}, fromAnchored[assocKey])

	toAnchored := ToAnchoredAssociationKeyByLogic(links)
	require.Equal(t, assocKey, toAnchored[logicB])
}

func (suite *AssociationInvariantSuite) TestAddAndQueryAssociationInvariantLinks() {
	model := t_AddModel(suite.T(), suite.db)
	domain := t_AddDomain(suite.T(), suite.db, model.Key, helper.Must(identity.NewDomainKey("finance")))
	subdomain := t_AddSubdomain(suite.T(), suite.db, model.Key, domain.Key, helper.Must(identity.NewSubdomainKey(domain.Key, "wallet")))
	partner := t_AddClass(suite.T(), suite.db, model.Key, subdomain.Key, helper.Must(identity.NewClassKey(subdomain.Key, "partner")))
	jurisdiction := t_AddClass(suite.T(), suite.db, model.Key, subdomain.Key, helper.Must(identity.NewClassKey(subdomain.Key, "jurisdiction")))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomain.Key, partner.Key, jurisdiction.Key, "configures_customers_for"))

	err := AddAssociations(suite.db, model.Key, []model_class.Association{
		model_class.NewAssociation(
			assocKey,
			model_class.AssociationDetails{Name: "Configures Customers For"},
			model_class.AssociationEnd{ClassKey: partner.Key, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
			model_class.AssociationEnd{ClassKey: jurisdiction.Key, Multiplicity: helper.Must(model_class.NewMultiplicity("any"))},
			model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
		),
	})
	suite.Require().NoError(err)

	fromLogicKey := helper.Must(identity.NewClassAssociationInvariantKey(assocKey, "0"))
	toLogicKey := helper.Must(identity.NewClassInvariantKey(jurisdiction.Key, "0"))
	for _, logicKey := range []identity.Key{fromLogicKey, toLogicKey} {
		err = AddLogic(suite.db, model.Key, model_logic.Logic{
			Key:         logicKey,
			Type:        model_logic.LogicTypeAssessment,
			Description: logicKey.String(),
			Spec:        logic_spec.ExpressionSpec{Notation: "tla_plus", Specification: "TRUE"},
		})
		suite.Require().NoError(err)
	}

	links := []AssociationInvariantLink{
		{AssociationKey: assocKey, LogicKey: fromLogicKey},
		{AssociationKey: assocKey, LogicKey: toLogicKey, ToClassAnchor: true},
	}
	err = AddAssociationInvariantLinks(suite.db, model.Key, links)
	suite.Require().NoError(err)

	loaded, err := QueryAssociationInvariantLinks(suite.db, model.Key)
	suite.Require().NoError(err)
	suite.Require().Len(loaded, 2)

	fromAnchored := FromAnchoredInvariantKeysByAssociation(loaded)
	suite.Require().Len(fromAnchored[assocKey], 1)

	logics, err := QueryLogics(suite.db, model.Key)
	suite.Require().NoError(err)
	logicsByKey := make(map[identity.Key]model_logic.Logic, len(logics))
	for _, logic := range logics {
		logicsByKey[logic.Key] = logic
	}
	_, ok := logicsByKey[fromAnchored[assocKey][0]]
	suite.True(ok)

	toAnchored := ToAnchoredAssociationKeyByLogic(loaded)
	suite.Equal(assocKey, toAnchored[toLogicKey])

	classInvariants, err := QueryClassInvariants(suite.db, model.Key)
	suite.Require().NoError(err)
	suite.Equal([]identity.Key{toLogicKey}, classInvariants[jurisdiction.Key])
}
