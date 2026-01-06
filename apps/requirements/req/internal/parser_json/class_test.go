package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestClassInOutRoundTrip(t *testing.T) {
	original := requirements.Class{
		Key:             "class1",
		Name:            "TestClass",
		Details:         "A test class",
		ActorKey:        "actor1",
		SuperclassOfKey: "super1",
		SubclassOfKey:   "sub1",
		UmlComment:      "comment",
		Attributes: []requirements.Attribute{
			{Key: "attr1", Name: "Attr1", Details: "Details", DataTypeRules: "string", Nullable: false, UmlComment: "comment"},
		},
		States: []requirements.State{
			{Key: "state1", Name: "State1", Details: "Details", UmlComment: "comment"},
		},
		Events: []requirements.Event{
			{Key: "event1", Name: "Event1", Details: "Details"},
		},
		Guards: []requirements.Guard{
			{Key: "guard1", Name: "Guard1", Details: "Details"},
		},
		Actions: []requirements.Action{
			{Key: "action1", Name: "Action1", Details: "Details", Requires: []string{"req1"}, Guarantees: []string{"guar1"}},
		},
		Transitions: []requirements.Transition{
			{Key: "trans1", FromStateKey: "state1", EventKey: "event1", ToStateKey: "state2", UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsClass(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
