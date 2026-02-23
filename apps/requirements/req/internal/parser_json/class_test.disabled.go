package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	actorKey, err := identity.NewActorKey("actor1")
	require.NoError(t, err)
	genKey, err := identity.NewGeneralizationKey(subdomainKey, "super1")
	require.NoError(t, err)
	subGenKey, err := identity.NewGeneralizationKey(subdomainKey, "sub1gen")
	require.NoError(t, err)
	attrKey, err := identity.NewAttributeKey(classKey, "attr1")
	require.NoError(t, err)
	stateKey, err := identity.NewStateKey(classKey, "state1")
	require.NoError(t, err)
	state2Key, err := identity.NewStateKey(classKey, "state2")
	require.NoError(t, err)
	eventKey, err := identity.NewEventKey(classKey, "event1")
	require.NoError(t, err)
	guardKey, err := identity.NewGuardKey(classKey, "guard1")
	require.NoError(t, err)
	actionKey, err := identity.NewActionKey(classKey, "action1")
	require.NoError(t, err)
	transKey, err := identity.NewTransitionKey(classKey, "state1", "event1", "", "", "state2")
	require.NoError(t, err)

	original := model_class.Class{
		Key:             classKey,
		Name:            "TestClass",
		Details:         "A test class",
		ActorKey:        &actorKey,
		SuperclassOfKey: &genKey,
		SubclassOfKey:   &subGenKey,
		UmlComment:      "comment",
		Attributes: map[identity.Key]model_class.Attribute{
			attrKey: {Key: attrKey, Name: "Attr1", Details: "Details", DataTypeRules: "string", Nullable: false, UmlComment: "comment"},
		},
		States: map[identity.Key]model_state.State{
			stateKey: {Key: stateKey, Name: "State1", Details: "Details", UmlComment: "comment"},
		},
		Events: map[identity.Key]model_state.Event{
			eventKey: {Key: eventKey, Name: "Event1", Details: "Details"},
		},
		Guards: map[identity.Key]model_state.Guard{
			guardKey: {Key: guardKey, Name: "Guard1", Details: "Details"},
		},
		Actions: map[identity.Key]model_state.Action{
			actionKey: {Key: actionKey, Name: "Action1", Details: "Details", Requires: []string{"req1"}, Guarantees: []string{"guar1"}},
		},
		Transitions: map[identity.Key]model_state.Transition{
			transKey: {Key: transKey, FromStateKey: &stateKey, EventKey: eventKey, ToStateKey: &state2Key, UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsClass(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
