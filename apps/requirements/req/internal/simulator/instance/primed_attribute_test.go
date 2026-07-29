package instance

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/stretchr/testify/require"
)

// TestPrimedAndUnprimedAttributeLookup builds a test-only schema and State, then
// checks that exported instance APIs return both current and primed field values.
func TestPrimedAndUnprimedAttributeLookup(t *testing.T) {
	sch := emptySchema()
	st := NewState(sch)
	require.NotNil(t, st.schema())

	domainKey, err := identity.NewDomainKey("sandbox")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "orders")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "order")
	require.NoError(t, err)

	inst := st.CreateInstance(classKey, object.NewRecordFromFields(map[string]object.Object{
		"status": object.NewString("pending"),
		"total":  object.NewInteger(100),
	}))
	require.NotNil(t, inst)

	// Non-primed (current) values come from GetAttribute / live Attributes.
	require.Equal(t, "pending", inst.GetAttribute("status").(*object.String).Value())
	require.Equal(t, "100", inst.GetAttribute("total").(*object.Number).Inspect())

	// No prime until SetPrimedAttribute.
	_, ok := inst.GetPrimedAttribute("status")
	require.False(t, ok)

	inst.SetPrimedAttribute("status", object.NewString("shipped"))
	inst.SetPrimedAttribute("total", object.NewInteger(150))

	// Primed values are visible via GetPrimedAttribute…
	primedStatus, ok := inst.GetPrimedAttribute("status")
	require.True(t, ok)
	require.Equal(t, "shipped", primedStatus.(*object.String).Value())

	primedTotal, ok := inst.GetPrimedAttribute("total")
	require.True(t, ok)
	require.Equal(t, "150", primedTotal.(*object.Number).Inspect())

	// …while current (unprimed) values stay unchanged until commit.
	require.Equal(t, "pending", inst.GetAttribute("status").(*object.String).Value())
	require.Equal(t, "100", inst.GetAttribute("total").(*object.Number).Inspect())

	// Lookup through State still sees both sides (same live instance pointer).
	fromState := st.GetInstance(inst.GetID())
	require.Same(t, inst, fromState)
	require.Equal(t, "pending", fromState.GetAttribute("status").(*object.String).Value())
	primedFromState, ok := fromState.GetPrimedAttribute("status")
	require.True(t, ok)
	require.Equal(t, "shipped", primedFromState.(*object.String).Value())
}
