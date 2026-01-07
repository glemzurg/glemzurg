package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"
	"github.com/stretchr/testify/assert"
)

func TestClassInOutRoundTrip(t *testing.T) {
	original := model_class.Class{
		Key:             "class1",
		Name:            "TestClass",
		Details:         "A test class",
		ActorKey:        "actor1",
		SuperclassOfKey: "super1",
		SubclassOfKey:   "sub1",
		UmlComment:      "comment",
		Attributes: []model_class.Attribute{
			{Key: "attr1", Name: "Attr1", Details: "Details", DataTypeRules: "string", Nullable: false, UmlComment: "comment"},
		},
		States: []model_state.State{
			{Key: "state1", Name: "State1", Details: "Details", UmlComment: "comment"},
		},
		Events: []model_state.Event{
			{Key: "event1", Name: "Event1", Details: "Details"},
		},
		Guards: []model_state.Guard{
			{Key: "guard1", Name: "Guard1", Details: "Details"},
		},
		Actions: []model_state.Action{
			{Key: "action1", Name: "Action1", Details: "Details", Requires: []string{"req1"}, Guarantees: []string{"guar1"}},
		},
		Transitions: []model_state.Transition{
			{Key: "trans1", FromStateKey: "state1", EventKey: "event1", ToStateKey: "state2", UmlComment: "comment"},
		},
	}

	inOut := FromRequirementsClass(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
