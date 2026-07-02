package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestResolveRelativeClassAssociationKeyRoundTrip(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))

	relative, err := RelativeClassAssociationKey(jurisdictionKey, assocKey)
	require.NoError(t, err)
	require.Equal(t, "partner/configures_customers_for", relative)

	resolved, err := ResolveClassAssociationKeyFromRelative(subdomainKey, jurisdictionKey, relative)
	require.NoError(t, err)
	require.Equal(t, assocKey, resolved)
}

func TestResolveClassAssociationKeyFromRelativeFullKey(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	assocKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))

	resolved, err := ResolveClassAssociationKeyFromRelative(subdomainKey, jurisdictionKey, assocKey.String())
	require.NoError(t, err)
	require.Equal(t, assocKey, resolved)
}
