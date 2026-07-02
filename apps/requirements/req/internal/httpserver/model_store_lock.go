package httpserver

import (
	"context"
	"sync"
	"time"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/perftrack"
)

// storeState holds published model content. Every field is accessed only through
// the locking helpers in this file so HTTP readers and reload writers share one
// clear contract: generate outside the lock, publish snapshots under a brief write
// lock, serve reads under a read lock.
type storeState struct {
	mu          sync.RWMutex
	models      map[string]*core.Model
	markdown    map[string]map[string][]byte
	css         map[string][]byte
	svg         map[string]map[string][]byte
	modelErrors map[string]string
	parseIssues map[string]*generate.ParseIssueIndex
}

func newStoreState() storeState {
	return storeState{
		models:      make(map[string]*core.Model),
		markdown:    make(map[string]map[string][]byte),
		css:         make(map[string][]byte),
		svg:         make(map[string]map[string][]byte),
		modelErrors: make(map[string]string),
		parseIssues: make(map[string]*generate.ParseIssueIndex),
	}
}

// publishedSnapshot is generated off-lock and swapped in atomically under write lock.
type publishedSnapshot struct {
	model       *core.Model
	markdown    map[string][]byte
	svg         map[string][]byte
	css         []byte
	parseIssues *generate.ParseIssueIndex
}

// withReadLock runs fn while holding the store read lock.
func (st *storeState) withReadLock(ctx context.Context, fn func()) {
	lockStart := time.Now()
	st.mu.RLock()
	defer st.mu.RUnlock()
	recordLockWait(ctx, time.Since(lockStart))
	fn()
}

// withWriteLock runs fn while holding the store write lock. When tracker is set,
// time spent waiting for the lock is recorded as store.lock_wait.
func (st *storeState) withWriteLock(tracker *perftrack.Tracker, fn func()) {
	lockStart := time.Now()
	st.mu.Lock()
	defer st.mu.Unlock()
	if tracker != nil {
		tracker.Add("store.lock_wait", time.Since(lockStart))
	}
	fn()
}

// publish swaps a freshly generated snapshot into the store. Call only after
// generation completes; the write lock is held only for map assignments.
func (st *storeState) publish(name string, snapshot publishedSnapshot, tracker *perftrack.Tracker) {
	st.withWriteLock(tracker, func() {
		st.models[name] = snapshot.model
		st.markdown[name] = snapshot.markdown
		st.svg[name] = snapshot.svg
		st.css[name] = snapshot.css
		st.parseIssues[name] = snapshot.parseIssues
		delete(st.modelErrors, name)
	})
}

// setModelError records a reload failure while leaving the previous snapshot served.
func (st *storeState) setModelError(name string, err error) {
	st.withWriteLock(nil, func() {
		msg := "unknown error"
		if err != nil {
			msg = err.Error()
		}
		st.modelErrors[name] = msg
	})
}

func (st *storeState) modelError(name string) (string, bool) {
	var msg string
	var ok bool
	st.withReadLock(context.Background(), func() {
		msg, ok = st.modelErrors[name]
	})
	return msg, ok
}

func (st *storeState) parseIssuesFor(name string) (*generate.ParseIssueIndex, bool) {
	var idx *generate.ParseIssueIndex
	var ok bool
	st.withReadLock(context.Background(), func() {
		idx, ok = st.parseIssues[name]
	})
	return idx, ok
}

func (st *storeState) model(name string) (*core.Model, bool) {
	var m *core.Model
	var ok bool
	st.withReadLock(context.Background(), func() {
		m, ok = st.models[name]
	})
	return m, ok
}

func (st *storeState) markdownFile(ctx context.Context, model, file string) ([]byte, bool) {
	var content []byte
	var ok bool
	st.withReadLock(ctx, func() {
		if files, found := st.markdown[model]; found {
			content, ok = files[file]
		}
	})
	return content, ok
}

func (st *storeState) cssFor(ctx context.Context, model string) ([]byte, bool) {
	var content []byte
	var ok bool
	st.withReadLock(ctx, func() {
		content, ok = st.css[model]
	})
	return content, ok
}

func (st *storeState) svgFile(ctx context.Context, model, file string) ([]byte, bool) {
	var content []byte
	var ok bool
	st.withReadLock(ctx, func() {
		if files, found := st.svg[model]; found {
			content, ok = files[file]
		}
	})
	return content, ok
}

func (st *storeState) modelNames(ctx context.Context) []string {
	var names []string
	st.withReadLock(ctx, func() {
		names = make([]string, 0, len(st.models))
		for name := range st.models {
			names = append(names, name)
		}
	})
	return names
}

func recordLockWait(ctx context.Context, wait time.Duration) {
	if tracker := perftrack.FromContext(ctx); tracker != nil {
		tracker.Add("store.lock_wait", wait)
	}
}
