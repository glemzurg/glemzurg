package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/state"
	"github.com/stretchr/testify/assert"
)

func TestClassInOutRoundTrip(t *testing.T) {
	original := class.Class{
		Key:             "class1",
		Name:            "TestClass",
		Details:         "A test class",
		ActorKey:        "actor1",
		SuperclassOfKey: "super1",
		SubclassOfKey:   "sub1",
		UmlComment:      "comment",
		Attributes: []class.Attribute{
			{Key: "attr1", Name: "Attr1", Details: "Details", DataTypeRules: "string", Nullable: false, UmlComment: "comment"},
		},
		States: []state.State{
			{Key: "state1", Name: "State1", Details: "Details", UmlComment: "comment"},
		},
		Events: []state.Event{
			{Key: "event1", Name: "Event1", Details: "Details"},
		},
		Guards: []state.Guard{
			{Key: "guard1", Name: "Guard1", Details: "Details"},
		},
		Actions: []state.Action{
			{Key: "action1", Name: "Action1", Details: "Details", Requires: []string{"req1"}, Guarantees: []string{"guar1"}},
		},
		Transitions: []state.Transition{
			{Key: "trans1", FromStateKey: "state1", EventKey: "event1", ToStateKey: "state2", UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsClass(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
