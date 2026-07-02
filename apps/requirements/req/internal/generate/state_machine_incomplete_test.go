package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestStateMachineIncompleteMarker(t *testing.T) {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "active"))
	eventAddKey := helper.Must(identity.NewEventKey(classKey, "add"))
	eventNewKey := helper.Must(identity.NewEventKey(classKey, "_new"))

	tests := []struct {
		name  string
		class model_class.Class
		want  string
	}{
		{
			name: "no state machine",
			class: model_class.NewClass(
				classKey,
				model_class.ClassLinks{},
				model_class.ClassDetails{Name: "Empty", Details: ""},
			),
			want: "",
		},
		{
			name: "state machine without «new»",
			class: func() model_class.Class {
				class := model_class.NewClass(
					classKey,
					model_class.ClassLinks{},
					model_class.ClassDetails{Name: "Order", Details: ""},
				)
				class.SetStates(map[identity.Key]model_state.State{
					stateKey: model_state.NewState(stateKey, "Active", "", ""),
				})
				class.SetEvents(map[identity.Key]model_state.Event{
					eventAddKey: model_state.NewEvent(eventAddKey, "Add", "", nil),
				})
				return class
			}(),
			want: "\n\n**«incomplete»** — no «new» event defined for creation transitions.\n",
		},
		{
			name: "state machine with «new»",
			class: func() model_class.Class {
				class := model_class.NewClass(
					classKey,
					model_class.ClassLinks{},
					model_class.ClassDetails{Name: "Order", Details: ""},
				)
				class.SetStates(map[identity.Key]model_state.State{
					stateKey: model_state.NewState(stateKey, "Active", "", ""),
				})
				class.SetEvents(map[identity.Key]model_state.Event{
					eventNewKey: model_state.NewEvent(eventNewKey, model_state.EventNameNew, "", nil),
				})
				return class
			}(),
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, stateMachineIncompleteMarker(tc.class))
		})
	}
}
