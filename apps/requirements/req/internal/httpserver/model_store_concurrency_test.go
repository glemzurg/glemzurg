package httpserver

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/stretchr/testify/require"
)

func TestModelStoreServesReadsDuringRegeneration(t *testing.T) {
	model, _, err := parser_human.Parse("/workspaces/glemzurg/data_sandbox/model/evenplay")
	if err != nil {
		t.Skip("evenplay model not available:", err)
	}

	store := NewModelStore()
	require.NoError(t, store.SetModel("evenplay", &model, nil))

	var regenRunning atomic.Bool
	regenDone := make(chan error, 1)
	go func() {
		regenRunning.Store(true)
		regenDone <- store.SetModel("evenplay", &model, nil)
	}()

	require.Eventually(t, regenRunning.Load, 2*time.Second, 5*time.Millisecond, "regeneration should start")

	readStart := time.Now()
	ctx := context.Background()
	_, ok := store.GetMarkdown(ctx, "evenplay", "domain-domain.finance.md")
	readElapsed := time.Since(readStart)
	require.True(t, ok, "expected finance domain page in store")
	require.Less(t, readElapsed, 200*time.Millisecond,
		"page read blocked during regeneration for %s; store must serve stale content while generating", readElapsed)

	require.NoError(t, <-regenDone)
}
