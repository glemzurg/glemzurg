package parser_ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewParseErrorPanicsForMissingErrorDoc(t *testing.T) {
	// Use an error code that has no corresponding .md file
	invalidCode := 99999

	defer func() {
		r := recover()
		assert.NotNil(t, r, "NewParseError should panic for missing error doc")

		panicMsg, ok := r.(string)
		assert.True(t, ok, "panic value should be a string")

		// Verify the panic message contains the internal error prefix
		assert.True(t, strings.HasPrefix(panicMsg, InternalErrorPrefix),
			"panic message should start with InternalErrorPrefix, got: "+panicMsg)

		// Verify it mentions this is not something input changes can fix
		assert.Contains(t, panicMsg, "no alteration of input will resolve",
			"panic message should indicate input changes won't help")
	}()

	// This should panic
	_ = NewParseError(invalidCode, "test message", "test.json")

	// Should not reach here
	t.Fatal("NewParseError should have panicked")
}

func TestParseErrorConciseOutput(t *testing.T) {
	err := NewParseError(ErrModelNameRequired, "model name is required", "model.json")
	output := err.Error()

	assert.Contains(t, output, "E1001:")
	assert.Contains(t, output, "model name is required")
	assert.Contains(t, output, "file: model.json")
	// Should NOT contain verbose markdown
	assert.NotContains(t, output, "--- Error Detail ---")
}

func TestParseErrorWithField(t *testing.T) {
	err := NewParseError(ErrModelNameRequired, "model name is required", "model.json").
		WithField("name")
	output := err.Error()

	assert.Contains(t, output, "field: name")
}

func TestParseErrorWithHint(t *testing.T) {
	err := NewParseError(ErrModelNameRequired, "model name is required", "model.json").
		WithHint(`required: "name" (non-empty string)`)
	output := err.Error()

	assert.Contains(t, output, "hint: required: \"name\" (non-empty string)")
}

func TestParseErrorVerboseOutput(t *testing.T) {
	err := NewParseError(ErrModelNameRequired, "model name is required", "model.json").
		WithHint("add a name field")

	verbose := err.VerboseError()
	assert.Contains(t, verbose, "E1001:")
	assert.Contains(t, verbose, "--- Error Detail ---")
}

func TestParseErrorBuildersPreserveFields(t *testing.T) {
	base := NewParseError(ErrModelNameRequired, "msg", "file.json")

	withHint := base.WithHint("hint text")
	assert.Equal(t, "hint text", withHint.Hint)
	assert.Equal(t, base.Code, withHint.Code)

	withField := withHint.WithField("name")
	assert.Equal(t, "name", withField.Field)
	assert.Equal(t, "hint text", withField.Hint)
}
