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

	ctx.createLink(hostKey, partner, j1)
	ctx.createLink(hostKey, partner, j2)
	ctx.AddAssociationClassRow(hostKey, partner, j1, link1)
	ctx.AddAssociationClassRow(hostKey, partner, j2, link2)

	links := ctx.GetAssociationClassLinksByEndpoint(partner, hostKey, false)
	require.Len(t, links, 2)
	require.Equal(t, link1, links[j1])
	require.Equal(t, link2, links[j2])
}

// Set images clone extent elements; CHOOSE/set peers are not pointer-equal to the
// registered anchor, but AC rows must still resolve (Player Initialize path).
func TestGetAssociationClassLinksByEndpointAcceptsClonedExtentAnchor(t *testing.T) {
	t.Parallel()

	ctx := NewRelationContext()
	hostKey := AssociationKey("host")
	partnerAttrs := object.NewRecord()
	partnerAttrs.Set("_state", object.NewString("Active"))
	jurisdictionAttrs := object.NewRecord()
	jurisdictionAttrs.Set("_state", object.NewString("Active"))
	linkAttrs := object.NewRecord()
	linkAttrs.Set("_state", object.NewString("Active"))

	partnerExtent := object.NewExtentElement(1, partnerAttrs)
	jurisdictionExtent := object.NewExtentElement(2, jurisdictionAttrs)
	linkExtent := object.NewExtentElement(3, linkAttrs)

	ctx.CreateInstanceLink(hostKey,
		InstanceEndpoint{ID: 1, Extent: partnerExtent, Data: partnerAttrs},
		InstanceEndpoint{ID: 2, Extent: jurisdictionExtent, Data: jurisdictionAttrs},
	)
	ctx.EnsureInstance(3, linkAttrs)
	ctx.AddAssociationClassRow(hostKey, partnerExtent, jurisdictionExtent, linkExtent)

	// Mimic set membership: clone is not registered by pointer.
	set := object.NewSet()
	set.Add(partnerExtent)
	clonedAnchor := set.Elements()[0].(*object.Record)
	require.NotSame(t, partnerExtent, clonedAnchor)

	links := ctx.GetAssociationClassLinksByEndpoint(clonedAnchor, hostKey, false)
	require.Len(t, links, 1)
	require.Equal(t, linkExtent, links[jurisdictionExtent])
}
