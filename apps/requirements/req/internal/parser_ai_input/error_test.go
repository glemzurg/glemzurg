package parser_ai_input

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
	NewParseError(invalidCode, "test message")

	// Should not reach here
	t.Fatal("NewParseError should have panicked")
}
