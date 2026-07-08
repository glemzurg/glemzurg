package parser_human

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestBuildSubdomainAssociationsLookupSortsByKey(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("test_domain"))
	problemSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "billing"))
	fulfillmentKey := helper.Must(identity.NewSubdomainKey(domainKey, "fulfillment"))
	analyticsKey := helper.Must(identity.NewSubdomainKey(domainKey, "analytics"))

	fulfillmentAssocKey := helper.Must(identity.NewSubdomainAssociationKey(domainKey, problemSubdomainKey, fulfillmentKey))
	analyticsAssocKey := helper.Must(identity.NewSubdomainAssociationKey(domainKey, problemSubdomainKey, analyticsKey))

	domain := model_domain.NewDomain(domainKey, "Test Domain", "", "", true, "")
	domain.SubdomainAssociations = map[identity.Key]model_domain.SubdomainAssociation{
		fulfillmentAssocKey: model_domain.NewSubdomainAssociation(
			fulfillmentAssocKey, problemSubdomainKey, fulfillmentKey, "orders require fulfillment capacity"),
		analyticsAssocKey: model_domain.NewSubdomainAssociation(
			analyticsAssocKey, problemSubdomainKey, analyticsKey, ""),
	}

	lookup := buildSubdomainAssociationsLookup(map[identity.Key]model_domain.Domain{
		domainKey: domain,
	})

	associations := lookup[problemSubdomainKey.String()]
	require.Len(t, associations, 2)
	require.Equal(t, "analytics", associations[0].SolutionSubdomainKey.SubKey)
	require.Equal(t, "fulfillment", associations[1].SolutionSubdomainKey.SubKey)
}
