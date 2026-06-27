package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestIsReverseInvariantOnlyAssociation(t *testing.T) {
	t.Parallel()

	subdomainKey := helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("finance")), "wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	acKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdictional_wallet_definition"))
	hostKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))
	reverseKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, jurisdictionKey, partnerKey, "configures_customers_for"))
	anyMult := helper.Must(NewMultiplicity("any"))

	host := NewAssociation(
		hostKey,
		AssociationDetails{Name: "Configures Customers For", Details: ""},
		AssociationEnd{ClassKey: partnerKey, Multiplicity: anyMult},
		AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: anyMult},
		&acKey,
		"",
	)
	reverse := NewAssociation(
		reverseKey,
		AssociationDetails{Name: "Configures Customers For", Details: ""},
		AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: anyMult},
		AssociationEnd{ClassKey: partnerKey, Multiplicity: anyMult},
		nil,
		"",
	)

	associations := map[identity.Key]Association{hostKey: host, reverseKey: reverse}

	require.False(t, IsReverseInvariantOnlyAssociation(associations, host))
	require.True(t, IsReverseInvariantOnlyAssociation(associations, reverse))
}
