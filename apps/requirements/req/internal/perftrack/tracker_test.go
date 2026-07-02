package perftrack

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTrackerAccumulatesSpans(t *testing.T) {
	tracker := New("test-op")
	tracker.run("phase-a", func() {
		time.Sleep(2 * time.Millisecond)
	})
	tracker.run("phase-b", func() {
		time.Sleep(1 * time.Millisecond)
	})

	assert.Len(t, tracker.spans, 2)
	assert.GreaterOrEqual(t, tracker.spans["phase-a"], 2*time.Millisecond)
	assert.GreaterOrEqual(t, tracker.spans["phase-b"], 1*time.Millisecond)
}

func TestRunUsesContextTracker(t *testing.T) {
	tracker := New("ctx-op")
	ctx := WithContext(context.Background(), tracker)

	Run(ctx, "work", func() {
		time.Sleep(1 * time.Millisecond)
	})

	assert.Contains(t, tracker.spans, "work")
}

func TestRunNoTrackerDoesNotPanic(t *testing.T) {
	Run(context.Background(), "work", func() {})
	RunOn(nil, "work", func() {})
}

func TestLogIfSlowBelowThreshold(t *testing.T) {
	tracker := New("fast-op")
	tracker.LogIfSlow()
	assert.Less(t, tracker.Elapsed(), SlowThreshold)
}

func TestFormatSpansOrdersByDuration(t *testing.T) {
	formatted := formatSpans(map[string]time.Duration{
		"b": 2 * time.Millisecond,
		"a": 5 * time.Millisecond,
		"c": 1 * time.Millisecond,
	})
	assert.Equal(t, "a=5ms b=2ms c=1ms", formatted)
}
