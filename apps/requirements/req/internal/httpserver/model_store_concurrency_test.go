package httpserver

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/require"
)

func TestModelStoreServesReadsDuringRegeneration(t *testing.T) {
	store := NewModelStore()
	model := test_helper.GetTestModel()
	require.NoError(t, store.SetModel("test_model", &model, nil))

	var regenRunning atomic.Bool
	regenDone := make(chan error, 1)
	go func() {
		regenRunning.Store(true)
		regenDone <- store.SetModel("test_model", &model, nil)
	}()

	require.Eventually(t, regenRunning.Load, 2*time.Second, 5*time.Millisecond, "regeneration should start")

	readStart := time.Now()
	ctx := context.Background()
	_, ok := store.GetMarkdown(ctx, "test_model", "model.md")
	readElapsed := time.Since(readStart)
	require.True(t, ok, "expected model hub page in store")
	require.Less(t, readElapsed, 200*time.Millisecond,
		"page read blocked during regeneration for %s; store must serve stale content while generating", readElapsed)

	require.NoError(t, <-regenDone)
}
