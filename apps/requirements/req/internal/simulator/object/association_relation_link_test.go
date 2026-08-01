package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLinkForEndpointMatchesClonedEndpoint(t *testing.T) {
	attrs := NewRecord()
	attrs.Set("_state", NewString("Active"))
	ep := NewExtentElement(7, attrs)
	link := NewRecord()
	link.Set("_state", NewString("Active"))

	set := NewSet()
	set.Add(ep) // clones
	cloned := set.Elements()[0].(*Record)
	require.NotSame(t, ep, cloned)
	require.True(t, ep.Equals(cloned), "clone should equal original")

	rel := NewAssociationRelation(set, "CurrencyWalletDefinition", map[*Record]*Record{ep: link})
	got, ok := rel.LinkForEndpoint(cloned)
	require.True(t, ok, "LinkForEndpoint must find link via Equals when endpoint is a set clone")
	require.Equal(t, link, got)
}
