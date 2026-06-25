package evaluator

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

func TestGetAssociationClassLinksByEndpointPairsRows(t *testing.T) {
	t.Parallel()

	ctx := NewRelationContext()
	hostKey := AssociationKey("host")
	partner := object.NewRecord()
	j1 := object.NewRecord()
	j1.Set("Code", object.NewString("US"))
	j2 := object.NewRecord()
	j2.Set("Code", object.NewString("UK"))
	link1 := object.NewRecord()
	link2 := object.NewRecord()

	ctx.CreateLink(hostKey, partner, j1)
	ctx.CreateLink(hostKey, partner, j2)
	ctx.AddAssociationClassRow(hostKey, partner, j1, link1)
	ctx.AddAssociationClassRow(hostKey, partner, j2, link2)

	links := ctx.GetAssociationClassLinksByEndpoint(partner, hostKey, false)
	require.Len(t, links, 2)
	require.Equal(t, link1, links[j1])
	require.Equal(t, link2, links[j2])
}
