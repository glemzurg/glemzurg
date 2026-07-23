package actions

import (
	"testing"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/require"
)

func TestIsSystemCreationEventCall(t *testing.T) {
	ownerNew, err := identity.NewEventKey(
		helper.Must(identity.NewClassKey(
			helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
			"owner",
		)),
		model_state.EventNameNew,
	)
	require.NoError(t, err)
	peerUpdate, err := identity.NewEventKey(
		helper.Must(identity.NewClassKey(
			helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
			"peer",
		)),
		"Update",
	)
	require.NoError(t, err)

	tests := []struct {
		name string
		call *me.EventCall
		want bool
	}{
		{
			name: "owner class _new key is creation",
			call: &me.EventCall{EventKey: ownerNew},
			want: true,
		},
		{
			name: "non-creation event is not creation",
			call: &me.EventCall{EventKey: peerUpdate},
			want: false,
		},
		{
			name: "nil call is not creation",
			call: nil,
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isSystemCreationEventCall(tc.call))
		})
	}
}

func TestDiscoverToEndpointFromRow(t *testing.T) {
	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")),
		"account",
	))
	simState := instance.NewState(nil)
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Exists"))
	inst := simState.CreateInstance(classKey, attrs)

	extent := state.ClassExtentElement(inst.ID, inst.Attributes)
	row := object.NewRecordFromFields(map[string]object.Object{
		"account": extent,
		"amount":  object.NewInteger(75),
	})

	id, ok := discoverToEndpointFromRow(simState, classKey, row)
	require.True(t, ok)
	require.Equal(t, inst.ID, id)

	// Bare data match (no id field).
	flat := object.NewRecordFromFields(map[string]object.Object{
		"account": attrs,
		"amount":  object.NewInteger(100),
	})
	id, ok = discoverToEndpointFromRow(simState, classKey, flat)
	require.True(t, ok)
	require.Equal(t, inst.ID, id)

	// No matching endpoint.
	other := object.NewRecordFromFields(map[string]object.Object{
		"amount": object.NewInteger(1),
	})
	_, ok = discoverToEndpointFromRow(simState, classKey, other)
	require.False(t, ok)
}
