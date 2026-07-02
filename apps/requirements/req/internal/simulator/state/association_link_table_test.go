package state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestAssociationLinkTableIndexesByEndpoints(t *testing.T) {
	hostKey := mustAssocKey("host")
	link := AssociationLink{
		HostAssocKey:   hostKey,
		FromEndpointID: 1,
		ToEndpointID:   2,
		LinkInstanceID: 3,
	}

	table := NewAssociationLinkTable()
	require.NoError(t, table.AddLink(link))

	fromLinks := table.LinksFromEndpoint(hostKey, 1)
	require.Len(t, fromLinks, 1)
	require.Equal(t, InstanceID(3), fromLinks[0].LinkInstanceID)

	toLinks := table.LinksToEndpoint(hostKey, 2)
	require.Len(t, toLinks, 1)
	require.Equal(t, InstanceID(3), toLinks[0].LinkInstanceID)

	got, ok := table.LinkByInstance(3)
	require.True(t, ok)
	require.Equal(t, link, got)
}

func mustAssocKey(name string) identity.Key {
	key, err := identity.NewClassAssociationKey(
		mustParseKey("domain/d/subdomain/s"),
		mustParseKey("domain/d/subdomain/s/class/a"),
		mustParseKey("domain/d/subdomain/s/class/b"),
		name,
	)
	if err != nil {
		panic(err)
	}
	return key
}

func mustParseKey(s string) identity.Key {
	key, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return key
}
