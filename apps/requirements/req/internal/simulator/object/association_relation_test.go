package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssociationRelationEndpointsAndMember(t *testing.T) {
	t.Parallel()

	endpoint := NewRecord()
	link := NewRecord()
	endpoints := NewSetFromElements([]Object{endpoint})
	linkByEndpoint := map[*Record]*Record{endpoint: link}
	rel := NewAssociationRelation(endpoints, "LinkDef", linkByEndpoint)

	require.Equal(t, TypeAssociationRelation, rel.Type())
	require.Equal(t, "LinkDef", rel.LinkClassMember())
	require.Equal(t, 1, rel.Endpoints().Size())
	resolved, ok := rel.LinkForEndpoint(endpoint)
	require.True(t, ok)
	require.Equal(t, link, resolved)
	require.True(t, rel.Equals(rel.Clone().(*AssociationRelation)))
}
