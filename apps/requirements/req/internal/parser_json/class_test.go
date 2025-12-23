package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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

	// Check basic fields
	if back.Key != original.Key || back.Name != original.Name || back.Details != original.Details ||
		back.ActorKey != original.ActorKey || back.SuperclassOfKey != original.SuperclassOfKey ||
		back.SubclassOfKey != original.SubclassOfKey || back.UmlComment != original.UmlComment {
		t.Errorf("Basic fields round trip failed: got %+v, want %+v", back, original)
	}

	// Check lengths of slices
	if len(back.Attributes) != len(original.Attributes) || len(back.States) != len(original.States) ||
		len(back.Events) != len(original.Events) || len(back.Guards) != len(original.Guards) ||
		len(back.Actions) != len(original.Actions) || len(back.Transitions) != len(original.Transitions) {
		t.Errorf("Slice lengths don't match: got attrs=%d, states=%d, events=%d, guards=%d, actions=%d, transitions=%d",
			len(back.Attributes), len(back.States), len(back.Events), len(back.Guards), len(back.Actions), len(back.Transitions))
	}
}
