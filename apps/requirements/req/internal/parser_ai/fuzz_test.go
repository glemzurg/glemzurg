package parser_ai

import (
	"testing"
)

// FuzzParseModel ensures parseModel never panics on arbitrary input.
func FuzzParseModel(f *testing.F) {
	// Seed corpus with valid and edge-case inputs.
	f.Add([]byte(`{"name": "Test Model"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`{"name": ""}`))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"name": "Test", "details": "Some details"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should never panic — either returns a result or an error.
		_, _ = parseModel(data, "fuzz_model.json")
	})
}

// FuzzParseActor ensures parseActor never panics on arbitrary input.
func FuzzParseActor(f *testing.F) {
	f.Add([]byte(`{"name": "Customer", "type": "human"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"name": "", "type": "invalid"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseActor(data, "fuzz_actor.json")
	})
}

// FuzzParseClass ensures parseClass never panics on arbitrary input.
func FuzzParseClass(f *testing.F) {
	f.Add([]byte(`{"name": "Order"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"name": "Order", "attributes": {"id": {"name": "ID"}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseClass(data, "fuzz_class.json")
	})
}

// FuzzParseAction ensures parseAction never panics on arbitrary input.
func FuzzParseAction(f *testing.F) {
	f.Add([]byte(`{"name": "Create"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseAction(data, "fuzz_action.json")
	})
}

// FuzzParseStateMachine ensures parseStateMachine never panics on arbitrary input.
func FuzzParseStateMachine(f *testing.F) {
	f.Add([]byte(`{"states": {"pending": {"name": "Pending"}}, "events": {}, "transitions": []}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseStateMachine(data, "fuzz_state_machine.json")
	})
}

// FuzzParseAssociation ensures parseAssociation never panics on arbitrary input.
func FuzzParseAssociation(f *testing.F) {
	f.Add([]byte(`{"name": "Test", "from_class_key": "a", "to_class_key": "b", "from_multiplicity": "1", "to_multiplicity": "0..*"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseAssociation(data, "fuzz_assoc.json")
	})
}
